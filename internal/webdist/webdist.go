// Package webdist는 프론트 빌드 결과를 실행 파일 안에 넣는다 (ARCHITECTURE 7절).
// web/build를 dist/로 복사한 뒤 go build 하면 그 내용이 들어간다. go:embed는 자기 폴더 아래만 넣을 수 있어서 복사한다.
package webdist

import (
	"embed"
	"io/fs"
)

// all: 접두사가 있어야 _app/(밑줄로 시작)과 .gitkeep도 들어간다.
// go:embed는 기본으로 . _ 로 시작하는 파일을 조용히 빼는데, SvelteKit 결과물의 JS · CSS가 전부 _app/ 아래에 있다.
//
//go:embed all:dist
var files embed.FS

// FS는 dist 폴더 안을 맨 위로 잘라 돌려준다: "index.html", "_app/..."처럼 바로 찾게.
func FS() fs.FS {
	sub, err := fs.Sub(files, "dist")
	if err != nil {
		// dist는 컴파일할 때 있는 게 확인된 폴더라 여기 올 일이 없다. 오면 빌드 자체가 잘못된 것이다.
		panic(err)
	}
	return sub
}
