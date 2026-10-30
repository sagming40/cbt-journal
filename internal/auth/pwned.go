package auth

import (
	"bufio"
	"context"
	"crypto/sha1" //nolint:gosec // Pwned Passwords API가 SHA-1 앞 5글자를 요구한다. 저장 · 보안 비교에는 쓰지 않는다
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// BreachChecker는 비밀번호가 알려진 유출 목록에 있는지 본다 (REQUIREMENTS A10).
// 인터페이스로 둬서 테스트에서는 가짜를 꽂는다.
type BreachChecker interface {
	IsBreached(ctx context.Context, password string) (bool, error)
}

// PwnedPasswords는 Have I Been Pwned의 Pwned Passwords range API로 확인한다 (ARCHITECTURE 4절).
// 비밀번호의 SHA-1 해시 앞 5글자만 보내고, 같은 앞 5글자로 시작하는 해시 수백 개를 받아 나머지 35글자를 여기서 비교한다(k-익명성).
// 비유: "김씨 명단 전부 줘" 하고 받아서, 내 이름이 있는지는 집에서 찾는다 — 상대는 내가 누군지 모른다.
type PwnedPasswords struct {
	BaseURL string // 끝에 /가 붙은 주소. 테스트에서 가짜 서버로 바꾼다
	Client  *http.Client
}

// NewPwnedPasswords는 실제 API를 쓰는 검사기를 만든다.
// 2초 안에 답이 없으면 실패로 끝낸다 — 남의 서버가 느려도 가입 · 비밀번호 변경이 매달리지 않게 (호출하는 쪽이 건너뛴다).
func NewPwnedPasswords() *PwnedPasswords {
	return &PwnedPasswords{
		BaseURL: "https://api.pwnedpasswords.com/range/",
		Client:  &http.Client{Timeout: 2 * time.Second},
	}
}

// IsBreached는 유출 목록에 있으면 true. 네트워크 · 응답 오류는 err로 돌려주고, 건너뛸지는 호출하는 쪽이 정한다.
func (p *PwnedPasswords) IsBreached(ctx context.Context, password string) (bool, error) {
	sum := sha1.Sum([]byte(password)) //nolint:gosec // 위 import 주석과 같은 이유
	full := strings.ToUpper(hex.EncodeToString(sum[:]))
	prefix, suffix := full[:5], full[5:]

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.BaseURL+prefix, nil)
	if err != nil {
		return false, err
	}
	// Add-Padding: 응답에 가짜 줄(개수 0)을 섞어, 응답 크기로 어떤 앞 5글자를 물었는지 짐작하지 못하게 한다.
	req.Header.Set("Add-Padding", "true")
	req.Header.Set("User-Agent", "cbt-journal")

	resp, err := p.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("pwned passwords: 상태 코드 %d", resp.StatusCode)
	}

	// 한 줄이 "나머지35글자:유출횟수". 패딩 줄은 횟수가 0이라 유출로 치지 않는다.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		hashPart, count, ok := strings.Cut(line, ":")
		if !ok || !strings.EqualFold(hashPart, suffix) {
			continue
		}
		return count != "0", nil
	}
	return false, scanner.Err()
}
