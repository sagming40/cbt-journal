package server

// 인증 통합 테스트: 진짜 Postgres 18 · Valkey 9 컨테이너(testutil) + 진짜 HTTPS 서버 + 쿠키 저장소.
// PLAN M2 완료 기준과 사전 검증에서 더한 시나리오를 그대로 옮겼다.
// Docker가 없으면 로컬에서는 건너뛰고, CI에서는 실패한다.

import (
	"bytes"
	"context"
	"crypto/sha1" //nolint:gosec // 가짜 Pwned Passwords 응답을 만들려고 쓴다
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/sagming40/cbt-journal/internal/auth"
	"github.com/sagming40/cbt-journal/internal/testutil"
)

const (
	testEmail    = "me@example.com"
	testPassword = "처음 비밀번호는 충분히 길다"
	newPassword  = "새 비밀번호도 충분히 길다"
)

type testApp struct {
	t    *testing.T
	srv  *httptest.Server
	pool *pgxpool.Pool
	rdb  *redis.Client
	svc  *auth.Service
}

func newApp(t *testing.T, configure ...func(*auth.Options)) *testApp {
	t.Helper()
	env := testutil.Start(t)
	pool, rdb := env.Pool(t), env.Valkey(t)
	testutil.Reset(t, pool, rdb)

	discard := slog.New(slog.DiscardHandler)
	opts := auth.Options{Pool: pool, Valkey: rdb, Logger: discard}
	for _, c := range configure {
		c(&opts)
	}
	svc := auth.NewService(opts)
	h := New(Deps{Logger: discard, Checks: Checks{Postgres: okCheck, Valkey: okCheck}, Web: fstest.MapFS{}, Auth: svc})
	// Secure 쿠키는 HTTPS에서만 오가므로 TLS 서버로 띄운다.
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	return &testApp{t: t, srv: srv, pool: pool, rdb: rdb, svc: svc}
}

// device는 쿠키 저장소를 따로 가진 클라이언트 하나 — 기기 한 대처럼 쓴다.
func (a *testApp) device() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Transport: a.srv.Client().Transport, Jar: jar}
}

func (a *testApp) createUser(email, password string) {
	a.t.Helper()
	if _, err := a.svc.CreateUser(context.Background(), email, password, "테스터"); err != nil {
		a.t.Fatalf("사용자 만들기: %v", err)
	}
}

type response struct {
	status int
	header http.Header
	body   []byte
}

func (r response) errorCode() string {
	var e struct {
		Error struct {
			Code   string            `json:"code"`
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	_ = json.Unmarshal(r.body, &e)
	return e.Error.Code
}

func (r response) fields() map[string]string {
	var e struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	_ = json.Unmarshal(r.body, &e)
	return e.Error.Fields
}

func (a *testApp) call(c *http.Client, method, path string, body any, headers ...string) response {
	a.t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(context.Background(), method, a.srv.URL+path, rd)
	if err != nil {
		a.t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	resp, err := c.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return response{status: resp.StatusCode, header: resp.Header, body: b}
}

func (a *testApp) login(c *http.Client, email, password string, remember bool) response {
	a.t.Helper()
	return a.call(c, "POST", "/api/auth/login", map[string]any{"email": email, "password": password, "remember": remember})
}

func sessionCookie(t *testing.T, r response) *http.Cookie {
	t.Helper()
	for _, c := range (&http.Response{Header: r.header}).Cookies() {
		if c.Name == auth.CookieName {
			return c
		}
	}
	t.Fatalf("세션 쿠키가 없다 (상태 %d, 본문 %s)", r.status, r.body)
	return nil
}

// ── 로그인 ──────────────────────────────────────────────

func TestLoginCookieAttributes(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)

	t.Run("로그인 유지 안 함: 세션 쿠키(Max-Age 없음)", func(t *testing.T) {
		c := a.device()
		r := a.login(c, testEmail, testPassword, false)
		if r.status != http.StatusOK {
			t.Fatalf("상태 %d: %s", r.status, r.body)
		}
		raw := r.header.Get("Set-Cookie")
		for _, want := range []string{"HttpOnly", "Secure", "SameSite=Lax", "Path=/"} {
			if !strings.Contains(raw, want) {
				t.Errorf("Set-Cookie에 %s가 없다: %s", want, raw)
			}
		}
		if strings.Contains(raw, "Max-Age") {
			t.Errorf("로그인 유지가 아니면 Max-Age가 없어야 한다(브라우저를 닫으면 사라지게): %s", raw)
		}
		if me := a.call(c, "GET", "/api/me", nil); me.status != http.StatusOK || !strings.Contains(string(me.body), testEmail) {
			t.Errorf("로그인 뒤 /api/me: %d %s", me.status, me.body)
		}
	})

	t.Run("로그인 유지: Max-Age 30일", func(t *testing.T) {
		r := a.login(a.device(), "  ME@example.com ", testPassword, true) // 이메일은 정규화해서 찾는다
		c := sessionCookie(t, r)
		if c.MaxAge < 29*24*3600 || c.MaxAge > 30*24*3600 {
			t.Errorf("Max-Age = %d, 30일이어야 한다", c.MaxAge)
		}
	})
}

func TestLoginFailuresLookTheSame(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)

	wrong := a.login(a.device(), testEmail, "틀린 비밀번호입니다아아", false)
	unknown := a.login(a.device(), "nobody@example.com", "틀린 비밀번호입니다아아", false)
	if wrong.status != http.StatusUnauthorized || unknown.status != http.StatusUnauthorized {
		t.Fatalf("둘 다 401이어야 한다: %d / %d", wrong.status, unknown.status)
	}
	if !bytes.Equal(wrong.body, unknown.body) {
		t.Errorf("가입 여부가 드러나면 안 된다:\n틀린 비번 %s\n없는 계정 %s", wrong.body, unknown.body)
	}
	if wrong.errorCode() != "INVALID_CREDENTIALS" {
		t.Errorf("코드 = %s", wrong.errorCode())
	}
}

func TestLoginRateLimits(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)

	t.Run("계정: 실패 5번 뒤에는 맞는 비밀번호도 429", func(t *testing.T) {
		c := a.device()
		for i := range 5 {
			if r := a.login(c, testEmail, "틀린 비밀번호입니다아아", false); r.status != http.StatusUnauthorized {
				t.Fatalf("%d번째 실패: 상태 %d", i+1, r.status)
			}
		}
		r := a.login(c, testEmail, testPassword, false)
		if r.status != http.StatusTooManyRequests {
			t.Fatalf("상태 %d, 429여야 한다", r.status)
		}
		if secs, _ := strconv.Atoi(r.header.Get("Retry-After")); secs <= 0 || secs > 15*60 {
			t.Errorf("Retry-After = %q", r.header.Get("Retry-After"))
		}
	})

	t.Run("성공하면 계정 카운터가 초기화된다", func(t *testing.T) {
		testutil.Reset(t, a.pool, a.rdb)
		a.createUser(testEmail, testPassword)
		c := a.device()
		for range 4 {
			a.login(c, testEmail, "틀린 비밀번호입니다아아", false)
		}
		if r := a.login(c, testEmail, testPassword, false); r.status != http.StatusOK {
			t.Fatalf("4번 틀린 뒤 성공: 상태 %d", r.status)
		}
		for range 4 {
			a.login(c, testEmail, "틀린 비밀번호입니다아아", false)
		}
		if r := a.login(c, testEmail, testPassword, false); r.status != http.StatusOK {
			t.Errorf("초기화됐다면 다시 4번은 괜찮아야 한다: 상태 %d", r.status)
		}
	})

	t.Run("IP: 15분에 20번까지", func(t *testing.T) {
		testutil.Reset(t, a.pool, a.rdb)
		c := a.device()
		for i := range 20 {
			// 계정 제한에 먼저 걸리지 않게 이메일을 바꿔 가며
			if r := a.login(c, fmt.Sprintf("u%d@example.com", i), "틀린 비밀번호입니다아아", false); r.status != http.StatusUnauthorized {
				t.Fatalf("%d번째: 상태 %d", i+1, r.status)
			}
		}
		if r := a.login(c, "u99@example.com", "틀린 비밀번호입니다아아", false); r.status != http.StatusTooManyRequests {
			t.Errorf("21번째: 상태 %d, 429여야 한다", r.status)
		}
	})
}

// ── 세션 ────────────────────────────────────────────────

func TestLogoutKillsSessionImmediately(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	c := a.device()
	cookie := sessionCookie(t, a.login(c, testEmail, testPassword, false))
	a.call(c, "GET", "/api/me", nil) // Valkey 캐시를 데워 둔다 — 캐시까지 지우는지 보려고

	if r := a.call(c, "POST", "/api/auth/logout", nil); r.status != http.StatusNoContent {
		t.Fatalf("로그아웃: %d %s", r.status, r.body)
	}
	// 지운 쿠키를 다시 들고 와도(훔친 쿠키라고 치고) 401이어야 한다.
	other := a.device()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", a.srv.URL+"/api/me", nil)
	req.AddCookie(cookie)
	resp, err := other.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("로그아웃한 쿠키로 /api/me: %d, 401이어야 한다", resp.StatusCode)
	}
}

func TestDBStoresOnlyTokenHash(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	cookie := sessionCookie(t, a.login(a.device(), testEmail, testPassword, false))

	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	var stored []byte
	if err := a.pool.QueryRow(context.Background(), "SELECT token_hash FROM sessions").Scan(&stored); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	if !bytes.Equal(stored, sum[:]) {
		t.Error("DB에는 토큰의 SHA-256 해시가 있어야 한다")
	}
	if bytes.Contains(stored, raw) || strings.Contains(hex.EncodeToString(stored), cookie.Value) {
		t.Error("DB에 토큰 원문이 있다")
	}
}

func TestExpiredSessionIs401AndCookieCleared(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	c := a.device()
	a.login(c, testEmail, testPassword, false)

	ctx := context.Background()
	if _, err := a.pool.Exec(ctx, "UPDATE sessions SET expires_at = now() - interval '1 second'"); err != nil {
		t.Fatal(err)
	}
	a.rdb.FlushDB(ctx)
	r := a.call(c, "GET", "/api/me", nil)
	if r.status != http.StatusUnauthorized {
		t.Fatalf("만료된 세션: %d, 401이어야 한다", r.status)
	}
	if !strings.Contains(r.header.Get("Set-Cookie"), "Max-Age=0") {
		t.Errorf("만료된 쿠키는 지우게 해야 한다: %q", r.header.Get("Set-Cookie"))
	}
}

func TestSessionRefresh(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	ctx := context.Background()

	t.Run("로그인 유지: 기한을 늘리면 쿠키도 다시 30일", func(t *testing.T) {
		c := a.device()
		a.login(c, testEmail, testPassword, true)
		// 이틀 전에 마지막으로 썼다고 치고, 캐시를 비워 DB를 읽게 한다.
		if _, err := a.pool.Exec(ctx, "UPDATE sessions SET last_seen_at = now() - interval '2 days', expires_at = now() + interval '28 days'"); err != nil {
			t.Fatal(err)
		}
		a.rdb.FlushDB(ctx)
		r := a.call(c, "GET", "/api/me", nil)
		ck := sessionCookie(t, r)
		if ck.MaxAge < 29*24*3600 {
			t.Errorf("다시 내려준 쿠키 Max-Age = %d, 30일이어야 한다", ck.MaxAge)
		}
		var left time.Duration
		if err := a.pool.QueryRow(ctx, "SELECT expires_at - now() FROM sessions WHERE remember").Scan(&left); err != nil {
			t.Fatal(err)
		}
		if left < 29*24*time.Hour {
			t.Errorf("DB 기한이 안 늘었다: %v", left)
		}
	})

	t.Run("유지 안 함: DB 기한만 늘고 쿠키는 그대로(세션 쿠키)", func(t *testing.T) {
		testutil.Reset(t, a.pool, a.rdb)
		a.createUser(testEmail, testPassword)
		c := a.device()
		a.login(c, testEmail, testPassword, false)
		if _, err := a.pool.Exec(ctx, "UPDATE sessions SET last_seen_at = now() - interval '1 hour', expires_at = now() + interval '11 hours'"); err != nil {
			t.Fatal(err)
		}
		a.rdb.FlushDB(ctx)
		r := a.call(c, "GET", "/api/me", nil)
		if r.status != http.StatusOK {
			t.Fatalf("상태 %d", r.status)
		}
		if r.header.Get("Set-Cookie") != "" {
			t.Errorf("세션 쿠키는 다시 내려줄 필요가 없다: %q", r.header.Get("Set-Cookie"))
		}
		var left time.Duration
		_ = a.pool.QueryRow(ctx, "SELECT expires_at - now() FROM sessions").Scan(&left)
		if left < 11*time.Hour+50*time.Minute {
			t.Errorf("DB 기한이 12시간으로 안 늘었다: %v", left)
		}
	})
}

// ── 비밀번호 변경 · 다른 기기 로그아웃 ───────────────────

func TestChangePasswordRevokesOtherSessions(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	phone, pc := a.device(), a.device()
	a.login(phone, testEmail, testPassword, true)
	a.login(pc, testEmail, testPassword, false)
	a.call(pc, "GET", "/api/me", nil) // 캐시를 데워 둔다

	r := a.call(phone, "PUT", "/api/me/password", map[string]string{"current_password": testPassword, "new_password": newPassword})
	if r.status != http.StatusNoContent {
		t.Fatalf("비밀번호 변경: %d %s", r.status, r.body)
	}
	if r := a.call(pc, "GET", "/api/me", nil); r.status != http.StatusUnauthorized {
		t.Errorf("다른 기기 세션이 살아 있다(캐시 포함 즉시 끊겨야 한다): %d", r.status)
	}
	if r := a.call(phone, "GET", "/api/me", nil); r.status != http.StatusOK {
		t.Errorf("바꾼 기기의 세션은 남아야 한다: %d", r.status)
	}
	if r := a.login(a.device(), testEmail, testPassword, false); r.status != http.StatusUnauthorized {
		t.Errorf("옛 비밀번호로 로그인되면 안 된다: %d", r.status)
	}
	if r := a.login(a.device(), testEmail, newPassword, false); r.status != http.StatusOK {
		t.Errorf("새 비밀번호로 로그인: %d", r.status)
	}
}

func TestChangePasswordValidationAndLimit(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	c := a.device()
	a.login(c, testEmail, testPassword, false)

	tests := []struct {
		name, current, next, field string
	}{
		{"새 비밀번호가 짧다", testPassword, "짧다", "new_password"},
		{"지금 비밀번호와 같다", testPassword, testPassword, "new_password"},
	}
	for _, tt := range tests {
		r := a.call(c, "PUT", "/api/me/password", map[string]string{"current_password": tt.current, "new_password": tt.next})
		if r.status != http.StatusBadRequest || r.fields()[tt.field] == "" {
			t.Errorf("%s: %d %s", tt.name, r.status, r.body)
		}
	}

	for i := range 5 {
		r := a.call(c, "PUT", "/api/me/password", map[string]string{"current_password": "틀린 지금 비밀번호입니다", "new_password": newPassword})
		if r.status != http.StatusBadRequest || r.fields()["current_password"] == "" {
			t.Fatalf("%d번째 틀림: %d %s", i+1, r.status, r.body)
		}
	}
	r := a.call(c, "PUT", "/api/me/password", map[string]string{"current_password": testPassword, "new_password": newPassword})
	if r.status != http.StatusTooManyRequests {
		t.Errorf("현재 비밀번호 5번 틀린 뒤: %d, 429여야 한다", r.status)
	}
}

func TestLogoutOtherSessions(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	phone, pc, tablet := a.device(), a.device(), a.device()
	for _, d := range []*http.Client{phone, pc, tablet} {
		a.login(d, testEmail, testPassword, false)
		a.call(d, "GET", "/api/me", nil)
	}

	r := a.call(pc, "POST", "/api/me/sessions/logout-others", nil)
	if r.status != http.StatusOK || !strings.Contains(string(r.body), `"revoked":2`) {
		t.Fatalf("다른 기기 로그아웃: %d %s", r.status, r.body)
	}
	for name, d := range map[string]*http.Client{"휴대폰": phone, "태블릿": tablet} {
		if r := a.call(d, "GET", "/api/me", nil); r.status != http.StatusUnauthorized {
			t.Errorf("%s 세션이 살아 있다: %d", name, r.status)
		}
	}
	if r := a.call(pc, "GET", "/api/me", nil); r.status != http.StatusOK {
		t.Errorf("지금 기기는 남아야 한다: %d", r.status)
	}
}

// ── 가입 · 내 정보 ──────────────────────────────────────

func TestSignup(t *testing.T) {
	body := map[string]string{"email": "new@example.com", "password": testPassword, "nickname": "새 사람"}

	t.Run("꺼져 있으면 403", func(t *testing.T) {
		a := newApp(t)
		if r := a.call(a.device(), "POST", "/api/auth/signup", body); r.status != http.StatusForbidden || r.errorCode() != "SIGNUP_DISABLED" {
			t.Errorf("%d %s", r.status, r.body)
		}
		if r := a.call(a.device(), "GET", "/api/config", nil); !strings.Contains(string(r.body), `"allow_signup":false`) {
			t.Errorf("공개 설정: %s", r.body)
		}
	})

	t.Run("켜져 있으면 가입 + 로그인, 중복 409, 입력 오류 400", func(t *testing.T) {
		a := newApp(t, func(o *auth.Options) { o.AllowSignup = true })
		c := a.device()
		r := a.call(c, "POST", "/api/auth/signup", body)
		if r.status != http.StatusCreated {
			t.Fatalf("가입: %d %s", r.status, r.body)
		}
		if me := a.call(c, "GET", "/api/me", nil); me.status != http.StatusOK {
			t.Errorf("가입 뒤 로그인돼 있어야 한다: %d", me.status)
		}
		if r := a.call(a.device(), "POST", "/api/auth/signup", body); r.status != http.StatusConflict {
			t.Errorf("같은 이메일 다시: %d", r.status)
		}
		bad := a.call(a.device(), "POST", "/api/auth/signup", map[string]string{"email": "not-email", "password": "짧다", "nickname": ""})
		f := bad.fields()
		if bad.status != http.StatusBadRequest || f["email"] == "" || f["password"] == "" || f["nickname"] == "" {
			t.Errorf("칸별 오류를 한 번에 알려 줘야 한다: %d %s", bad.status, bad.body)
		}
	})
}

func TestUpdateMe(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	c := a.device()
	a.login(c, testEmail, testPassword, false)

	if r := a.call(c, "PATCH", "/api/me", map[string]string{"timezone": "Mars/Olympus"}); r.status != http.StatusBadRequest || r.fields()["timezone"] == "" {
		t.Errorf("잘못된 시간대: %d %s", r.status, r.body)
	}
	r := a.call(c, "PATCH", "/api/me", map[string]string{"nickname": " 새 이름 ", "timezone": "America/New_York"})
	if r.status != http.StatusOK || !strings.Contains(string(r.body), `"nickname":"새 이름"`) || !strings.Contains(string(r.body), "America/New_York") {
		t.Errorf("바꾸기: %d %s", r.status, r.body)
	}
	// 보낸 칸만 바뀐다
	r = a.call(c, "PATCH", "/api/me", map[string]string{"nickname": "또 바꿈"})
	if !strings.Contains(string(r.body), "America/New_York") {
		t.Errorf("보내지 않은 시간대가 바뀌었다: %s", r.body)
	}
}

// ── 잠금 · 보안 ─────────────────────────────────────────

func TestDefaultLock(t *testing.T) {
	a := newApp(t)
	c := a.device()
	// openapi.yaml에서 security: []로 연 경로 말고는 전부 401이어야 한다 (ARCHITECTURE 3.1절 "기본 잠금").
	locked := []struct{ method, path string }{
		{"POST", "/api/auth/logout"},
		{"GET", "/api/me"},
		{"PATCH", "/api/me"},
		{"PUT", "/api/me/password"},
		{"POST", "/api/me/sessions/logout-others"},
	}
	for _, l := range locked {
		if r := a.call(c, l.method, l.path, map[string]string{}); r.status != http.StatusUnauthorized || r.errorCode() != "UNAUTHORIZED" {
			t.Errorf("%s %s: %d %s, 401이어야 한다", l.method, l.path, r.status, r.body)
		}
	}
	public := []struct{ method, path string }{
		{"GET", "/api/health"},
		{"GET", "/api/config"},
		{"POST", "/api/auth/login"},
		{"POST", "/api/auth/signup"},
	}
	for _, p := range public {
		if r := a.call(c, p.method, p.path, map[string]string{}); r.status == http.StatusUnauthorized && r.errorCode() == "UNAUTHORIZED" {
			t.Errorf("%s %s는 공개 경로인데 잠겨 있다", p.method, p.path)
		}
	}
}

func TestCrossOriginPostBlocked(t *testing.T) {
	a := newApp(t)
	a.createUser(testEmail, testPassword)
	r := a.call(a.device(), "POST", "/api/auth/login",
		map[string]any{"email": testEmail, "password": testPassword},
		"Sec-Fetch-Site", "cross-site")
	if r.status != http.StatusForbidden || r.errorCode() != "CROSS_ORIGIN" {
		t.Errorf("다른 사이트에서 온 POST: %d %s, 403이어야 한다", r.status, r.body)
	}
}

func TestBodyTooLarge(t *testing.T) {
	a := newApp(t)
	huge := map[string]string{"email": testEmail, "password": strings.Repeat("a", 2<<20)}
	if r := a.call(a.device(), "POST", "/api/auth/login", huge); r.status != http.StatusRequestEntityTooLarge {
		t.Errorf("2MiB 본문: %d, 413이어야 한다", r.status)
	}
}

// ── 유출 비밀번호 (A10) ─────────────────────────────────

func TestBreachedPasswordCheck(t *testing.T) {
	const leaked = "유출된 비밀번호라고 치자"
	sum := sha1.Sum([]byte(leaked)) //nolint:gosec // 위 import 주석과 같은 이유
	full := strings.ToUpper(hex.EncodeToString(sum[:]))

	var lastPath atomic.Value
	var mode atomic.Value // "ok" | "fail" | "slow"
	mode.Store("ok")
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath.Store(r.URL.Path)
		switch mode.Load() {
		case "fail":
			w.WriteHeader(http.StatusInternalServerError)
		case "slow":
			time.Sleep(500 * time.Millisecond)
		default:
			fmt.Fprintf(w, "%s:7\n", full[5:])
		}
	}))
	defer fake.Close()

	checker := &auth.PwnedPasswords{BaseURL: fake.URL + "/range/", Client: &http.Client{Timeout: 200 * time.Millisecond}}
	a := newApp(t, func(o *auth.Options) { o.AllowSignup = true; o.Breach = checker })
	signup := func(email string) response {
		return a.call(a.device(), "POST", "/api/auth/signup", map[string]string{"email": email, "password": leaked, "nickname": "나"})
	}

	r := signup("a@example.com")
	if r.status != http.StatusBadRequest || r.fields()["password"] == "" {
		t.Fatalf("유출 비밀번호 가입: %d %s, 400이어야 한다", r.status, r.body)
	}
	if p, _ := lastPath.Load().(string); p != "/range/"+full[:5] {
		t.Errorf("가짜 서버가 받은 경로 %q — 해시 앞 5글자만 보내야 한다", p)
	}

	mode.Store("fail")
	if r := signup("b@example.com"); r.status != http.StatusCreated {
		t.Errorf("확인 서버가 500이어도 가입은 돼야 한다: %d %s", r.status, r.body)
	}
	mode.Store("slow")
	if r := signup("c@example.com"); r.status != http.StatusCreated {
		t.Errorf("확인 서버가 타임아웃이어도 가입은 돼야 한다: %d %s", r.status, r.body)
	}
}
