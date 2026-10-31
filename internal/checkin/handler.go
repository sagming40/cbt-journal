package checkin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sagming40/cbt-journal/internal/apierr"
	"github.com/sagming40/cbt-journal/internal/apigen"
)

// UserIDFunc는 지금 로그인한 사용자 id를 꺼낸다. 로그인 안 했으면 false.
// server가 조립할 때 꽂아 준다 — checkin이 auth를 import하지 않게 (ARCHITECTURE 2절 "사용자 id도 함수로").
type UserIDFunc func(ctx context.Context) (pgtype.UUID, bool)

// Handler는 체크인 경로의 생성 인터페이스 메서드를 채운다. server의 handler가 임베드한다 (ARCHITECTURE 2절).
// 예상된 실패(입력 오류 · 없음)는 계약서에 적힌 응답 타입으로, 예상 못 한 오류만 error로(→ server가 500 봉투).
type Handler struct {
	svc    *Service
	userID UserIDFunc
}

// NewHandler는 체크인 handler를 만든다.
func NewHandler(svc *Service, userID UserIDFunc) *Handler {
	return &Handler{svc: svc, userID: userID}
}

// 아래 경로는 전부 openapi.yaml에서 잠긴 경로라 인증 미들웨어가 세션 없는 요청을 먼저 막는다.
// 그래도 사용자가 없으면 401 — 미들웨어 설정이 빠져도 열리지 않게 한 번 더.

func (h *Handler) ListEmotions(ctx context.Context, _ apigen.ListEmotionsRequestObject) (apigen.ListEmotionsResponseObject, error) {
	if _, ok := h.userID(ctx); !ok {
		return apigen.ListEmotions401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	ems, err := h.svc.Emotions(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]apigen.Emotion, len(ems))
	for i, e := range ems {
		items[i] = apigen.Emotion{Id: int(e.ID), Name: e.Name}
	}
	return apigen.ListEmotions200JSONResponse{Items: items}, nil
}

func (h *Handler) ListCheckins(ctx context.Context, req apigen.ListCheckinsRequestObject) (apigen.ListCheckinsResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.ListCheckins401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	date := ""
	if req.Params.Date != nil {
		date = *req.Params.Date
	}
	day, items, err := h.svc.Day(ctx, uid, date)
	var ve *apierr.ValidationError
	switch {
	case errors.As(err, &ve):
		return apigen.ListCheckins400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case err != nil:
		return nil, err
	}
	out := make([]apigen.Checkin, len(items))
	for i, c := range items {
		out[i] = toAPI(c)
	}
	return apigen.ListCheckins200JSONResponse{Date: day, Items: out}, nil
}

func (h *Handler) CreateCheckin(ctx context.Context, req apigen.CreateCheckinRequestObject) (apigen.CreateCheckinResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.CreateCheckin401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	c, err := h.svc.Create(ctx, uid, toInput(req.Body))
	var ve *apierr.ValidationError
	switch {
	case errors.As(err, &ve):
		return apigen.CreateCheckin400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case err != nil:
		return nil, err
	}
	return apigen.CreateCheckin201JSONResponse(toAPI(c)), nil
}

func (h *Handler) GetCheckin(ctx context.Context, req apigen.GetCheckinRequestObject) (apigen.GetCheckinResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.GetCheckin401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	c, err := h.svc.Get(ctx, uid, req.Id)
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.GetCheckin404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.GetCheckin200JSONResponse(toAPI(c)), nil
}

func (h *Handler) UpdateCheckin(ctx context.Context, req apigen.UpdateCheckinRequestObject) (apigen.UpdateCheckinResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.UpdateCheckin401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	c, err := h.svc.Update(ctx, uid, req.Id, toInput(req.Body))
	var ve *apierr.ValidationError
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.UpdateCheckin404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case errors.As(err, &ve):
		return apigen.UpdateCheckin400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case err != nil:
		return nil, err
	}
	return apigen.UpdateCheckin200JSONResponse(toAPI(c)), nil
}

func (h *Handler) DeleteCheckin(ctx context.Context, req apigen.DeleteCheckinRequestObject) (apigen.DeleteCheckinResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.DeleteCheckin401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	err := h.svc.Delete(ctx, uid, req.Id)
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.DeleteCheckin404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.DeleteCheckin204Response{}, nil
}

// toInput은 계약서 입력을 서비스 입력으로 옮긴다. 본문이 비어 있으면 빈 입력 → 서비스가 "감정을 골라 주세요"로 막는다.
func toInput(b *apigen.CheckinInput) Input {
	if b == nil {
		return Input{}
	}
	in := Input{OccurredAt: b.OccurredAt, Emotions: make([]InputEmotion, len(b.Emotions))}
	if b.Note != nil {
		in.Note = *b.Note
	}
	for i, e := range b.Emotions {
		in.Emotions[i] = InputEmotion{EmotionID: e.EmotionId, Intensity: e.Intensity}
	}
	return in
}

func toAPI(c Checkin) apigen.Checkin {
	ems := make([]apigen.CheckinEmotion, len(c.Emotions))
	for i, e := range c.Emotions {
		ems[i] = apigen.CheckinEmotion{EmotionId: int(e.EmotionID), Name: e.Name, Intensity: int(e.Intensity)}
	}
	// 시각은 RFC 3339 UTC로 (ARCHITECTURE 3.2절)
	return apigen.Checkin{Id: c.ID.String(), OccurredAt: c.OccurredAt.UTC(), Note: c.Note, Emotions: ems}
}

func unauthorized() apigen.UnauthorizedJSONResponse {
	return apigen.UnauthorizedJSONResponse(apierr.Unauthorized())
}

func notFound() apigen.NotFoundJSONResponse {
	return apigen.NotFoundJSONResponse(apierr.NotFound())
}
