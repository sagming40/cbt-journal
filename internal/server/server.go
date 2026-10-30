// Package server는 HTTP 라우터를 조립한다 (ARCHITECTURE 2절).
// 경로와 요청 · 응답 모양은 api/openapi.yaml이 원본이고, 여기서는 생성된 인터페이스를 채우기만 한다.
// 미들웨어 · 오류 봉투(3.1 · 3.2절)는 M2에서 붙인다.
package server

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/sagming40/cbt-journal/internal/apigen"
)

// Checker는 연결 하나가 살아 있는지 확인한다. 살아 있으면 nil.
// 함수 타입으로 둬서 테스트에서는 가짜를 쉽게 꽂는다 (콘센트 규격만 맞으면 된다).
type Checker func(ctx context.Context) error

// Checks는 health가 확인할 연결들이다. 칸은 openapi.yaml의 Health 스키마와 짝을 이룬다.
type Checks struct {
	Postgres Checker
	Valkey   Checker
}

// New는 모든 경로를 등록한 핸들러 하나를 돌려준다.
// web: 프론트 빌드 결과 (운영은 webdist.FS(), 테스트는 가짜 파일 묶음).
func New(logger *slog.Logger, checks Checks, web fs.FS) http.Handler {
	mux := http.NewServeMux()
	h := &handler{logger: logger, checks: checks}
	// 생성된 코드가 openapi.yaml의 경로를 mux에 등록한다. 경로를 손으로 적지 않으니 문서와 코드가 어긋날 수 없다.
	// 비유: 안내 데스크 직원이 계약서(openapi.yaml)에 적힌 방 번호만 보고 손님을 보낸다.
	apigen.HandlerFromMux(apigen.NewStrictHandler(h, nil), mux)
	// 나머지 GET은 프론트가 받는다. "GET /"는 가장 덜 구체적인 패턴이라 GET /api/health 같은 경로가 먼저 고른다.
	// GET으로만 여는 이유: 메서드 없이 "/"로 열면 POST /api/health도 여기로 빠져 405 대신 404가 된다.
	mux.Handle("GET /", spa(web))
	return mux
}

// handler는 생성된 StrictServerInterface의 빈칸(메서드)을 채운다. 경로마다 파일 하나에 메서드 하나.
type handler struct {
	logger *slog.Logger
	checks Checks
}

// 컴파일할 때 확인: handler가 인터페이스의 메서드를 다 갖췄는가.
// openapi.yaml에 경로를 추가하고 생성을 다시 돌리면, 구현하기 전까지 여기서 빌드가 깨진다(일부러).
var _ apigen.StrictServerInterface = (*handler)(nil)
