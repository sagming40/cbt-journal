// Package server는 HTTP 라우터를 조립한다 (ARCHITECTURE 1 · 2 · 3절).
// 경로와 요청 · 응답 모양은 api/openapi.yaml이 원본이고, 여기서는 생성된 인터페이스를 채우고 미들웨어를 두른다.
package server

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"net/netip"
	"strings"

	"github.com/sagming40/cbt-journal/internal/apigen"
	"github.com/sagming40/cbt-journal/internal/auth"
	"github.com/sagming40/cbt-journal/internal/clientip"
)

// Checker는 연결 하나가 살아 있는지 확인한다. 살아 있으면 nil.
// 함수 타입으로 둬서 테스트에서는 가짜를 쉽게 꽂는다 (콘센트 규격만 맞으면 된다).
type Checker func(ctx context.Context) error

// Checks는 health가 확인할 연결들이다. 칸은 openapi.yaml의 Health 스키마와 짝을 이룬다.
type Checks struct {
	Postgres Checker
	Valkey   Checker
}

// Deps는 서버를 조립할 재료다.
type Deps struct {
	Logger         *slog.Logger
	Checks         Checks
	Web            fs.FS // 프론트 빌드 결과 (운영은 webdist.FS(), 테스트는 가짜 파일 묶음)
	Auth           *auth.Service
	TrustedProxies []netip.Prefix
}

// handler는 생성된 StrictServerInterface를 채운다.
// 기능 패키지의 Handler를 임베드하면 그 메서드들이 위로 올라와 인터페이스를 채운다 (ARCHITECTURE 2절).
type handler struct {
	*auth.Handler
	logger *slog.Logger
	checks Checks
}

// 컴파일할 때 확인: handler가 인터페이스의 메서드를 다 갖췄는가.
// openapi.yaml에 경로를 추가하고 생성을 다시 돌리면, 구현하기 전까지 여기서 빌드가 깨진다(일부러).
var _ apigen.StrictServerInterface = (*handler)(nil)

// New는 모든 경로를 등록한 핸들러 하나를 돌려준다.
//
//	바깥 mux ─ /api/ ──▶ 미들웨어 줄 ──▶ API mux (생성 코드가 경로 등록, 없는 GET은 JSON 404)
//	         └ GET / ──▶ 프론트 (SPA)
//
// 정적 파일 요청은 세션 조회 같은 API 미들웨어를 거치지 않는다 (ARCHITECTURE 1절).
func New(d Deps) http.Handler {
	h := &handler{Handler: auth.NewHandler(d.Auth), logger: d.Logger, checks: d.Checks}
	// 인증 서비스 없이 조립하는 건 health · 프론트만 보는 단위 테스트뿐이다.
	var sessions *auth.Sessions
	if d.Auth != nil {
		sessions = d.Auth.Sessions()
	}

	strict := apigen.NewStrictHandlerWithOptions(h, nil, apigen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  requestError(d.Logger),
		ResponseErrorHandlerFunc: internalError(d.Logger),
	})

	api := http.NewServeMux()
	apigen.HandlerWithOptions(strict, apigen.StdHTTPServerOptions{
		BaseRouter: api,
		// 경로별 미들웨어: 생성 코드가 인증 필요 경로에 표시를 넣은 뒤에 돈다 → 표시가 있는데 세션이 없으면 401.
		Middlewares:      []apigen.MiddlewareFunc{requireAuth},
		ErrorHandlerFunc: requestError(d.Logger),
	})
	// 계약서에 없는 GET /api/*는 JSON 404. GET으로만 여는 이유는 ARCHITECTURE 1절(405가 404로 바뀌지 않게).
	api.HandleFunc("GET /api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "없는 API 경로야", nil)
	})

	// 미들웨어 줄 (ARCHITECTURE 3.1절). 아래에서 위로 감싸므로, 실제로 도는 순서는 위에서 아래다.
	var apiChain http.Handler = api
	apiChain = sessionMiddleware(sessions, d.Logger)(apiChain) // 6
	apiChain = clientip.Middleware(d.TrustedProxies)(apiChain) // 5
	apiChain = crossOrigin(apiChain)                           // 4
	apiChain = limitBody(apiChain)                             // 3
	apiChain = recoverPanic(d.Logger)(apiChain)                // 2
	apiChain = accessLog(d.Logger)(apiChain)                   // 1 (가장 바깥)

	// 바깥은 mux 대신 경로 앞부분만 보고 나눈다: 메서드 없는 "/api/"와 "GET /"를 한 mux에 두면 서로 겹친다고 등록이 거부된다.
	web := http.NewServeMux()
	web.Handle("GET /", spa(d.Web)) // 프론트는 GET · HEAD만. 그 밖의 메서드는 mux가 405
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
			apiChain.ServeHTTP(w, r)
			return
		}
		web.ServeHTTP(w, r)
	})
}
