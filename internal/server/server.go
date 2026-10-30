// Package server는 HTTP 라우터를 조립한다 (ARCHITECTURE 2절).
// 미들웨어 · 오류 봉투(3.1 · 3.2절)는 M2에서 붙인다.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// Checker는 연결 하나(Postgres, Valkey …)가 살아 있는지 확인한다. 살아 있으면 nil.
// 함수 타입으로 둬서 테스트에서는 가짜를 쉽게 꽂는다 (콘센트 규격만 맞으면 된다).
type Checker func(ctx context.Context) error

// New는 모든 경로를 등록한 핸들러 하나를 돌려준다.
// checks: 이름 → 확인 함수. health가 이것들을 차례로 불러 본다.
func New(logger *slog.Logger, checks map[string]Checker) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth(logger, checks))
	return mux
}

// handleHealth는 각 연결을 확인해 {"postgres":"ok","status":"ok","valkey":"ok"}처럼 답한다.
// 하나라도 실패하면 503. 로그인 없이 열리는 경로라 응답에는 ok / error만 싣는다 (ARCHITECTURE 3.1절).
func handleHealth(logger *slog.Logger, checks map[string]Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 한쪽이 멈춰도 health 요청 자체가 매달리지 않게 2초로 자른다.
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		code := http.StatusOK
		body := map[string]string{"status": "ok"}
		for name, check := range checks {
			if err := check(ctx); err != nil {
				// 원인(주소 · 사용자 이름이 섞일 수 있음)은 로그에만 남긴다.
				logger.Error("health 확인 실패", "target", name, "err", err)
				body[name] = "error"
				body["status"] = "error"
				code = http.StatusServiceUnavailable
				continue
			}
			body[name] = "ok"
		}

		w.Header().Set("Content-Type", "application/json")
		// WriteHeader는 본문을 쓰기 전에 불러야 한다. 본문을 먼저 쓰면 200으로 확정된다.
		w.WriteHeader(code)
		if err := json.NewEncoder(w).Encode(body); err != nil {
			logger.Error("health 응답 쓰기 실패", "err", err)
		}
	}
}
