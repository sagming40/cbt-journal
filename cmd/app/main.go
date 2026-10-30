// app은 CBT 저널의 실행 파일 하나다. 서브커맨드로 역할을 고른다 (ARCHITECTURE 1절).
// 비유: 맥가이버 칼 — 칼은 하나, 꺼내는 날(serve · migrate …)만 다르다.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sagming40/cbt-journal/internal/config"
	"github.com/sagming40/cbt-journal/internal/server"
)

const usage = `사용법: app <명령>

명령:
  serve    API 서버를 띄운다
`

// commands는 서브커맨드 이름 → 실행할 함수 표다.
// 다음 단계에서 "migrate"를 한 줄 더한다.
var commands = map[string]func(cfg config.Config, logger *slog.Logger) error{
	"serve": serve,
}

func main() {
	// os.Exit은 defer를 건너뛴다. 그래서 정리 작업이 있는 코드는 run 안에 두고,
	// main은 결과만 보고 종료 코드를 정한다.
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "오류:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// 명령부터 확인한다. 오타 난 명령에 "설정이 없다"부터 보여 주면 헷갈린다.
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		return errors.New("명령이 없다")
	}
	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("모르는 명령: %q", args[0])
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("설정을 읽지 못했다:\n%w", err)
	}
	return cmd(cfg, newLogger(cfg))
}

// newLogger: 개발 중엔 사람이 읽기 쉬운 글자, 운영에선 기계가 모으기 쉬운 JSON.
// 로그에는 id · 경로 · 상태 코드 · 걸린 시간만 남긴다 (REQUIREMENTS 5.1절).
func newLogger(cfg config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var h slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if cfg.IsProduction() {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

func serve(cfg config.Config, logger *slog.Logger) error {
	// Ctrl+C(Interrupt)나 종료 신호(SIGTERM)가 오면 ctx가 "취소됨" 상태가 된다.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: server.New(logger),
		// 헤더를 아주 느리게 보내 연결을 붙잡아 두는 공격(Slowloris)을 막는다.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ListenAndServe는 서버가 끝날 때까지 돌아오지 않는다 → 고루틴(가벼운 동시 실행)으로 따로 돌리고,
	// 결과는 채널(고루틴끼리 값을 건네는 관)로 받는다.
	errCh := make(chan error, 1)
	go func() {
		logger.Info("서버 시작", "addr", cfg.HTTPAddr, "env", cfg.Env)
		errCh <- srv.ListenAndServe()
	}()

	// select: 여러 채널 중 먼저 오는 쪽을 처리한다.
	select {
	case err := <-errCh:
		// 포트가 이미 쓰이는 중 같은 시작 실패
		return fmt.Errorf("서버: %w", err)
	case <-ctx.Done():
	}

	// 우아한 종료: 새 요청은 안 받고, 처리 중인 요청은 10초까지 기다린다.
	// 비유: 가게 문에 "영업 끝" 팻말을 걸고, 안에 있는 손님 계산은 마저 해 준다.
	logger.Info("종료 신호를 받아 정리 중")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("종료: %w", err)
	}
	logger.Info("서버 종료")
	return nil
}
