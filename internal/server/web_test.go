package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// fakeWeb은 SvelteKit 빌드 결과를 흉내 낸 가짜 파일 묶음이다.
var fakeWeb = fstest.MapFS{
	"index.html":              {Data: []byte("<!doctype html>앱")},
	"robots.txt":              {Data: []byte("User-agent: *")},
	"_app/immutable/entry.js": {Data: []byte("console.log(1)")},
}

func get(h http.Handler, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), method, path, nil))
	return rec
}

func TestSPA(t *testing.T) {
	h := New(slog.New(slog.DiscardHandler), Checks{Postgres: okCheck, Valkey: okCheck}, fakeWeb)

	tests := []struct {
		name     string
		path     string
		wantBody string
	}{
		{"첫 화면", "/", "<!doctype html>앱"},
		{"화면 경로는 index.html", "/checkin", "<!doctype html>앱"},
		{"깊은 화면 경로도 index.html", "/records/123", "<!doctype html>앱"},
		{"있는 파일은 그 파일", "/robots.txt", "User-agent: *"},
		{"빌드 결과 JS", "/_app/immutable/entry.js", "console.log(1)"},
		{"폴더 주소는 목록 대신 index.html", "/_app/", "<!doctype html>앱"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := get(h, http.MethodGet, tt.path)
			if rec.Code != http.StatusOK {
				t.Fatalf("상태 코드 = %d, 200이어야 한다", rec.Code)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("본문 = %q, 기대 = %q", got, tt.wantBody)
			}
		})
	}

	t.Run("위로 올라가는 경로는 mux가 정리된 주소로 돌려보낸다", func(t *testing.T) {
		// Go의 ServeMux는 ..이 섞인 경로를 spa까지 보내지 않고 정리된 주소로 리다이렉트한다.
		// 따라가도 embed된 파일 안의 /go.mod(없음) → index.html이라 디스크의 파일에는 닿지 않는다.
		rec := get(h, http.MethodGet, "/../../go.mod")
		if rec.Code != http.StatusTemporaryRedirect {
			t.Fatalf("상태 코드 = %d, 307이어야 한다", rec.Code)
		}
		if loc := rec.Header().Get("Location"); loc != "/go.mod" {
			t.Errorf("Location = %q, /go.mod여야 한다", loc)
		}
	})

	t.Run("없는 API는 JSON 404 봉투", func(t *testing.T) {
		rec := get(h, http.MethodGet, "/api/nope")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("상태 코드 = %d, 404여야 한다", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("Content-Type = %q, JSON이어야 한다", ct)
		}
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error.Code != "NOT_FOUND" {
			t.Errorf("오류 봉투가 아니다: %q (%v)", rec.Body.String(), err)
		}
	})

	t.Run("API 경로에 다른 메서드는 그대로 405", func(t *testing.T) {
		if rec := get(h, http.MethodPost, "/api/health"); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("상태 코드 = %d, 405여야 한다", rec.Code)
		}
	})

	t.Run("immutable 파일은 오래 캐시, index.html은 캐시 안 함", func(t *testing.T) {
		if cc := get(h, http.MethodGet, "/_app/immutable/entry.js").Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
			t.Errorf("JS Cache-Control = %q", cc)
		}
		if cc := get(h, http.MethodGet, "/checkin").Header().Get("Cache-Control"); cc != "no-cache" {
			t.Errorf("index.html Cache-Control = %q", cc)
		}
	})

	t.Run("프론트 빌드가 없으면 503 안내", func(t *testing.T) {
		empty := New(slog.New(slog.DiscardHandler), Checks{Postgres: okCheck, Valkey: okCheck}, fstest.MapFS{".gitkeep": {}})
		if rec := get(empty, http.MethodGet, "/"); rec.Code != http.StatusServiceUnavailable {
			t.Errorf("상태 코드 = %d, 503이어야 한다", rec.Code)
		}
	})
}
