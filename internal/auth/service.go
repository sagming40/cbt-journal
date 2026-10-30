package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/sagming40/cbt-journal/internal/db/dbgen"
)

// 서비스가 돌려주는 예상된 실패들. handler가 이걸 보고 상태 코드를 정한다.
var (
	ErrInvalidCredentials = errors.New("이메일 또는 비밀번호가 맞지 않아")
	ErrSignupDisabled     = errors.New("지금은 가입을 받지 않아")
	ErrEmailTaken         = errors.New("이미 가입된 이메일이야")
	ErrUserNotFound       = errors.New("사용자가 없다")
)

// RateLimitedError는 요청 제한에 걸렸다는 뜻이다. RetryAfter 뒤에 다시 시도할 수 있다.
type RateLimitedError struct{ RetryAfter time.Duration }

func (e *RateLimitedError) Error() string { return "잠시 뒤에 다시 시도해 줘" }

// User는 화면 · CLI에 돌려줄 사용자 정보다 (비밀번호 해시는 밖으로 내보내지 않는다).
type User struct {
	ID       pgtype.UUID
	Email    string
	Nickname string
	Timezone string
}

func toUser(u dbgen.User) User {
	return User{ID: u.ID, Email: u.Email, Nickname: u.Nickname, Timezone: u.Timezone}
}

// Service는 로그인 · 가입 · 계정 규칙을 한곳에 모은다 (handler → service → db, ARCHITECTURE 2절).
type Service struct {
	pool        *pgxpool.Pool
	q           *dbgen.Queries
	sessions    *Sessions
	limits      *limiter
	breach      BreachChecker
	logger      *slog.Logger
	allowSignup bool
}

// Options는 Service를 만들 때 넘기는 값이다.
type Options struct {
	Pool        *pgxpool.Pool
	Valkey      *redis.Client
	Breach      BreachChecker // nil이면 유출 확인을 건너뛴다 (테스트용)
	Logger      *slog.Logger
	AllowSignup bool
}

// NewService는 인증 서비스를 만든다.
func NewService(o Options) *Service {
	return &Service{
		pool:        o.Pool,
		q:           dbgen.New(o.Pool),
		sessions:    NewSessions(dbgen.New(o.Pool), o.Valkey, o.Logger),
		limits:      &limiter{rdb: o.Valkey, logger: o.Logger},
		breach:      o.Breach,
		logger:      o.Logger,
		allowSignup: o.AllowSignup,
	}
}

// Sessions는 세션 미들웨어가 쓸 저장소다.
func (s *Service) Sessions() *Sessions { return s.sessions }

// AllowSignup은 가입이 열려 있는지 알려 준다 (GET /api/config).
func (s *Service) AllowSignup() bool { return s.allowSignup }

// LoginResult는 로그인 · 가입이 성공했을 때 쿠키를 만들 재료다.
type LoginResult struct {
	User     User
	Token    string
	Expires  time.Time
	Remember bool
}

// Login은 이메일 · 비밀번호를 확인하고 새 세션을 만든다.
// 없는 이메일 · 틀린 비밀번호는 같은 오류이고, 걸리는 시간도 비슷하게 맞춘다 (ARCHITECTURE 4절).
func (s *Service) Login(ctx context.Context, rawEmail, password string, remember bool, ip netip.Addr) (LoginResult, error) {
	email, err := NormalizeEmail(rawEmail)
	if err != nil {
		return LoginResult{}, err
	}
	if password == "" {
		return LoginResult{}, fieldError("password", "비밀번호를 적어 줘")
	}

	// 제한 검사는 비밀번호 확인 "전에" 한다: 잠긴 동안은 맞는 비밀번호도 429 — 맞았는지 알려 주지 않는다.
	ipKey, acctKey := loginIPKey(ip), loginAcctKey(email)
	if blocked, wait := s.limits.blocked(ctx, ipKey, loginIPRule); blocked {
		return LoginResult{}, &RateLimitedError{RetryAfter: wait}
	}
	if blocked, wait := s.limits.blocked(ctx, acctKey, loginAcctRule); blocked {
		return LoginResult{}, &RateLimitedError{RetryAfter: wait}
	}
	s.limits.hit(ctx, ipKey, loginIPRule)

	user, err := s.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		// 없는 계정도 같은 무게의 계산을 한 번 한다 — 응답 시간으로 가입 여부가 드러나지 않게.
		_, _ = VerifyPassword(password, dummyHash())
		s.limits.hit(ctx, acctKey, loginAcctRule)
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, err
	}
	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return LoginResult{}, fmt.Errorf("저장된 해시 확인: %w", err)
	}
	if !ok {
		s.limits.hit(ctx, acctKey, loginAcctRule)
		return LoginResult{}, ErrInvalidCredentials
	}
	s.limits.reset(ctx, acctKey)

	// 파라미터를 올렸다면 맞는 비밀번호를 아는 지금 조용히 새로 해시한다. 실패해도 로그인은 계속.
	if NeedsRehash(user.PasswordHash) {
		if h, err := HashPassword(password); err == nil {
			if _, err := s.q.UpdateUserPassword(ctx, dbgen.UpdateUserPasswordParams{ID: user.ID, PasswordHash: h}); err != nil {
				s.logger.Warn("비밀번호 해시를 새 파라미터로 바꾸지 못했다", "user_id", user.ID.String(), "err", err)
			}
		}
	}

	// 그 사용자의 만료된 세션은 여기서 치운다 (전체 정리는 M7 worker, DATA_MODEL 3.2절).
	if err := s.q.DeleteExpiredUserSessions(ctx, user.ID); err != nil {
		s.logger.Warn("만료된 세션을 지우지 못했다", "user_id", user.ID.String(), "err", err)
	}

	token, expires, err := s.sessions.create(ctx, s.q, user.ID, remember)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{User: toUser(user), Token: token, Expires: expires, Remember: remember}, nil
}

// Signup은 가입시키고 바로 로그인 세션을 만든다(로그인 유지 안 함). ALLOW_SIGNUP이 꺼져 있으면 ErrSignupDisabled.
func (s *Service) Signup(ctx context.Context, rawEmail, password, rawNickname string, ip netip.Addr) (LoginResult, error) {
	if !s.allowSignup {
		return LoginResult{}, ErrSignupDisabled
	}
	ipKey := signupIPKey(ip)
	if blocked, wait := s.limits.blocked(ctx, ipKey, signupIPRule); blocked {
		return LoginResult{}, &RateLimitedError{RetryAfter: wait}
	}
	s.limits.hit(ctx, ipKey, signupIPRule)

	user, err := s.createUser(ctx, rawEmail, password, rawNickname)
	if err != nil {
		return LoginResult{}, err
	}
	token, expires, err := s.sessions.create(ctx, s.q, user.ID, false)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{User: user, Token: token, Expires: expires}, nil
}

// CreateUser는 관리 명령(app user create)용이다. 가입 플래그 · 요청 제한과 상관없이 같은 입력 규칙을 쓴다 (REQUIREMENTS A9 · A10).
func (s *Service) CreateUser(ctx context.Context, rawEmail, password, rawNickname string) (User, error) {
	return s.createUser(ctx, rawEmail, password, rawNickname)
}

func (s *Service) createUser(ctx context.Context, rawEmail, password, rawNickname string) (User, error) {
	// 칸마다 오류를 모아 한 번에 알려 준다.
	fields := map[string]string{}
	email, err := NormalizeEmail(rawEmail)
	mergeFields(fields, err)
	nickname, err := NormalizeNickname(rawNickname)
	mergeFields(fields, err)
	mergeFields(fields, ValidatePassword("password", password))
	if len(fields) == 0 {
		mergeFields(fields, s.checkBreached(ctx, "password", password))
	}
	if len(fields) > 0 {
		return User{}, &ValidationError{Fields: fields}
	}

	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	u, err := s.q.CreateUser(ctx, dbgen.CreateUserParams{Email: email, PasswordHash: hash, Nickname: nickname})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation: users.email
		return User{}, ErrEmailTaken
	}
	if err != nil {
		return User{}, err
	}
	return toUser(u), nil
}

// Logout은 지금 세션을 지운다.
func (s *Service) Logout(ctx context.Context, sess *Session) error {
	hashes, err := s.q.DeleteSession(ctx, sess.TokenHash)
	if err != nil {
		return err
	}
	s.sessions.forget(ctx, hashes)
	return nil
}

// Me는 내 정보를 돌려준다. 세션은 있는데 사용자가 없으면(그 사이 지워짐) ErrUserNotFound.
func (s *Service) Me(ctx context.Context, userID pgtype.UUID) (User, error) {
	u, err := s.q.GetUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	return toUser(u), nil
}

// UpdateProfile은 보낸 칸(nil이 아닌 것)만 바꾼다 (A4).
func (s *Service) UpdateProfile(ctx context.Context, userID pgtype.UUID, nickname, timezone *string) (User, error) {
	cur, err := s.q.GetUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, err
	}
	next := dbgen.UpdateUserProfileParams{ID: userID, Nickname: cur.Nickname, Timezone: cur.Timezone}
	fields := map[string]string{}
	if nickname != nil {
		n, err := NormalizeNickname(*nickname)
		mergeFields(fields, err)
		next.Nickname = n
	}
	if timezone != nil {
		mergeFields(fields, ValidateTimezone(*timezone))
		next.Timezone = *timezone
	}
	if len(fields) > 0 {
		return User{}, &ValidationError{Fields: fields}
	}
	u, err := s.q.UpdateUserProfile(ctx, next)
	if err != nil {
		return User{}, err
	}
	return toUser(u), nil
}

// ChangePassword는 현재 비밀번호를 확인하고 바꾼 뒤, 지금 세션을 뺀 다른 세션을 모두 끊는다 (A3).
func (s *Service) ChangePassword(ctx context.Context, sess *Session, current, next string) error {
	key := pwChangeKey(sess.UserID.String())
	if blocked, wait := s.limits.blocked(ctx, key, pwChangeRule); blocked {
		return &RateLimitedError{RetryAfter: wait}
	}

	user, err := s.q.GetUserByID(ctx, sess.UserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	ok, err := VerifyPassword(current, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("저장된 해시 확인: %w", err)
	}
	if !ok {
		// 세션을 훔친 사람이 현재 비밀번호를 무차별로 넣어 보지 못하게 센다 (REQUIREMENTS 5.2절).
		s.limits.hit(ctx, key, pwChangeRule)
		return fieldError("current_password", "지금 비밀번호가 맞지 않아")
	}
	if err := ValidatePassword("new_password", next); err != nil {
		return err
	}
	if next == current {
		return fieldError("new_password", "지금 비밀번호와 다른 걸로 골라 줘")
	}
	if err := s.checkBreached(ctx, "new_password", next); err != nil {
		return err
	}
	hash, err := HashPassword(next)
	if err != nil {
		return err
	}

	// 비밀번호 변경과 다른 세션 삭제는 한 트랜잭션 — 비밀번호만 바뀌고 옛 세션이 남는 일이 없게.
	var hashes [][]byte
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if _, err := q.UpdateUserPassword(ctx, dbgen.UpdateUserPasswordParams{ID: sess.UserID, PasswordHash: hash}); err != nil {
			return err
		}
		hashes, err = q.DeleteOtherSessions(ctx, dbgen.DeleteOtherSessionsParams{UserID: sess.UserID, KeepTokenHash: sess.TokenHash})
		return err
	})
	if err != nil {
		return err
	}
	// 캐시는 커밋이 끝난 뒤에 지운다.
	s.sessions.forget(ctx, hashes)
	s.limits.reset(ctx, key)
	return nil
}

// LogoutOthers는 지금 세션을 뺀 내 세션을 모두 끊고 끊은 수를 돌려준다 (A8).
func (s *Service) LogoutOthers(ctx context.Context, sess *Session) (int, error) {
	hashes, err := s.q.DeleteOtherSessions(ctx, dbgen.DeleteOtherSessionsParams{UserID: sess.UserID, KeepTokenHash: sess.TokenHash})
	if err != nil {
		return 0, err
	}
	s.sessions.forget(ctx, hashes)
	return len(hashes), nil
}

// ResetPassword는 관리 명령(app user reset-password)용이다. 비밀번호를 바꾸고 그 사용자의 세션을 전부 끊는다 (REQUIREMENTS A9).
func (s *Service) ResetPassword(ctx context.Context, rawEmail, password string) (int, error) {
	email, err := NormalizeEmail(rawEmail)
	if err != nil {
		return 0, err
	}
	if err := ValidatePassword("password", password); err != nil {
		return 0, err
	}
	if err := s.checkBreached(ctx, "password", password); err != nil {
		return 0, err
	}
	user, err := s.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return 0, err
	}
	var hashes [][]byte
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if _, err := q.UpdateUserPassword(ctx, dbgen.UpdateUserPasswordParams{ID: user.ID, PasswordHash: hash}); err != nil {
			return err
		}
		hashes, err = q.DeleteUserSessions(ctx, user.ID)
		return err
	})
	if err != nil {
		return 0, err
	}
	s.sessions.forget(ctx, hashes)
	return len(hashes), nil
}

// checkBreached는 유출 목록에 있으면 입력 오류를 돌려준다.
// 확인 자체가 실패하면(타임아웃 · 500 등) 경고만 남기고 통과시킨다 — 남의 장애로 가입 · 변경이 막히지 않게 (REQUIREMENTS A10).
func (s *Service) checkBreached(ctx context.Context, field, password string) error {
	if s.breach == nil {
		return nil
	}
	breached, err := s.breach.IsBreached(ctx, password)
	if err != nil {
		// 오류 문구에 요청 주소(해시 앞 5글자)가 섞이지 않게, URL 오류면 안쪽 원인만 남긴다.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		s.logger.Warn("유출 비밀번호 확인을 건너뛴다", "err", err)
		return nil
	}
	if breached {
		return fieldError(field, "알려진 유출 목록에 있는 비밀번호야. 다른 걸로 골라 줘")
	}
	return nil
}

func mergeFields(dst map[string]string, err error) {
	var ve *ValidationError
	if errors.As(err, &ve) {
		for k, v := range ve.Fields {
			dst[k] = v
		}
	}
}
