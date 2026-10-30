// Package testutil은 통합 테스트용 진짜 Postgres · Valkey를 띄운다 (ARCHITECTURE 8절).
// testcontainers가 Docker로 컨테이너를 띄우고, 테스트 프로그램이 끝나면 Ryuk 컨테이너가 알아서 치운다.
// 비유: 시험 볼 때마다 새 실험실을 빌렸다가 반납한다 — 내 PC의 compose DB를 건드리지 않는다.
//
// _test.go 파일에서만 import한다. 실행 파일에는 들어가지 않는다.
package testutil

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"

	"github.com/sagming40/cbt-journal/internal/db"
)

// compose.yaml과 같은 이미지 · 같은 큰 버전을 쓴다.
const (
	postgresImage = "postgres:18"
	valkeyImage   = "valkey/valkey:9"
)

// Env는 테스트 프로그램(패키지) 하나가 같이 쓰는 컨테이너 접속 정보다.
type Env struct {
	DatabaseURL string
	ValkeyURL   string
}

var (
	once   sync.Once
	shared *Env
	setErr error
)

// Start는 컨테이너를 (패키지당 한 번만) 띄우고 마이그레이션을 끝까지 적용한다.
// Docker가 없으면 로컬에서는 건너뛰고, CI(환경변수 CI가 있음)에서는 실패한다 — CI에서 조용히 건너뛰면 테스트가 없는 것과 같다.
func Start(t testing.TB) *Env {
	t.Helper()
	once.Do(func() { shared, setErr = start() })
	if setErr != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("통합 테스트 컨테이너를 띄우지 못했다: %v", setErr)
		}
		t.Skipf("Docker가 없어 통합 테스트를 건너뛴다 (Docker Desktop을 켜면 돈다): %v", setErr)
	}
	return shared
}

func start() (*Env, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pg, err := tcpostgres.Run(ctx, postgresImage,
		tcpostgres.WithDatabase("cbt_test"),
		tcpostgres.WithUsername("cbt"),
		tcpostgres.WithPassword("test-only-password"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, err
	}
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	vk, err := tcvalkey.Run(ctx, valkeyImage)
	if err != nil {
		return nil, err
	}
	vkURL, err := vk.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	provider, sqlDB, err := db.OpenMigrator(dsn)
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()
	if _, err := provider.Up(ctx); err != nil {
		return nil, err
	}

	return &Env{DatabaseURL: dsn, ValkeyURL: vkURL}, nil
}

// Pool은 테스트용 Postgres 연결 풀이다. 테스트가 끝나면 닫힌다.
func (e *Env) Pool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), e.DatabaseURL)
	if err != nil {
		t.Fatalf("postgres 풀: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// Valkey는 테스트용 Valkey 클라이언트다. 테스트가 끝나면 닫힌다.
func (e *Env) Valkey(t testing.TB) *redis.Client {
	t.Helper()
	opts, err := redis.ParseURL(e.ValkeyURL)
	if err != nil {
		t.Fatalf("valkey URL: %v", err)
	}
	rdb := redis.NewClient(opts)
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

// Reset은 테스트끼리 데이터가 섞이지 않게 모든 사용자 데이터와 Valkey 키를 지운다.
// 테스트는 패키지 안에서 차례로 돈다(t.Parallel을 쓰지 않는다) — 같은 DB를 나눠 쓰기 때문이다.
func Reset(t testing.TB, pool *pgxpool.Pool, rdb *redis.Client) {
	t.Helper()
	ctx := context.Background()
	// users를 지우면 CASCADE로 사용자 데이터가 전부 지워진다 (DATA_MODEL 1절).
	if _, err := pool.Exec(ctx, "TRUNCATE users CASCADE"); err != nil {
		t.Fatalf("TRUNCATE: %v", err)
	}
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("FLUSHDB: %v", err)
	}
}
