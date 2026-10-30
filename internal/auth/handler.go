package auth

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/sagming40/cbt-journal/internal/apigen"
	"github.com/sagming40/cbt-journal/internal/clientip"
)

// Handler는 인증 · 계정 경로의 생성 인터페이스 메서드를 채운다.
// server의 handler가 이걸 임베드해서 StrictServerInterface를 완성한다 (ARCHITECTURE 2절).
// 예상된 실패(입력 오류 · 로그인 실패 · 제한)는 계약서에 적힌 응답 타입으로 돌려주고,
// 예상 못 한 오류만 error로 돌려준다 → server가 500 봉투로 바꾼다.
type Handler struct {
	svc *Service
}

// NewHandler는 인증 handler를 만든다.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// GetConfig는 로그인 화면이 알아야 하는 공개 설정이다.
func (h *Handler) GetConfig(_ context.Context, _ apigen.GetConfigRequestObject) (apigen.GetConfigResponseObject, error) {
	return apigen.GetConfig200JSONResponse{AllowSignup: h.svc.AllowSignup()}, nil
}

func (h *Handler) Login(ctx context.Context, req apigen.LoginRequestObject) (apigen.LoginResponseObject, error) {
	remember := req.Body.Remember != nil && *req.Body.Remember
	res, err := h.svc.Login(ctx, req.Body.Email, req.Body.Password, remember, clientip.FromContext(ctx))
	if err != nil {
		var ve *ValidationError
		var rl *RateLimitedError
		switch {
		case errors.As(err, &ve):
			return apigen.Login400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(validationBody(ve))}, nil
		case errors.As(err, &rl):
			return apigen.Login429JSONResponse{TooManyRequestsJSONResponse: rateLimited(rl)}, nil
		case errors.Is(err, ErrInvalidCredentials):
			// 없는 이메일 · 틀린 비밀번호는 같은 문구 (DESIGN 4.8절).
			return apigen.Login401JSONResponse{UnauthorizedJSONResponse: apigen.UnauthorizedJSONResponse(errorBody("INVALID_CREDENTIALS", err.Error()))}, nil
		}
		return nil, err
	}
	cookie := Cookie(res.Token, res.Expires, res.Remember).String()
	return apigen.Login200JSONResponse{Body: toMe(res.User), Headers: apigen.Login200ResponseHeaders{SetCookie: &cookie}}, nil
}

func (h *Handler) Signup(ctx context.Context, req apigen.SignupRequestObject) (apigen.SignupResponseObject, error) {
	res, err := h.svc.Signup(ctx, req.Body.Email, req.Body.Password, req.Body.Nickname, clientip.FromContext(ctx))
	if err != nil {
		var ve *ValidationError
		var rl *RateLimitedError
		switch {
		case errors.Is(err, ErrSignupDisabled):
			return apigen.Signup403JSONResponse{ForbiddenJSONResponse: apigen.ForbiddenJSONResponse(errorBody("SIGNUP_DISABLED", err.Error()))}, nil
		case errors.As(err, &ve):
			return apigen.Signup400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(validationBody(ve))}, nil
		case errors.As(err, &rl):
			return apigen.Signup429JSONResponse{TooManyRequestsJSONResponse: rateLimited(rl)}, nil
		case errors.Is(err, ErrEmailTaken):
			// 가입이 열려 있을 때만 닿는다. 공개 전(M8)에 이메일 확인과 같이 다시 본다(가입 여부가 드러남).
			return apigen.Signup409JSONResponse{ConflictJSONResponse: apigen.ConflictJSONResponse(errorBody("EMAIL_TAKEN", err.Error()))}, nil
		}
		return nil, err
	}
	cookie := Cookie(res.Token, res.Expires, false).String()
	return apigen.Signup201JSONResponse{Body: toMe(res.User), Headers: apigen.Signup201ResponseHeaders{SetCookie: &cookie}}, nil
}

// 아래 경로들은 openapi.yaml에서 잠긴 경로라, 세션 없이는 인증 미들웨어가 여기까지 오지 못하게 막는다.
// 그래도 세션이 없으면 401을 돌려준다 — 미들웨어 설정이 빠져도 열리지 않게 한 번 더.

func (h *Handler) Logout(ctx context.Context, _ apigen.LogoutRequestObject) (apigen.LogoutResponseObject, error) {
	sess := CurrentSession(ctx)
	if sess == nil {
		return apigen.Logout401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err := h.svc.Logout(ctx, sess); err != nil {
		return nil, err
	}
	cookie := ClearCookie().String()
	return apigen.Logout204Response{Headers: apigen.Logout204ResponseHeaders{SetCookie: &cookie}}, nil
}

func (h *Handler) GetMe(ctx context.Context, _ apigen.GetMeRequestObject) (apigen.GetMeResponseObject, error) {
	sess := CurrentSession(ctx)
	if sess == nil {
		return apigen.GetMe401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	u, err := h.svc.Me(ctx, sess.UserID)
	if errors.Is(err, ErrUserNotFound) {
		return apigen.GetMe401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	if err != nil {
		return nil, err
	}
	return apigen.GetMe200JSONResponse(toMe(u)), nil
}

func (h *Handler) UpdateMe(ctx context.Context, req apigen.UpdateMeRequestObject) (apigen.UpdateMeResponseObject, error) {
	sess := CurrentSession(ctx)
	if sess == nil {
		return apigen.UpdateMe401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	u, err := h.svc.UpdateProfile(ctx, sess.UserID, req.Body.Nickname, req.Body.Timezone)
	var ve *ValidationError
	switch {
	case errors.As(err, &ve):
		return apigen.UpdateMe400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(validationBody(ve))}, nil
	case errors.Is(err, ErrUserNotFound):
		return apigen.UpdateMe401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.UpdateMe200JSONResponse(toMe(u)), nil
}

func (h *Handler) ChangePassword(ctx context.Context, req apigen.ChangePasswordRequestObject) (apigen.ChangePasswordResponseObject, error) {
	sess := CurrentSession(ctx)
	if sess == nil {
		return apigen.ChangePassword401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	err := h.svc.ChangePassword(ctx, sess, req.Body.CurrentPassword, req.Body.NewPassword)
	var ve *ValidationError
	var rl *RateLimitedError
	switch {
	case errors.As(err, &ve):
		return apigen.ChangePassword400JSONResponse{BadRequestJSONResponse: apigen.BadRequestJSONResponse(validationBody(ve))}, nil
	case errors.As(err, &rl):
		return apigen.ChangePassword429JSONResponse{TooManyRequestsJSONResponse: rateLimited(rl)}, nil
	case errors.Is(err, ErrUserNotFound):
		return apigen.ChangePassword401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	case err != nil:
		return nil, err
	}
	return apigen.ChangePassword204Response{}, nil
}

func (h *Handler) LogoutOtherSessions(ctx context.Context, _ apigen.LogoutOtherSessionsRequestObject) (apigen.LogoutOtherSessionsResponseObject, error) {
	sess := CurrentSession(ctx)
	if sess == nil {
		return apigen.LogoutOtherSessions401JSONResponse{UnauthorizedJSONResponse: unauthorized()}, nil
	}
	n, err := h.svc.LogoutOthers(ctx, sess)
	if err != nil {
		return nil, err
	}
	return apigen.LogoutOtherSessions200JSONResponse{Revoked: n}, nil
}

func toMe(u User) apigen.Me {
	return apigen.Me{Id: u.ID.String(), Email: u.Email, Nickname: u.Nickname, Timezone: u.Timezone}
}

// errorBody는 오류 봉투(ARCHITECTURE 3.2절)를 계약서 타입으로 만든다.
func errorBody(code, message string) apigen.Error {
	var e apigen.Error
	e.Error.Code = code
	e.Error.Message = message
	return e
}

// UnauthorizedBody는 세션 없음 401 봉투다. 인증 미들웨어도 같은 모양을 쓴다.
func UnauthorizedBody() apigen.Error {
	return errorBody("UNAUTHORIZED", "로그인이 필요해요")
}

func unauthorized() apigen.UnauthorizedJSONResponse {
	return apigen.UnauthorizedJSONResponse(UnauthorizedBody())
}

func validationBody(ve *ValidationError) apigen.Error {
	e := errorBody("INVALID_INPUT", "입력한 내용을 다시 확인해 주세요")
	fields := ve.Fields
	e.Error.Fields = &fields
	return e
}

func rateLimited(rl *RateLimitedError) apigen.TooManyRequestsJSONResponse {
	// 남은 시간은 초 단위로 올림한다 — 0초로 내려 주면 화면이 곧바로 다시 시도한다.
	secs := int(math.Ceil(max(rl.RetryAfter, time.Second).Seconds()))
	return apigen.TooManyRequestsJSONResponse{
		Body:    errorBody("RATE_LIMITED", rl.Error()),
		Headers: apigen.TooManyRequestsResponseHeaders{RetryAfter: &secs},
	}
}
