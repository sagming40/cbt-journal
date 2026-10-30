// Package auth는 로그인 · 세션 · 요청 제한 · 비밀번호 규칙과 인증 API handler를 담는다 (ARCHITECTURE 2절 · 4절).
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/argon2"
)

// argon2id 파라미터: OWASP 권장 최소값(메모리 19MiB, 반복 2, 병렬 1).
// 일부러 느리고 메모리를 먹게 만들어, 해시가 새도 비밀번호를 하나씩 맞혀 보는 비용을 키운다.
// 비유: 금고 다이얼을 한 번 돌릴 때마다 무거운 손잡이를 같이 돌려야 하게 만든다.
const (
	argonMemoryKiB = 19 * 1024
	argonTime      = 2
	argonThreads   = 1
	argonKeyLen    = 32
	argonSaltLen   = 16

	// DB에서 읽은 해시의 파라미터가 터무니없이 크면 거부한다(검증 한 번에 메모리를 수 GB 쓰지 않게).
	maxArgonMemoryKiB = 1024 * 1024
	maxArgonTime      = 10
)

var errBadHash = errors.New("argon2id 해시 형식이 아니다")

// HashPassword는 비밀번호를 argon2id PHC 문자열로 만든다: $argon2id$v=19$m=…,t=…,p=…$소금$해시
// 파라미터가 문자열 안에 같이 들어 있어서, 나중에 파라미터를 올려도 옛 해시를 그대로 검증할 수 있다 (DATA_MODEL 3.1절).
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemoryKiB, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemoryKiB, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
	salt    []byte
	key     []byte
}

func parseHash(encoded string) (argonParams, error) {
	var p argonParams
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return p, errBadHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return p, errBadHash
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads); err != nil {
		return p, errBadHash
	}
	if p.memory == 0 || p.memory > maxArgonMemoryKiB || p.time == 0 || p.time > maxArgonTime || p.threads == 0 {
		return p, errBadHash
	}
	var err error
	if p.salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil || len(p.salt) == 0 {
		return p, errBadHash
	}
	if p.key, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil || len(p.key) == 0 {
		return p, errBadHash
	}
	return p, nil
}

// VerifyPassword는 비밀번호가 해시와 맞는지 본다. 형식이 틀린 해시는 오류다.
func VerifyPassword(password, encoded string) (bool, error) {
	p, err := parseHash(encoded)
	if err != nil {
		return false, err
	}
	key := argon2.IDKey([]byte(password), p.salt, p.time, p.memory, p.threads, uint32(len(p.key))) //nolint:gosec // 키 길이는 parseHash가 디코드한 값이라 작다
	// 비교에 걸리는 시간이 "몇 글자째에서 달랐는지"를 드러내지 않게 상수 시간으로 비교한다.
	return subtle.ConstantTimeCompare(key, p.key) == 1, nil
}

// NeedsRehash는 해시가 지금 파라미터보다 약하게 만들어졌는지 알려 준다.
// 로그인에 성공했을 때 true면 새 파라미터로 다시 해시해 저장한다 — 사용자는 모르게 조금씩 올라간다.
func NeedsRehash(encoded string) bool {
	p, err := parseHash(encoded)
	if err != nil {
		return true
	}
	return p.memory != argonMemoryKiB || p.time != argonTime || p.threads != argonThreads || len(p.key) != argonKeyLen
}

// dummyHash는 없는 이메일로 로그인할 때 검증할 가짜 해시다 (ARCHITECTURE 4절).
// 없는 계정이면 바로 실패를 돌려주는 대신 같은 무게의 계산을 한 번 해서, 걸리는 시간으로 가입 여부가 드러나지 않게 한다.
var dummyHash = sync.OnceValue(func() string {
	h, err := HashPassword("시간을 맞추기 위한 가짜 비밀번호")
	if err != nil {
		panic(err) // crypto/rand가 실패하는 환경이면 어차피 아무것도 안전하게 할 수 없다
	}
	return h
})
