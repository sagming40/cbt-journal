package server

import (
	"context"
	"time"

	"github.com/sagming40/cbt-journal/internal/apigen"
)

// GetHealth는 각 연결을 확인한다. 하나라도 실패하면 503 (openapi.yaml getHealth).
func (h *handler) GetHealth(ctx context.Context, _ apigen.GetHealthRequestObject) (apigen.GetHealthResponseObject, error) {
	// 한쪽이 멈춰도 health 요청 자체가 매달리지 않게 2초로 자른다.
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	body := apigen.Health{
		Postgres: h.check(ctx, "postgres", h.checks.Postgres),
		Valkey:   h.check(ctx, "valkey", h.checks.Valkey),
	}
	if body.Postgres == "ok" && body.Valkey == "ok" {
		body.Status = "ok"
		return apigen.GetHealth200JSONResponse(body), nil
	}
	body.Status = "error"
	// 계약서에 적힌 503만 타입으로 존재한다. 적히지 않은 상태 코드는 돌려줄 수조차 없다.
	return apigen.GetHealth503JSONResponse(body), nil
}

// check는 연결 하나를 확인해 "ok" / "error"로 바꾼다.
// 원인(주소 · 사용자 이름이 섞일 수 있음)은 로그에만 남긴다 (ARCHITECTURE 3.1절).
func (h *handler) check(ctx context.Context, name string, c Checker) apigen.CheckStatus {
	if err := c(ctx); err != nil {
		h.logger.Error("health 확인 실패", "target", name, "err", err)
		return "error"
	}
	return "ok"
}
