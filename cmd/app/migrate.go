package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sagming40/cbt-journal/internal/config"
	"github.com/sagming40/cbt-journal/internal/db"
)

// migrate는 app migrate up|down|status를 처리한다 (ARCHITECTURE 1절).
func migrate(ctx context.Context, cfg config.Config, logger *slog.Logger, args []string) error {
	if len(args) != 1 {
		return errors.New("사용법: app migrate up|down|status")
	}

	provider, sqlDB, err := db.OpenMigrator(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

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
