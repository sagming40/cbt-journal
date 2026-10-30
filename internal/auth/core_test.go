package auth

import (
	"context"
	"crypto/sha1" //nolint:gosec // 테스트에서 Pwned Passwords 응답을 만들려고 쓴다
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("올바른 말 배터리 스테이플")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Errorf("PHC 형식이 아니다: %s", hash)
	}

	ok, err := VerifyPassword("올바른 말 배터리 스테이플", hash)
	if err != nil || !ok {
		t.Errorf("맞는 비밀번호인데 ok=%v err=%v", ok, err)
	}
	ok, err = VerifyPassword("틀린 비밀번호입니다아", hash)
	if err != nil || ok {
		t.Errorf("틀린 비밀번호인데 ok=%v err=%v", ok, err)
	}

	again, _ := HashPassword("올바른 말 배터리 스테이플")
	if again == hash {
		t.Error("같은 비밀번호여도 소금이 달라 해시가 달라야 한다")
	}
	if NeedsRehash(hash) {
		t.Error("지금 파라미터로 만든 해시인데 NeedsRehash가 true")
	}
}

// 없는 이메일 로그인에 쓰는 가짜 해시는 진짜 해시와 같은 파라미터여야 걸리는 시간이 같다 (ARCHITECTURE 4절).
func TestDummyHashMatchesRealCost(t *testing.T) {
	if NeedsRehash(dummyHash()) {
		t.Error("가짜 해시의 파라미터가 진짜와 다르면 시간으로 가입 여부가 드러난다")
	}
	if ok, err := VerifyPassword("아무 비밀번호나 넣어 본다", dummyHash()); err != nil || ok {
		t.Errorf("가짜 해시는 어떤 비밀번호와도 맞으면 안 된다: ok=%v err=%v", ok, err)
	}
}

func TestVerifyPasswordRejectsBadHashes(t *testing.T) {
	bad := []string{
		"",
		"plain",
		"$argon2i$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA",     // argon2i (id가 아님)
		"$argon2id$v=16$m=19456,t=2,p=1$c2FsdA$aGFzaA",    // 옛 버전
		"$argon2id$v=19$m=99999999,t=2,p=1$c2FsdA$aGFzaA", // 터무니없는 메모리
		"$argon2id$v=19$m=19456,t=2,p=1$***$aGFzaA",       // 소금이 base64가 아님
		"$argon2id$v=19$m=19456,t=2,p=1$c2FsdA",           // 칸이 모자람
	}
	for _, h := range bad {
		if _, err := VerifyPassword("아무거나아무거나", h); err == nil {
			t.Errorf("%q: 오류여야 한다", h)
		}
	}
	if !NeedsRehash("$argon2id$v=19$m=4096,t=1,p=1$c2FsdA$aGFzaA") {
		t.Error("약한 파라미터인데 NeedsRehash가 false")
	}
}

func TestNormalizeEmail(t *testing.T) {
	got, err := NormalizeEmail("  Me@Example.COM ")
	if err != nil || got != "me@example.com" {
		t.Errorf("got %q, %v", got, err)
	}
	for _, bad := range []string{"", "not-an-email", "이름 <me@example.com>", strings.Repeat("a", 250) + "@x.co"} {
		if _, err := NormalizeEmail(bad); err == nil {
			t.Errorf("%q: 오류여야 한다", bad)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		ok       bool
	}{
		{"짧다", false},
		{"12345678901", false},  // 11자
		{"123456789012", true},  // 12자
		{"한글열두글자비밀번호입니다", true}, // 글자 수로 센다 (바이트 아님)
		{strings.Repeat("a", 128), true},
		{strings.Repeat("a", 129), false},
	}
	for _, tt := range tests {
		err := ValidatePassword("password", tt.password)
		if (err == nil) != tt.ok {
			t.Errorf("%q: ok=%v인데 err=%v", tt.password, tt.ok, err)
		}
		var ve *ValidationError
		if err != nil && (!errors.As(err, &ve) || ve.Fields["password"] == "") {
			t.Errorf("%q: password 칸 오류여야 한다: %v", tt.password, err)
		}
	}
}

func TestNormalizeNicknameAndTimezone(t *testing.T) {
	if got, err := NormalizeNickname("  상민 "); err != nil || got != "상민" {
		t.Errorf("got %q, %v", got, err)
	}
	for _, bad := range []string{"", "   ", strings.Repeat("가", 31), "탭\t문자"} {
		if _, err := NormalizeNickname(bad); err == nil {
			t.Errorf("%q: 오류여야 한다", bad)
		}
	}
	for _, ok := range []string{"Asia/Seoul", "UTC", "America/New_York"} {
		if err := ValidateTimezone(ok); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "Local", "Mars/Olympus", "../etc/passwd"} {
		if err := ValidateTimezone(bad); err == nil {
			t.Errorf("%q: 오류여야 한다", bad)
		}
	}
}

func TestSessionToken(t *testing.T) {
	token, hash, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	got, ok := hashSessionToken(token)
	if !ok || hex.EncodeToString(got) != hex.EncodeToString(hash) {
		t.Error("쿠키 토큰에서 같은 해시가 나와야 한다")
	}
	if strings.Contains(sessionCacheKey(hash), token) {
		t.Error("캐시 키에 토큰 원문이 들어가면 안 된다")
	}
	for _, bad := range []string{"", "짧음", "!!!!", strings.Repeat("A", 50)} {
		if _, ok := hashSessionToken(bad); ok {
			t.Errorf("%q: 형식이 틀린 토큰인데 ok", bad)
		}
	}
}

func TestPwnedPasswords(t *testing.T) {
	const password = "password123456"
	sum := sha1.Sum([]byte(password)) //nolint:gosec // 위 import 주석과 같은 이유
	full := strings.ToUpper(hex.EncodeToString(sum[:]))

	var gotPath, gotPadding string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotPadding = r.URL.Path, r.Header.Get("Add-Padding")
		// 다른 해시 한 줄, 패딩 줄(0), 찾는 해시
		fmt.Fprintf(w, "0000000000000000000000000000000000A:3\r\n%s:0\r\n%s:42\r\n", strings.Repeat("F", 35), full[5:])
	}))
	defer srv.Close()

	p := &PwnedPasswords{BaseURL: srv.URL + "/range/", Client: srv.Client()}
	breached, err := p.IsBreached(context.Background(), password)
	if err != nil || !breached {
		t.Fatalf("유출 목록에 있는데 breached=%v err=%v", breached, err)
	}
	if gotPath != "/range/"+full[:5] {
		t.Errorf("앞 5글자만 보내야 한다: %s", gotPath)
	}
	if gotPadding != "true" {
		t.Error("Add-Padding 헤더가 없다")
	}

	breached, err = p.IsBreached(context.Background(), "목록에 없는 아주 긴 비밀번호")
	if err != nil || breached {
		t.Errorf("목록에 없는데 breached=%v err=%v", breached, err)
	}
}

func TestPwnedPasswordsPaddingAndErrors(t *testing.T) {
	const password = "padding-only-password"
	sum := sha1.Sum([]byte(password)) //nolint:gosec // 위 import 주석과 같은 이유
	suffix := strings.ToUpper(hex.EncodeToString(sum[:]))[5:]

	padding := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "%s:0\n", suffix) // 패딩 줄은 유출이 아니다
	}))
	defer padding.Close()
	p := &PwnedPasswords{BaseURL: padding.URL + "/", Client: padding.Client()}
	if breached, err := p.IsBreached(context.Background(), password); err != nil || breached {
		t.Errorf("패딩 줄인데 breached=%v err=%v", breached, err)
	}

	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer broken.Close()
	p = &PwnedPasswords{BaseURL: broken.URL + "/", Client: broken.Client()}
	if _, err := p.IsBreached(context.Background(), password); err == nil {
		t.Error("500이면 오류여야 한다 (건너뛸지는 호출하는 쪽이 정한다)")
	}
}
