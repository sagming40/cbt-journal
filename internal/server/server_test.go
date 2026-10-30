package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	// 테스트에서는 로그를 버린다. slog.DiscardHandler = 아무것도 안 하는 로그 처리기.
	h := New(slog.New(slog.DiscardHandler))

	t.Run("GET은 200과 status ok", func(t *testing.T) {
		// httptest: 진짜 포트를 열지 않고 요청 · 응답을 메모리에서 주고받는다.
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("상태 코드 = %d, 200이어야 한다", rec.Code)
		}
		if got := rec.Body.String(); got != "{\"status\":\"ok\"}\n" {
			t.Errorf("본문 = %q", got)
		}
	})

	t.Run("POST는 405", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/health", nil))

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("상태 코드 = %d, 405여야 한다", rec.Code)
		}
	})
}
