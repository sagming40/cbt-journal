package config

import (
	"log/slog"
	"strings"
	"testing"
)

// fakeEnv는 map을 환경변수처럼 쓰게 해 준다.
func fakeEnv(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

// validEnv는 필수 값이 다 들어 있는 기본 환경이다. 테스트마다 한 칸씩 망가뜨려 쓴다.
func validEnv() map[string]string {
	return map[string]string{
		"APP_ENV":      "development",
		"DATABASE_URL": "postgres://u:secret-pw@localhost:5436/db",
		"VALKEY_URL":   "redis://localhost:6380/0",
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := load(fakeEnv(validEnv()))
	if err != nil {
		t.Fatalf("에러가 없어야 한다: %v", err)
	}
	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Errorf("HTTPAddr = %q, 기본값 127.0.0.1:8080이어야 한다", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, 기본값 info여야 한다", cfg.LogLevel)
	}
	if cfg.IsProduction() {
		t.Error("development인데 IsProduction이 true다")
	}
}

func TestLoadReportsAllMissingAtOnce(t *testing.T) {
	_, err := load(fakeEnv(map[string]string{}))
	if err == nil {
		t.Fatal("빈 환경인데 에러가 없다")
	}
	for _, key := range []string{"APP_ENV", "DATABASE_URL", "VALKEY_URL"} {
		if !strings.Contains(err.Error(), key) {
			t.Errorf("에러에 %s가 없다: %v", key, err)
		}
	}
}

func TestLoadRejectsBadValues(t *testing.T) {
	// 표 주도 테스트: 경우를 표로 적고 같은 검사를 돌린다. Go에서 가장 흔한 테스트 모양이다.
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"APP_ENV 오타", "APP_ENV", "prod"},
		{"LOG_LEVEL 오타", "LOG_LEVEL", "verbose"},
		{"공백만 있는 DATABASE_URL", "DATABASE_URL", "   "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := validEnv()
			env[tt.key] = tt.value
			_, err := load(fakeEnv(env))
			if err == nil || !strings.Contains(err.Error(), tt.key) {
				t.Errorf("%s 에러를 기대했다, 받은 것: %v", tt.key, err)
			}
		})
	}
}

func TestLoadErrorHidesSecrets(t *testing.T) {
	env := validEnv()
	env["APP_ENV"] = "" // 다른 칸 하나를 비워 에러를 일으킨다
	_, err := load(fakeEnv(env))
	if err == nil {
		t.Fatal("에러를 기대했다")
	}
	if strings.Contains(err.Error(), "secret-pw") {
		t.Errorf("에러 문구에 비밀번호가 들어 있다: %v", err)
	}
}
