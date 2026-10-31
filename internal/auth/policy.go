package auth

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/sagming40/cbt-journal/internal/apierr"

	// 시간대 DB를 실행 파일에 넣는다. Windows에는 OS에 IANA 시간대 DB가 없어서 이게 없으면 LoadLocation이 실패한다.
	_ "time/tzdata"
)

// 입력 규칙 (REQUIREMENTS 5.2절 · DATA_MODEL 3.1절). DB CHECK 제약과 숫자를 맞춘다.
const (
	minPasswordLen = 12
	maxPasswordLen = 128
	maxEmailLen    = 254
	maxNicknameLen = 30
)

// ValidationError는 입력 오류다. 모양은 기능 패키지들이 같이 쓰는 apierr에 있다 (ARCHITECTURE 2절).
// 별칭이라 auth.ValidationError로 써도 같은 타입이다(errors.As도 그대로 맞는다).
type ValidationError = apierr.ValidationError

func fieldError(field, message string) *ValidationError {
	return apierr.Field(field, message)
}

// NormalizeEmail은 이메일을 다듬고(앞뒤 공백 제거, 소문자) 형식을 확인한다.
// 소문자로 저장해야 "Me@x.com"과 "me@x.com"이 다른 계정이 되지 않는다 (DATA_MODEL 3.1절).
func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", fieldError("email", "이메일을 입력해 주세요")
	}
	if len(email) > maxEmailLen {
		return "", fieldError("email", "이메일이 너무 길어요")
	}
	// ParseAddress는 "이름 <주소>" 꼴도 받아 주므로, 주소만 있는 경우인지 다시 확인한다.
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || addr.Name != "" {
		return "", fieldError("email", "이메일 형식이 아니에요")
	}
	return email, nil
}

// ValidatePassword는 길이만 본다. 문자 종류 섞기 규칙은 두지 않는다 — 길이와 유출 여부가 실제로 효과 있는 기준이다 (REQUIREMENTS 5.2절).
// 길이는 바이트가 아니라 글자 수로 센다(한글 비밀번호도 같은 기준).
func ValidatePassword(field, password string) error {
	n := utf8.RuneCountInString(password)
	switch {
	case n < minPasswordLen:
		return fieldError(field, fmt.Sprintf("%d자 이상이어야 해요", minPasswordLen))
	case n > maxPasswordLen:
		return fieldError(field, fmt.Sprintf("%d자까지 쓸 수 있어요", maxPasswordLen))
	}
	return nil
}

// NormalizeNickname은 앞뒤 공백을 지우고 길이 · 제어 문자를 확인한다.
func NormalizeNickname(raw string) (string, error) {
	nickname := strings.TrimSpace(raw)
	n := utf8.RuneCountInString(nickname)
	if n == 0 {
		return "", fieldError("nickname", "닉네임을 입력해 주세요")
	}
	if n > maxNicknameLen {
		return "", fieldError("nickname", fmt.Sprintf("%d자까지 쓸 수 있어요", maxNicknameLen))
	}
	for _, r := range nickname {
		if unicode.IsControl(r) {
			return "", fieldError("nickname", "쓸 수 없는 글자가 있어요")
		}
	}
	return nickname, nil
}

// ValidateTimezone은 IANA 시간대 이름인지 확인한다 (REQUIREMENTS A4).
// LoadLocation은 ""(UTC)와 "Local"(서버 시간대)도 받아 주는데, 둘 다 사용자의 "하루"를 정하는 값이 아니라 거부한다.
func ValidateTimezone(tz string) error {
	if tz == "" || tz == "Local" {
		return fieldError("timezone", "시간대를 골라 주세요")
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return fieldError("timezone", "알 수 없는 시간대예요")
	}
	return nil
}
