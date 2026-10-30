package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sagming40/cbt-journal/internal/cache"
	"github.com/sagming40/cbt-journal/internal/config"
	"github.com/sagming40/cbt-journal/internal/db"
	"github.com/sagming40/cbt-journal/internal/server"
)

// serve는 API 서버를 띄운다. 추가 인자는 받지 않는다(_).
func serve(ctx context.Context, cfg config.Config, logger *slog.Logger, _ []string) error {
	// 시작할 때 Postgres · Valkey에 실제로 붙어 본다. 안 되면 바로 멈춘다:
	// 반쯤 뜬 서버가 요청마다 실패하는 것보다 시작할 때 이유를 보여 주는 게 낫다.
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres 연결: %w", err)
	}
	defer pool.Close()

	rdb, err := cache.Open(ctx, cfg.ValkeyURL)
	if err != nil {
		return fmt.Errorf("valkey 연결: %w", err)
	}
	defer rdb.Close()

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: server.New(logger, map[string]server.Checker{
			// 메서드 값: pool.Ping 자체가 "ctx를 받아 error를 돌려주는 함수"라 그대로 꽂힌다.
			"postgres": pool.Ping,
			// go-redis의 Ping은 결과 상자를 돌려줘서, 에러만 꺼내는 작은 어댑터를 씌운다.
			"valkey": func(ctx context.Context) error { return rdb.Ping(ctx).Err() },
		}),
		// 헤더를 아주 느리게 보내 연결을 붙잡아 두는 공격(Slowloris)을 막는다.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ListenAndServe는 서버가 끝날 때까지 돌아오지 않는다 → 고루틴으로 따로 돌리고 결과는 채널로 받는다.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("서버 시작", "addr", cfg.HTTPAddr, "env", cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		// 포트가 이미 쓰이는 중 같은 시작 실패
		return fmt.Errorf("서버: %w", err)
	case <-ctx.Done():
	}

	// 우아한 종료: 새 요청은 안 받고, 처리 중인 요청은 10초까지 기다린다.
	// 비유: 가게 문에 "영업 끝" 팻말을 걸고, 안에 있는 손님 계산은 마저 해 준다.
	// (이 함수가 끝난 뒤 defer로 Valkey → Postgres 순서로 닫힌다. defer는 나중에 건 것부터 실행된다.)
	logger.Info("종료 신호를 받아 정리 중")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("종료: %w", err)
	}
	logger.Info("서버 종료")
	return nil
}
