package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// 통계 캐시 버전 키의 수명 (DATA_MODEL 4.2절)
const statsVerTTL = 30 * 24 * time.Hour

// BumpStatsVersion은 통계 캐시 버전을 올린다 — 옛 집계 키는 아무도 안 읽다가 TTL로 사라진다 (DATA_MODEL 4.4절).
func BumpStatsVersion(ctx context.Context, rdb *redis.Client, userID string) error {
	key := "stats:ver:" + userID
	pipe := rdb.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, statsVerTTL)
	_, err := pipe.Exec(ctx)
	return err
}
