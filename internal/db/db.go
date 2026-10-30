package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open은 Postgres 연결 풀을 만들고 실제로 한 번 붙어 본다.
// pgxpool.New는 연결을 미리 만들지 않아서 접속 정보가 틀려도 성공한다 → Ping으로 바로 확인한다.
// 비유: 택시 회사(풀) 번호만 저장해 두지 않고, 한 대 불러서 오는지 본다.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	// 컨테이너가 꺼져 있을 때 한없이 기다리지 않게 5초로 자른다.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
