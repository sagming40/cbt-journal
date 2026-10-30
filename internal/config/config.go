// Package config는 환경변수에서 앱 설정을 읽고 검증한다
// 빠지거나 잘못된 값이 있으면 시작하지 않는다 (ARCHITECTURE 2절)
// 비유 ─ 출발 전 체크리스트 → 한 칸이라도 비면 비행기는 이륙하지 않는다
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config는 앱이 시작할 때 한 번 읽어 두는 설정 묶음이다.
type Config struct {
	Env         string // development | production
	HTTPAddr    string
	DatabaseURL string
	ValkeyURL   string
	LogLevel    slog.Level
}

// IsProduction은 운영 환경인지 알려 준다. 로그 형식 등을 고를 때 쓴다.
func (c Config) IsProduction() bool {
	return c.Env == "production"
}

// Load는 진짜 환경변수(os.Getenv)로 설정을 읽는다.
func Load() (Config, error) {
	return load(os.Getenv)
}

// load는 "키를 주면 값을 돌려주는 함수"를 받는다.
// 진짜 환경변수 대신 map을 꽂을 수 있어서 테스트가 서로 섞이지 않는다.
// 비유: 콘센트에 진짜 전기 대신 시험용 전원을 꽂는 것.
func load(getenv func(string) string) (Config, error) {
	var errs []error

	// 필수 값: 비어 있으면 바로 멈추지 않고 모아 뒀다가 한 번에 알려 준다.
	required := func(key string) string {
		v := strings.TrimSpace(getenv(key))
		if v == "" {
			errs = append(errs, fmt.Errorf("%s: 값이 없다", key))
		}
		return v
	}
	// 선택 값: 비어 있으면 기본값을 쓴다.
	optional := func(key, def string) string {
		if v := strings.TrimSpace(getenv(key)); v != "" {
			return v
		}
		return def
	}

	cfg := Config{
		Env: required("APP_ENV"),
		// 기본은 내 PC에서만 열리는 주소. 운영에서는 환경변수로 따로 정한다.
		HTTPAddr:    optional("HTTP_ADDR", "127.0.0.1:8080"),
		DatabaseURL: required("DATABASE_URL"),
		ValkeyURL:   required("VALKEY_URL"),
	}

	// 오타("prod")를 조용히 개발 환경으로 취급하면 운영에서 사고가 난다 → 정한 두 값만 받는다.
	if cfg.Env != "" && cfg.Env != "development" && cfg.Env != "production" {
		errs = append(errs, fmt.Errorf("APP_ENV: development 또는 production이어야 한다 (지금: %q)", cfg.Env))
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(optional("LOG_LEVEL", "info"))); err != nil {
		errs = append(errs, errors.New("LOG_LEVEL: debug · info · warn · error 중 하나여야 한다"))
	}

	// 주의: 에러 문구에 DATABASE_URL 같은 값을 넣지 않는다. 비밀번호가 들어 있다 (CLAUDE.md "데이터 취급").

	// errors.Join은 여러 에러를 하나로 묶는다. errs가 비어 있으면 nil을 돌려준다.
	if err := errors.Join(errs...); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
