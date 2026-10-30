// Package cache는 Valkey 연결과 (나중에) 키 이름 함수를 담는다 (ARCHITECTURE 2절).
// Valkey는 Redis와 프로토콜이 같아서 go-redis 클라이언트를 그대로 쓴다 (decisions 0005).
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Open은 Valkey 클라이언트를 만들고 PING으로 실제로 붙는지 확인한다.
func Open(ctx context.Context, url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("VALKEY_URL 형식: %w", err)
	}
	rdb := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// go-redis는 결과를 상자(*StatusCmd)에 담아 돌려준다. .Err()로 에러만 꺼낸다.
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	return rdb, nil
}
