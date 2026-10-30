package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// 세션 토큰은 무작위 32바이트(256비트). 쿠키에는 원문(base64url), DB에는 SHA-256 해시만 둔다 (ARCHITECTURE 4절).
// 토큰 자체가 충분히 무작위라 느린 해시(argon2)가 필요 없다 — 맞혀 볼 수 있는 경우의 수가 2^256이다.
const tokenBytes = 32

// newSessionToken은 쿠키에 넣을 토큰과 DB에 넣을 해시를 만든다.
func newSessionToken() (token string, hash []byte, err error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(raw), sum[:], nil
}

// hashSessionToken은 쿠키 값에서 해시를 구한다. 형식이 틀리면 ok=false — 그런 쿠키는 세션이 없는 것과 같다.
func hashSessionToken(token string) (hash []byte, ok bool) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != tokenBytes {
		return nil, false
	}
	sum := sha256.Sum256(raw)
	return sum[:], true
}

// sessionCacheKey는 Valkey 세션 캐시 키다 (DATA_MODEL 4.2절).
func sessionCacheKey(hash []byte) string {
	return "sess:" + hex.EncodeToString(hash)
}
