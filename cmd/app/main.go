// app은 CBT 저널의 실행 파일 하나다. 서브커맨드로 역할을 고른다 (ARCHITECTURE 1절).
// 비유: 맥가이버 칼 — 칼은 하나, 꺼내는 날(serve · migrate …)만 다르다.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sagming40/cbt-journal/internal/config"
)

const usage = `사용법: app <명령>

명령:
  serve                    API 서버를 띄운다
  migrate up|down|status   DB 마이그레이션 (goose)
`

// command는 서브커맨드 하나가 하는 일이다. args는 명령 이름 뒤에 붙은 나머지 인자다.
type command func(ctx context.Context, cfg config.Config, logger *slog.Logger, args []string) error

// commands는 서브커맨드 이름 → 실행할 함수 표다. 각 함수는 같은 이름의 파일에 있다.
var commands = map[string]command{
	"serve":   serve,
	"migrate": migrate,
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

	// Ctrl+C(Interrupt) · SIGTERM을 모든 명령이 같이 쓴다:
	// 서버는 우아하게 끄고, 마이그레이션은 하던 일을 멈춘다.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return cmd(ctx, cfg, newLogger(cfg), args[1:])
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
