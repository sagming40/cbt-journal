package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"

	"github.com/sagming40/cbt-journal/internal/db/dbgen"
)

// 세션 기한 (DATA_MODEL 3.2절).
const (
	rememberTTL   = 30 * 24 * time.Hour // 로그인 유지: 마지막 사용 후 30일
	sessionTTL    = 12 * time.Hour      // 유지 안 함: 마지막 사용 후 12시간
	rememberTouch = 24 * time.Hour      // 이만큼 지났을 때만 기한을 늘린다 (요청마다 DB에 쓰지 않게)
	sessionTouch  = 10 * time.Minute
	cacheTTL      = 10 * time.Minute // Valkey sess: 캐시 (DATA_MODEL 4.2절)

	// CookieName은 세션 쿠키 이름이다. openapi.yaml의 cookieAuth와 맞춘다.
	CookieName = "session"
)

// Session은 지금 요청의 로그인 세션이다. 세션 미들웨어가 context에 넣는다.
type Session struct {
	UserID    pgtype.UUID
	TokenHash []byte
	ExpiresAt time.Time
}

type ctxKey struct{}

// WithSession은 세션을 context에 넣는다.
func WithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

// CurrentSession은 context의 세션을 꺼낸다. 로그인 안 했으면 nil.
func CurrentSession(ctx context.Context) *Session {
	s, _ := ctx.Value(ctxKey{}).(*Session)
	return s
}

// Sessions는 세션을 Postgres(원본)와 Valkey(조회 캐시)에 나눠 둔다 (decisions 0005, ARCHITECTURE 4절).
// 비유: 장부(Postgres)는 금고에, 자주 보는 쪽지(Valkey)는 책상 위에. 쪽지가 없어지면 장부를 다시 본다.
type Sessions struct {
	q      *dbgen.Queries
	rdb    *redis.Client
	logger *slog.Logger
	now    func() time.Time
}

// NewSessions는 세션 저장소를 만든다.
func NewSessions(q *dbgen.Queries, rdb *redis.Client, logger *slog.Logger) *Sessions {
	return &Sessions{q: q, rdb: rdb, logger: logger, now: time.Now}
}

type cachedSession struct {
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// create는 새 세션을 만들고 쿠키에 넣을 토큰과 만료 시각을 돌려준다.
func (s *Sessions) create(ctx context.Context, q *dbgen.Queries, userID pgtype.UUID, remember bool) (token string, expires time.Time, err error) {
	token, hash, err := newSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expires = s.now().Add(ttlFor(remember))
	err = q.CreateSession(ctx, dbgen.CreateSessionParams{
		UserID:    userID,
		TokenHash: hash,
		Remember:  remember,
		ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
	})
	return token, expires, err
}

// Lookup은 쿠키 토큰으로 세션을 찾는다. 없거나 만료면 nil.
// refreshed가 true면 로그인 유지 세션의 기한을 늘렸다는 뜻이다 — 호출하는 쪽이 쿠키를 다시 내려준다 (ARCHITECTURE 4절).
func (s *Sessions) Lookup(ctx context.Context, token string) (sess *Session, refreshed bool, err error) {
	hash, ok := hashSessionToken(token)
	if !ok {
		return nil, false, nil
	}
	now := s.now()
	key := sessionCacheKey(hash)

	// 1) 책상 위 쪽지(Valkey)부터. 없거나 Valkey가 안 되면 장부(Postgres)로 넘어간다.
	if raw, err := s.rdb.Get(ctx, key).Bytes(); err == nil {
		var c cachedSession
		if json.Unmarshal(raw, &c) == nil && c.ExpiresAt.After(now) {
			var id pgtype.UUID
			if id.Scan(c.UserID) == nil {
				return &Session{UserID: id, TokenHash: hash, ExpiresAt: c.ExpiresAt}, false, nil
			}
		}
	} else if !errors.Is(err, redis.Nil) {
		s.logger.Warn("세션 캐시를 읽지 못해 DB에서 찾는다", "err", err)
	}

	// 2) 장부. 만료된 세션은 쿼리가 걸러 낸다.
	row, err := s.q.GetSession(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	expires := row.ExpiresAt.Time

	// 마지막 사용 뒤 일정 시간이 지났을 때만 기한을 늘린다 (DATA_MODEL 3.2절).
	touchAfter := sessionTouch
	if row.Remember {
		touchAfter = rememberTouch
	}
	if now.Sub(row.LastSeenAt.Time) >= touchAfter {
		expires = now.Add(ttlFor(row.Remember))
		if err := s.q.TouchSession(ctx, dbgen.TouchSessionParams{
			TokenHash: hash,
			ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
		}); err != nil {
			return nil, false, err
		}
		refreshed = row.Remember
	}

	sess = &Session{UserID: row.UserID, TokenHash: hash, ExpiresAt: expires}
	s.cache(ctx, sess)
	return sess, refreshed, nil
}

// cache는 세션을 Valkey에 쪽지로 남긴다. TTL은 10분이되 세션 만료가 더 이르면 그때까지 (DATA_MODEL 4.2절).
func (s *Sessions) cache(ctx context.Context, sess *Session) {
	ttl := min(cacheTTL, time.Until(sess.ExpiresAt))
	if ttl <= 0 {
		return
	}
	raw, _ := json.Marshal(cachedSession{UserID: sess.UserID.String(), ExpiresAt: sess.ExpiresAt})
	if err := s.rdb.Set(ctx, sessionCacheKey(sess.TokenHash), raw, ttl).Err(); err != nil {
		s.logger.Warn("세션 캐시를 쓰지 못했다", "err", err)
	}
}

// forget은 지운 세션들의 캐시를 지운다. 안 지우면 캐시 TTL(10분) 동안 옛 세션이 살아 있다 (ARCHITECTURE 4절).
func (s *Sessions) forget(ctx context.Context, hashes [][]byte) {
	if len(hashes) == 0 {
		return
	}
	keys := make([]string, len(hashes))
	for i, h := range hashes {
		keys[i] = sessionCacheKey(h)
	}
	if err := s.rdb.Del(ctx, keys...).Err(); err != nil {
		// 감수할 점: Valkey가 잠깐 죽은 사이라면 돌아왔을 때 옛 캐시가 최대 10분 남을 수 있다 (ARCHITECTURE 4절).
		s.logger.Error("지운 세션의 캐시를 지우지 못했다", "count", len(keys), "err", err)
	}
}

func ttlFor(remember bool) time.Duration {
	if remember {
		return rememberTTL
	}
	return sessionTTL
}

// Cookie는 세션 쿠키를 만든다.
// 로그인 유지면 Max-Age를 줘서 브라우저를 닫아도 남고, 아니면 Max-Age 없는 세션 쿠키라 브라우저를 닫으면 사라진다 (REQUIREMENTS A1).
func Cookie(token string, expires time.Time, remember bool) *http.Cookie {
	c := &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true, // 자바스크립트가 못 읽는다 — 화면에 XSS가 생겨도 토큰을 빼 가지 못하게
		Secure:   true, // HTTPS에서만 보낸다. 브라우저는 localhost를 안전한 출처로 쳐서 개발 중에도 된다
		SameSite: http.SameSiteLaxMode,
	}
	if remember {
		c.MaxAge = int(time.Until(expires).Seconds())
	}
	return c
}

// ClearCookie는 브라우저의 세션 쿠키를 지우게 하는 쿠키다.
func ClearCookie() *http.Cookie {
	return &http.Cookie{Name: CookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode}
}
