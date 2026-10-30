package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/netip"
	"time"

	"github.com/redis/go-redis/v9"
)

// 요청 제한 (ARCHITECTURE 4절 표, DATA_MODEL 4.2절).
type limitRule struct {
	limit  int64
	window time.Duration
}

var (
	loginIPRule   = limitRule{20, 15 * time.Minute}
	loginAcctRule = limitRule{5, 15 * time.Minute}
	signupIPRule  = limitRule{5, time.Hour}
	pwChangeRule  = limitRule{5, 15 * time.Minute}
)

func loginIPKey(ip netip.Addr) string  { return "rl:login:ip:" + ipString(ip) }
func signupIPKey(ip netip.Addr) string { return "rl:signup:ip:" + ipString(ip) }
func pwChangeKey(userID string) string { return "rl:pwchange:user:" + userID }

// loginAcctKey는 이메일을 해시해서 키에 넣는다. Valkey 키 이름에 개인정보를 넣지 않는다 (DATA_MODEL 4.1절).
func loginAcctKey(email string) string {
	sum := sha256.Sum256([]byte(email))
	return "rl:login:acct:" + hex.EncodeToString(sum[:])
}

func ipString(ip netip.Addr) string {
	if !ip.IsValid() {
		return "unknown"
	}
	return ip.String()
}

// limiter는 Valkey 카운터로 횟수를 센다.
// Valkey가 응답하지 않으면 막지 않고 통과시킨다(로그만) — Postgres가 진실의 원천이고 Valkey 없이도 로그인은 돼야 한다 (ARCHITECTURE 4절).
type limiter struct {
	rdb    *redis.Client
	logger *slog.Logger
}

// blocked는 이미 한도에 닿았는지 본다. 닿았으면 키가 사라질 때까지 남은 시간을 같이 준다.
func (l *limiter) blocked(ctx context.Context, key string, rule limitRule) (bool, time.Duration) {
	n, err := l.rdb.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		return false, 0
	}
	if err != nil {
		l.logger.Warn("요청 제한 카운터를 읽지 못해 통과시킨다", "err", err)
		return false, 0
	}
	if n < rule.limit {
		return false, 0
	}
	ttl, err := l.rdb.TTL(ctx, key).Result()
	if err != nil || ttl <= 0 {
		ttl = rule.window
	}
	return true, ttl
}

// hit은 한 번 센다. TTL은 처음 셀 때만 건다 — 셀 때마다 다시 걸면 계속 시도하는 동안 창이 끝나지 않는다 (DATA_MODEL 4.1절).
func (l *limiter) hit(ctx context.Context, key string, rule limitRule) {
	pipe := l.rdb.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, rule.window)
	if _, err := pipe.Exec(ctx); err != nil {
		l.logger.Warn("요청 제한 카운터를 올리지 못했다", "err", err)
	}
}

func (l *limiter) reset(ctx context.Context, key string) {
	if err := l.rdb.Del(ctx, key).Err(); err != nil {
		l.logger.Warn("요청 제한 카운터를 지우지 못했다", "err", err)
	}
}
