package record

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/sagming40/cbt-journal/internal/apierr"
	"github.com/sagming40/cbt-journal/internal/apigen"
)

// UserIDFunc는 지금 로그인한 사용자 id를 꺼낸다. 로그인 안 했으면 false.
// server가 조립할 때 꽂아 준다 — record가 auth를 import하지 않게 (ARCHITECTURE 2절 "사용자 id도 함수로").
type UserIDFunc func(ctx context.Context) (pgtype.UUID, bool)

// Handler는 사고 기록 · 초안 경로의 생성 인터페이스 메서드를 채운다. server의 handler가 임베드한다 (ARCHITECTURE 2절).
// 예상된 실패(입력 오류 · 없음 · 충돌)는 계약서에 적힌 응답 타입으로, 예상 못 한 오류만 error로(→ server가 500 봉투).
type Handler struct {
	svc    *Service
	userID UserIDFunc
}

// NewHandler는 사고 기록 handler를 만든다.
func NewHandler(svc *Service, userID UserIDFunc) *Handler {
	return &Handler{svc: svc, userID: userID}
}

// 아래 경로는 전부 openapi.yaml에서 잠긴 경로라 인증 미들웨어가 세션 없는 요청을 먼저 막는다.
// 그래도 사용자가 없으면 401 — 미들웨어 설정이 빠져도 열리지 않게 한 번 더.

func (h *Handler) ListDistortions(ctx context.Context, _ apigen.ListDistortionsRequestObject) (apigen.ListDistortionsResponseObject, error) {
	if _, ok := h.userID(ctx); !ok {
		return apigen.ListDistortions401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	ds, err := h.svc.Distortions(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]apigen.Distortion, len(ds))
	for i, d := range ds {
		items[i] = apigen.Distortion{Id: int(d.ID), Name: d.Name, Description: d.Description, Example: d.Example}
	}
	return apigen.ListDistortions200JSONResponse{Items: items}, nil
}

// ── 초안 ──

func (h *Handler) ListDrafts(ctx context.Context, _ apigen.ListDraftsRequestObject) (apigen.ListDraftsResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.ListDrafts401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	list, err := h.svc.ListDrafts(ctx, uid)
	switch {
	case errors.Is(err, ErrDraftUnavailable):
		return apigen.ListDrafts503JSONResponse{DraftsUnavailableJSONResponse: unavailable()}, nil
	case err != nil:
		return nil, err
	}
	items := make([]apigen.DraftSummary, len(list))
	for i, d := range list {
		items[i] = apigen.DraftSummary{Id: d.ID, Step: d.Step, SituationPreview: d.SituationPreview, UpdatedAt: d.UpdatedAt.UTC()}
	}
	return apigen.ListDrafts200JSONResponse{Items: items}, nil
}

func (h *Handler) CreateDraft(ctx context.Context, req apigen.CreateDraftRequestObject) (apigen.CreateDraftResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.CreateDraft401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	d, err := h.svc.CreateDraft(ctx, uid, toDraftContent(req.Body))
	var ve *apierr.ValidationError
	switch {
	case errors.As(err, &ve):
		return apigen.CreateDraft400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case errors.Is(err, ErrDraftLimit):
		return apigen.CreateDraft409JSONResponse{ConflictJSONResponse: apigen.ConflictJSONResponse(apierr.Body("DRAFT_LIMIT", ErrDraftLimit.Error()))}, nil
	case errors.Is(err, ErrDraftUnavailable):
		return apigen.CreateDraft503JSONResponse{DraftsUnavailableJSONResponse: unavailable()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.CreateDraft201JSONResponse(toAPIDraft(d)), nil
}

func (h *Handler) GetDraft(ctx context.Context, req apigen.GetDraftRequestObject) (apigen.GetDraftResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.GetDraft401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	d, err := h.svc.GetDraft(ctx, uid, req.Id)
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.GetDraft404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case errors.Is(err, ErrDraftUnavailable):
		return apigen.GetDraft503JSONResponse{DraftsUnavailableJSONResponse: unavailable()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.GetDraft200JSONResponse(toAPIDraft(d)), nil
}

func (h *Handler) SaveDraft(ctx context.Context, req apigen.SaveDraftRequestObject) (apigen.SaveDraftResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.SaveDraft401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	var rev int
	var content DraftContent
	if b := req.Body; b != nil {
		rev = b.Rev
		content = toDraftContent(&apigen.DraftInput{
			Step: b.Step, OccurredAt: b.OccurredAt, Situation: b.Situation, AutomaticThought: b.AutomaticThought,
			EvidenceFor: b.EvidenceFor, EvidenceAgainst: b.EvidenceAgainst, AlternativeThought: b.AlternativeThought,
			Emotions: b.Emotions, DistortionIds: b.DistortionIds,
		})
	}
	d, err := h.svc.SaveDraft(ctx, uid, req.Id, rev, content)
	var ve *apierr.ValidationError
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.SaveDraft404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case errors.As(err, &ve):
		return apigen.SaveDraft400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case errors.Is(err, ErrDraftConflict):
		return apigen.SaveDraft409JSONResponse{ConflictJSONResponse: apigen.ConflictJSONResponse(apierr.Body("DRAFT_CONFLICT", ErrDraftConflict.Error()))}, nil
	case errors.Is(err, ErrDraftUnavailable):
		return apigen.SaveDraft503JSONResponse{DraftsUnavailableJSONResponse: unavailable()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.SaveDraft200JSONResponse(toAPIDraft(d)), nil
}

func (h *Handler) DeleteDraft(ctx context.Context, req apigen.DeleteDraftRequestObject) (apigen.DeleteDraftResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.DeleteDraft401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	err := h.svc.DeleteDraft(ctx, uid, req.Id)
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.DeleteDraft404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case errors.Is(err, ErrDraftUnavailable):
		return apigen.DeleteDraft503JSONResponse{DraftsUnavailableJSONResponse: unavailable()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.DeleteDraft204Response{}, nil
}

// ── 완료한 기록 ──

func (h *Handler) ListRecords(ctx context.Context, req apigen.ListRecordsRequestObject) (apigen.ListRecordsResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.ListRecords401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	date := ""
	if req.Params.Date != nil {
		date = *req.Params.Date
	}
	day, items, err := h.svc.Day(ctx, uid, date)
	var ve *apierr.ValidationError
	switch {
	case errors.As(err, &ve):
		return apigen.ListRecords400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case err != nil:
		return nil, err
	}
	out := make([]apigen.Record, len(items))
	for i, r := range items {
		out[i] = toAPI(r)
	}
	return apigen.ListRecords200JSONResponse{Date: day, Items: out}, nil
}

func (h *Handler) CreateRecord(ctx context.Context, req apigen.CreateRecordRequestObject) (apigen.CreateRecordResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.CreateRecord401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	var in Input
	draftID := ""
	if b := req.Body; b != nil {
		in = toInput(&apigen.RecordInput{
			OccurredAt: b.OccurredAt, Situation: b.Situation, AutomaticThought: b.AutomaticThought,
			EvidenceFor: b.EvidenceFor, EvidenceAgainst: b.EvidenceAgainst, AlternativeThought: b.AlternativeThought,
			Emotions: b.Emotions, DistortionIds: b.DistortionIds,
		})
		if b.DraftId != nil {
			draftID = *b.DraftId
		}
	}
	r, created, err := h.svc.Create(ctx, uid, in, draftID)
	var ve *apierr.ValidationError
	switch {
	case errors.As(err, &ve):
		return apigen.CreateRecord400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case err != nil:
		return nil, err
	}
	if created {
		return apigen.CreateRecord201JSONResponse(toAPI(r)), nil
	}
	return apigen.CreateRecord200JSONResponse(toAPI(r)), nil
}

func (h *Handler) GetRecord(ctx context.Context, req apigen.GetRecordRequestObject) (apigen.GetRecordResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.GetRecord401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	r, err := h.svc.Get(ctx, uid, req.Id)
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.GetRecord404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.GetRecord200JSONResponse(toAPI(r)), nil
}

func (h *Handler) UpdateRecord(ctx context.Context, req apigen.UpdateRecordRequestObject) (apigen.UpdateRecordResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.UpdateRecord401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	r, err := h.svc.Update(ctx, uid, req.Id, toInput(req.Body))
	var ve *apierr.ValidationError
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.UpdateRecord404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case errors.As(err, &ve):
		return apigen.UpdateRecord400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(apierr.Validation(ve))}, nil
	case err != nil:
		return nil, err
	}
	return apigen.UpdateRecord200JSONResponse(toAPI(r)), nil
}

func (h *Handler) DeleteRecord(ctx context.Context, req apigen.DeleteRecordRequestObject) (apigen.DeleteRecordResponseObject, error) {
	uid, ok := h.userID(ctx)
	if !ok {
		return apigen.DeleteRecord401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	err := h.svc.Delete(ctx, uid, req.Id)
	switch {
	case errors.Is(err, ErrNotFound):
		return apigen.DeleteRecord404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.DeleteRecord204Response{}, nil
}

// ── 계약서 타입 ↔ 서비스 타입 ──

// toInput은 계약서 입력을 서비스 입력으로 옮긴다. 본문이 비어 있으면 빈 입력 → 서비스가 칸별 오류로 막는다.
func toInput(b *apigen.RecordInput) Input {
	if b == nil {
		return Input{}
	}
	in := Input{
		OccurredAt: b.OccurredAt, Situation: b.Situation, AutomaticThought: b.AutomaticThought,
		EvidenceFor: b.EvidenceFor, EvidenceAgainst: b.EvidenceAgainst, AlternativeThought: b.AlternativeThought,
		Emotions: make([]InputEmotion, len(b.Emotions)), DistortionIDs: b.DistortionIds,
	}
	for i, e := range b.Emotions {
		in.Emotions[i] = InputEmotion{EmotionID: e.EmotionId, Before: e.IntensityBefore, After: e.IntensityAfter}
	}
	return in
}

func toDraftContent(b *apigen.DraftInput) DraftContent {
	if b == nil {
		return DraftContent{}
	}
	c := DraftContent{
		Step: b.Step, OccurredAt: b.OccurredAt, Situation: b.Situation, AutomaticThought: b.AutomaticThought,
		EvidenceFor: b.EvidenceFor, EvidenceAgainst: b.EvidenceAgainst, AlternativeThought: b.AlternativeThought,
		Emotions: make([]DraftEmotion, len(b.Emotions)), DistortionIDs: b.DistortionIds,
	}
	for i, e := range b.Emotions {
		c.Emotions[i] = DraftEmotion{EmotionID: e.EmotionId, Before: e.IntensityBefore, After: e.IntensityAfter}
	}
	return c
}

func toAPIDraft(d Draft) apigen.Draft {
	ems := make([]apigen.RecordEmotionInput, len(d.Emotions))
	for i, e := range d.Emotions {
		ems[i] = apigen.RecordEmotionInput{EmotionId: e.EmotionID, IntensityBefore: e.Before, IntensityAfter: e.After}
	}
	ids := d.DistortionIDs
	if ids == nil {
		ids = []int{}
	}
	return apigen.Draft{
		Id: d.ID, Rev: d.Rev, UpdatedAt: d.UpdatedAt.UTC(), Step: d.Step, OccurredAt: d.OccurredAt.UTC(),
		Situation: d.Situation, AutomaticThought: d.AutomaticThought, EvidenceFor: d.EvidenceFor,
		EvidenceAgainst: d.EvidenceAgainst, AlternativeThought: d.AlternativeThought,
		Emotions: ems, DistortionIds: ids,
	}
}

func toAPI(r Record) apigen.Record {
	ems := make([]apigen.RecordEmotion, len(r.Emotions))
	for i, e := range r.Emotions {
		ems[i] = apigen.RecordEmotion{EmotionId: int(e.EmotionID), Name: e.Name, IntensityBefore: int(e.Before), IntensityAfter: int(e.After)}
	}
	ds := make([]apigen.RecordDistortion, len(r.Distortions))
	for i, d := range r.Distortions {
		ds[i] = apigen.RecordDistortion{Id: int(d.ID), Name: d.Name}
	}
	// 시각은 RFC 3339 UTC로 (ARCHITECTURE 3.2절)
	return apigen.Record{
		Id: r.ID.String(), OccurredAt: r.OccurredAt.UTC(), Situation: r.Situation, AutomaticThought: r.AutomaticThought,
		EvidenceFor: r.EvidenceFor, EvidenceAgainst: r.EvidenceAgainst, AlternativeThought: r.AlternativeThought,
		Emotions: ems, Distortions: ds, CreatedAt: r.CreatedAt.UTC(), UpdatedAt: r.UpdatedAt.UTC(),
	}
}

func unauthorized() apigen.UnauthorizedJSONResponse {
	return apigen.UnauthorizedJSONResponse(apierr.Unauthorized())
}

func notFound() apigen.NotFoundJSONResponse {
	return apigen.NotFoundJSONResponse(apierr.NotFound())
}

func unavailable() apigen.DraftsUnavailableJSONResponse {
	return apigen.DraftsUnavailableJSONResponse(apierr.Body("DRAFTS_UNAVAILABLE", ErrDraftUnavailable.Error()))
}
