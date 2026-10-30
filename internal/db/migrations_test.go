package db_test

import (
	"context"
	"testing"

	"github.com/sagming40/cbt-journal/internal/db"
	"github.com/sagming40/cbt-journal/internal/testutil"
)

// 마이그레이션을 끝까지 되돌렸다 다시 올려도 되는지 본다.
// Down이 깨져 있으면 실수했을 때 되돌릴 길이 없다.
func TestMigrationsDownAndUpAgain(t *testing.T) {
	env := testutil.Start(t)
	ctx := context.Background()

	provider, sqlDB, err := db.OpenMigrator(env.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatalf("끝까지 down: %v", err)
	}
	results, err := provider.Up(ctx)
	if err != nil {
		t.Fatalf("다시 up: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("다시 up했는데 적용된 마이그레이션이 없다")
	}
}

// DB가 앱 검증과 별개로 마지막 방어선 역할을 하는지 본다 (DATA_MODEL 1절 · 3.1절).
func TestUsersConstraints(t *testing.T) {
	env := testutil.Start(t)
	pool := env.Pool(t)
	testutil.Reset(t, pool, env.Valkey(t))
	ctx := context.Background()

	const hash = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
	tests := []struct {
		name     string
		email    string
		hash     string
		nickname string
	}{
		{"대문자 이메일", "Me@Example.com", hash, "나"},
		{"argon2id가 아닌 해시", "me@example.com", "plain-text", "나"},
		{"빈 닉네임", "me@example.com", hash, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.Exec(ctx,
				"INSERT INTO users (email, password_hash, nickname) VALUES ($1, $2, $3)",
				tt.email, tt.hash, tt.nickname)
			if err == nil {
				t.Fatal("CHECK 제약이 막아야 한다")
			}
		})
	}

	t.Run("사용자를 지우면 세션도 지워진다", func(t *testing.T) {
		var userID string
		err := pool.QueryRow(ctx,
			"INSERT INTO users (email, password_hash, nickname) VALUES ('a@example.com', $1, '나') RETURNING id::text",
			hash).Scan(&userID)
		if err != nil {
			t.Fatal(err)
		}
		_, err = pool.Exec(ctx,
			"INSERT INTO sessions (user_id, token_hash, remember, expires_at) VALUES ($1, $2, false, now() + interval '1 hour')",
			userID, make([]byte, 32))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID); err != nil {
			t.Fatal(err)
		}
		var n int
		if err := pool.QueryRow(ctx, "SELECT count(*) FROM sessions").Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("세션 %d개가 남았다 (CASCADE)", n)
		}
	})
}
