package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	// 밑줄 import: 이름은 안 쓰고 "pgx"라는 database/sql 드라이버를 등록만 한다.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/sagming40/cbt-journal/internal/config"
	"github.com/sagming40/cbt-journal/internal/db"
)

// migrate는 app migrate up|down|status를 처리한다 (ARCHITECTURE 1절).
func migrate(ctx context.Context, cfg config.Config, logger *slog.Logger, args []string) error {
	if len(args) != 1 {
		return errors.New("사용법: app migrate up|down|status")
	}

	// goose는 표준 database/sql 연결을 받는다. pgx가 그 어댑터(stdlib)를 같이 준다.
	sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres 연결: %w", err)
	}
	defer sqlDB.Close()

	// embed한 파일들은 "migrations/00001_init.sql"처럼 폴더 이름이 붙어 있다.
	// goose는 맨 위에서 파일을 찾으므로 migrations 폴더 안을 맨 위로 잘라 넘긴다.
	files, err := fs.Sub(db.Migrations, "migrations")
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, files)
	if err != nil {
		return fmt.Errorf("goose 준비: %w", err)
	}

	switch args[0] {
	case "up":
		results, err := provider.Up(ctx)
		if err != nil {
			return fmt.Errorf("up: %w", err)
		}
		if len(results) == 0 {
			logger.Info("적용할 마이그레이션 없음")
		}
		for _, r := range results {
			logger.Info("적용", "version", r.Source.Version, "file", r.Source.Path, "took", r.Duration)
		}
	case "down":
		// 한 번에 하나만 되돌린다. 실수로 여러 개를 날리지 않게.
		r, err := provider.Down(ctx)
		if err != nil {
			return fmt.Errorf("down: %w", err)
		}
		logger.Info("되돌림", "version", r.Source.Version, "file", r.Source.Path)
	case "status":
		statuses, err := provider.Status(ctx)
		if err != nil {
			return fmt.Errorf("status: %w", err)
		}
		for _, s := range statuses {
			logger.Info("상태", "version", s.Source.Version, "file", s.Source.Path, "state", s.State)
		}
	default:
		return fmt.Errorf("모르는 migrate 명령: %q (up · down · status)", args[0])
	}
	return nil
}
