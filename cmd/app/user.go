package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/sagming40/cbt-journal/internal/auth"
	"github.com/sagming40/cbt-journal/internal/cache"
	"github.com/sagming40/cbt-journal/internal/config"
	"github.com/sagming40/cbt-journal/internal/db"
)

const userUsage = `사용법:
  app user create --email <이메일> --nickname <닉네임>
  app user reset-password --email <이메일>

비밀번호는 명령줄 인자로 받지 않고 실행 중에 묻는다(셸 기록에 남지 않게).
`

// user는 가입이 막힌 동안 계정을 만들고 비밀번호를 재설정하는 관리 명령이다 (REQUIREMENTS A9, ARCHITECTURE 1절).
// 가입 API와 같은 입력 규칙 · 유출 비밀번호 확인을 거친다 (A10).
func user(ctx context.Context, cfg config.Config, logger *slog.Logger, args []string) error {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, userUsage)
		return errors.New("user 명령이 없다")
	}
	sub, rest := args[0], args[1:]

	fs := flag.NewFlagSet("user "+sub, flag.ContinueOnError)
	email := fs.String("email", "", "이메일")
	nickname := fs.String("nickname", "", "닉네임 (create)")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *email == "" {
		fmt.Fprint(os.Stderr, userUsage)
		return errors.New("--email이 필요하다")
	}

	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres 연결: %w", err)
	}
	defer pool.Close()
	// 비밀번호를 재설정하면 그 사용자의 세션 캐시도 지워야 해서 Valkey에도 붙는다 (ARCHITECTURE 4절).
	rdb, err := cache.Open(ctx, cfg.ValkeyURL)
	if err != nil {
		return fmt.Errorf("valkey 연결: %w", err)
	}
	defer rdb.Close()

	svc := auth.NewService(auth.Options{Pool: pool, Valkey: rdb, Breach: auth.NewPwnedPasswords(), Logger: logger})

	switch sub {
	case "create":
		if *nickname == "" {
			fmt.Fprint(os.Stderr, userUsage)
			return errors.New("--nickname이 필요하다")
		}
		password, err := askPassword()
		if err != nil {
			return err
		}
		u, err := svc.CreateUser(ctx, *email, password, *nickname)
		if err != nil {
			return explain(err)
		}
		logger.Info("계정을 만들었다", "user_id", u.ID.String())
	case "reset-password":
		password, err := askPassword()
		if err != nil {
			return err
		}
		n, err := svc.ResetPassword(ctx, *email, password)
		if err != nil {
			return explain(err)
		}
		logger.Info("비밀번호를 바꾸고 세션을 모두 끊었다", "sessions", n)
	default:
		fmt.Fprint(os.Stderr, userUsage)
		return fmt.Errorf("모르는 user 명령: %q", sub)
	}
	return nil
}

// askPassword는 비밀번호를 두 번 묻는다. 터미널이면 화면에 보이지 않게 받고,
// 파이프로 들어오면(스크립트) 한 줄씩 읽는다.
func askPassword() (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) { //nolint:gosec // 파일 기술자는 작은 수라 int로 바꿔도 넘치지 않는다
		first, err := readHidden("새 비밀번호: ")
		if err != nil {
			return "", err
		}
		second, err := readHidden("한 번 더: ")
		if err != nil {
			return "", err
		}
		if first != second {
			return "", errors.New("두 비밀번호가 다르다")
		}
		return first, nil
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func readHidden(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd())) //nolint:gosec // 위와 같은 이유
	fmt.Fprintln(os.Stderr)
	return string(b), err
}

// explain은 입력 오류를 칸별로 풀어 보여 준다.
func explain(err error) error {
	var ve *auth.ValidationError
	if errors.As(err, &ve) {
		return ve
	}
	return err
}
