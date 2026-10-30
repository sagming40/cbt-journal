package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 가짜 확인 함수들: 진짜 DB 없이 "살아 있음 / 죽음"을 흉내 낸다.
func okCheck(context.Context) error   { return nil }
func failCheck(context.Context) error { return errors.New("dial tcp 127.0.0.1:6380: 연결 거부") }

// do는 진짜 포트 없이 요청 하나를 보내고 응답을 돌려준다.
func do(h http.Handler, method string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, "/api/health", nil))
	return rec
}

func TestHealth(t *testing.T) {
	discard := slog.New(slog.DiscardHandler)

	t.Run("모두 살아 있으면 200", func(t *testing.T) {
		h := New(discard, map[string]Checker{"postgres": okCheck, "valkey": okCheck})
		rec := do(h, http.MethodGet)

		if rec.Code != http.StatusOK {
			t.Fatalf("상태 코드 = %d, 200이어야 한다", rec.Code)
		}
		want := `{"postgres":"ok","status":"ok","valkey":"ok"}` + "\n"
		if got := rec.Body.String(); got != want {
			t.Errorf("본문 = %q, 기대 = %q", got, want)
		}
	})

	t.Run("하나라도 죽으면 503이고 어느 쪽인지 알려 준다", func(t *testing.T) {
		h := New(discard, map[string]Checker{"postgres": okCheck, "valkey": failCheck})
		rec := do(h, http.MethodGet)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("상태 코드 = %d, 503이어야 한다", rec.Code)
		}
		want := `{"postgres":"ok","status":"error","valkey":"error"}` + "\n"
		if got := rec.Body.String(); got != want {
			t.Errorf("본문 = %q, 기대 = %q", got, want)
		}
		// 에러 원문(주소 등)이 응답으로 새면 안 된다.
		if strings.Contains(rec.Body.String(), "6380") {
			t.Errorf("응답에 내부 주소가 섞였다: %s", rec.Body.String())
		}
	})

	t.Run("POST는 405", func(t *testing.T) {
		h := New(discard, map[string]Checker{"postgres": okCheck})
		if rec := do(h, http.MethodPost); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("상태 코드 = %d, 405여야 한다", rec.Code)
		}
	})
}
