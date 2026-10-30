package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sagming40/cbt-journal/internal/apigen"
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

	tests := []struct {
		name     string
		checks   Checks
		wantCode int
		want     apigen.Health
	}{
		{
			name:     "모두 살아 있으면 200",
			checks:   Checks{Postgres: okCheck, Valkey: okCheck},
			wantCode: http.StatusOK,
			want:     apigen.Health{Status: "ok", Postgres: "ok", Valkey: "ok"},
		},
		{
			name:     "Valkey가 죽으면 503",
			checks:   Checks{Postgres: okCheck, Valkey: failCheck},
			wantCode: http.StatusServiceUnavailable,
			want:     apigen.Health{Status: "error", Postgres: "ok", Valkey: "error"},
		},
		{
			name:     "Postgres가 죽으면 503",
			checks:   Checks{Postgres: failCheck, Valkey: okCheck},
			wantCode: http.StatusServiceUnavailable,
			want:     apigen.Health{Status: "error", Postgres: "error", Valkey: "ok"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(New(discard, tt.checks, fstest.MapFS{}), http.MethodGet)

			if rec.Code != tt.wantCode {
				t.Fatalf("상태 코드 = %d, %d이어야 한다", rec.Code, tt.wantCode)
			}
			// 글자 그대로 비교하지 않고 계약서 타입으로 풀어서 비교한다: 필드 순서가 바뀌어도 깨지지 않게.
			var got apigen.Health
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("응답이 Health 모양이 아니다: %v (본문 %q)", err, rec.Body.String())
			}
			if got != tt.want {
				t.Errorf("본문 = %+v, 기대 = %+v", got, tt.want)
			}
			// 에러 원문(주소 등)이 응답으로 새면 안 된다.
			if strings.Contains(rec.Body.String(), "6380") {
				t.Errorf("응답에 내부 주소가 섞였다: %s", rec.Body.String())
			}
		})
	}

	t.Run("POST는 405", func(t *testing.T) {
		h := New(discard, Checks{Postgres: okCheck, Valkey: okCheck}, fstest.MapFS{})
		if rec := do(h, http.MethodPost); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("상태 코드 = %d, 405여야 한다", rec.Code)
		}
	})
}
