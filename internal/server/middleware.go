package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/sagming40/cbt-journal/internal/apigen"
	"github.com/sagming40/cbt-journal/internal/auth"
)

// maxBodyBytes: API 요청 본문 상한 (REQUIREMENTS 5.2절). 비밀번호 해시는 일부러 무거운 계산이라 큰 본문으로 괴롭히지 못하게.
const maxBodyBytes = 1 << 20

// logInfo는 접근 로그 한 줄에 들어갈 값을 안쪽 미들웨어가 채워 넣는 바구니다.
// 접근 로그는 가장 바깥이라 안쪽에서 정한 사용자 id를 직접 볼 수 없어서, 바구니를 context로 내려보낸다.
type logInfo struct {
	userID string
}

type logInfoKey struct{}

// statusRecorder는 handler가 쓴 상태 코드를 기억한다.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// 1. accessLog: 요청 id + 접근 로그. 경로 · 상태 코드 · 걸린 시간 · 사용자 id만 남긴다.
// 요청 본문 · 쿼리 문자열은 남기지 않는다 — 기록 내용이 로그로 새지 않게 (REQUIREMENTS 5.1절).
func accessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			id := newRequestID()
			w.Header().Set("X-Request-ID", id)
			info := &logInfo{}
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r.WithContext(context.WithValue(r.Context(), logInfoKey{}, info)))

			status := rec.status
			if status == 0 {
				status = http.StatusOK
			}
			logger.Info("요청",
				"id", id,
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"took_ms", time.Since(start).Milliseconds(),
				"user_id", info.userID,
			)
		})
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// 2. recoverPanic: handler가 패닉을 내도 서버 전체가 죽지 않고 그 요청만 500 봉투로 끝낸다.
func recoverPanic(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler { //nolint:errorlint // 표준 라이브러리가 쓰는 신호값 그대로 비교한다
					panic(v) // net/http가 연결을 조용히 끊으라고 쓰는 신호는 그대로 넘긴다
				}
				logger.Error("패닉", "path", r.URL.Path, "panic", v, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, "INTERNAL", "잠시 문제가 생겼어요. 다시 시도해 주세요", nil)
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// 3. limitBody: 본문을 1MiB까지만 읽게 한다. 넘으면 읽는 쪽(JSON 해석)이 오류를 받는다 → requestError가 413으로.
func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		next.ServeHTTP(w, r)
	})
}

// 4. crossOrigin: 다른 사이트에서 몰래 보낸 쓰기 요청(CSRF)을 막는다 (Go 1.25 표준).
// 브라우저가 붙이는 Sec-Fetch-Site · Origin 헤더로 판단하고, GET · HEAD처럼 읽기만 하는 요청은 통과시킨다.
func crossOrigin(next http.Handler) http.Handler {
	cop := http.NewCrossOriginProtection()
	cop.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusForbidden, "CROSS_ORIGIN", "다른 사이트에서 온 요청은 받을 수 없어요", nil)
	}))
	return cop.Handler(next)
}

// 6. sessionMiddleware: 쿠키로 세션을 찾아 context에 넣는다. 로그인이 필요한지는 여기서 정하지 않는다(7번 몫).
func sessionMiddleware(sessions *auth.Sessions, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(auth.CookieName)
			if sessions == nil || err != nil || c.Value == "" {
				next.ServeHTTP(w, r)
				return
			}
			sess, refreshed, err := sessions.Lookup(r.Context(), c.Value)
			if err != nil {
				// DB 장애: 로그인 안 한 것으로 치지 않고 500으로 끝낸다(멀쩡한 사용자를 로그인 화면으로 튕기지 않게).
				logger.Error("세션 조회 실패", "err", err)
				writeError(w, http.StatusInternalServerError, "INTERNAL", "잠시 문제가 생겼어요. 다시 시도해 주세요", nil)
				return
			}
			if sess == nil {
				// 형식이 틀렸거나 지워진 · 만료된 세션의 쿠키는 지우게 한다 (ARCHITECTURE 4절).
				http.SetCookie(w, auth.ClearCookie())
				next.ServeHTTP(w, r)
				return
			}
			if refreshed {
				// 로그인 유지 세션의 기한을 늘렸으면 쿠키 Max-Age도 다시 30일로 (ARCHITECTURE 4절).
				http.SetCookie(w, auth.Cookie(c.Value, sess.ExpiresAt, true))
			}
			if info, ok := r.Context().Value(logInfoKey{}).(*logInfo); ok {
				info.userID = sess.UserID.String()
			}
			next.ServeHTTP(w, r.WithContext(auth.WithSession(r.Context(), sess)))
		})
	}
}

// 7. requireAuth: 생성 코드가 "인증 필요"로 표시한 경로인데 세션이 없으면 401 (ARCHITECTURE 3.1절 "기본 잠금").
// 표시는 openapi.yaml의 security에서 온다 — 잠글지 열지는 계약서가 정한다.
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, locked := r.Context().Value(apigen.CookieAuthScopes).([]string); locked && auth.CurrentSession(r.Context()) == nil {
			body := auth.UnauthorizedBody()
			writeError(w, http.StatusUnauthorized, body.Error.Code, body.Error.Message, nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requestError: 요청을 해석하지 못함(잘못된 JSON 등). 본문이 너무 크면 413.
func requestError(logger *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "요청이 너무 커요", nil)
			return
		}
		logger.Debug("요청 해석 실패", "path", r.URL.Path, "err", err)
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "요청 형식이 맞지 않아요", nil)
	}
}

// internalError: handler가 돌려준 예상 못 한 오류 → 500 봉투. 원인은 로그에만.
func internalError(logger *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("처리 중 오류", "path", r.URL.Path, "err", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL", "잠시 문제가 생겼어요. 다시 시도해 주세요", nil)
	}
}
