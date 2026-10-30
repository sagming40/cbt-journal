package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// spa는 프론트 빌드 결과(SvelteKit SPA)를 내준다 (ARCHITECTURE 1절).
//   - 있는 파일(/_app/..., /robots.txt)은 그 파일
//   - 없는 경로(/checkin 등)는 index.html → 어느 화면인지는 SvelteKit이 브라우저에서 고른다
//   - /api/로 시작하는데 여기까지 온 것은 없는 API → JSON 404 (HTML을 주면 프론트가 JSON을 기대하다 깨진다)
//
// 비유: 안내 데스크에 없는 방 번호를 물으면 "로비(index.html)로 가세요", 단 직원 전용 층(/api)은 "그런 방 없어요".
func spa(web fs.FS) http.Handler {
	files := http.FileServerFS(web)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "없는 API 경로예요", nil)
			return
		}

		// path.Clean으로 /../ 같은 걸 정리한 뒤 찾는다. embed된 파일만 보므로 디스크의 다른 파일에는 애초에 닿지 않는다.
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		info, err := fs.Stat(web, name)
		if name == "" || err != nil || info.IsDir() {
			// 첫 화면, 화면 경로, 폴더 주소는 전부 index.html. 폴더 목록은 보여 주지 않는다.
			serveIndex(w, r, web)
			return
		}

		// SvelteKit은 _app/immutable/ 아래 파일 이름에 내용 해시를 붙인다.
		// 내용이 바뀌면 이름도 바뀌니 오래 캐시해도 옛 파일이 남을 일이 없다.
		if strings.HasPrefix(name, "_app/immutable/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		files.ServeHTTP(w, r)
	})
}

// serveIndex는 index.html을 준다. 새 버전을 올리면 바로 받도록 캐시하지 않게 한다.
func serveIndex(w http.ResponseWriter, r *http.Request, web fs.FS) {
	if _, err := fs.Stat(web, "index.html"); err != nil {
		// 프론트를 빌드 · 복사하지 않고 Go만 빌드한 경우 (README "실행" 참고)
		http.Error(w, "프론트 빌드가 없다: web에서 pnpm build 뒤 internal/webdist/dist로 복사하고 다시 빌드한다", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFileFS(w, r, web, "index.html")
}
