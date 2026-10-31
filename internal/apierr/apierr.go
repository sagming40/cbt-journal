// Package apierr는 API 오류 봉투(ARCHITECTURE 3.2절)와 입력 오류를 모은다.
// 기능 패키지(auth · checkin …)끼리는 서로 import하지 않으므로(ARCHITECTURE 2절),
// 모두가 같은 모양으로 써야 하는 오류는 기능이 아닌 이 작은 패키지에 둔다.
// 비유: 부서마다 서식을 따로 만들지 않고 총무팀 서식 하나를 같이 쓴다.
package apierr

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sagming40/cbt-journal/internal/apigen"
)

// ValidationError는 입력 오류를 칸별 문구로 모은다. API에서 400 봉투의 fields가 된다 (ARCHITECTURE 3.2절).
// 문구는 화면에 그대로 보이므로 부드러운 존댓말(해요체)로 쓴다 (DESIGN 8장).
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, e.Fields[k]))
	}
	return "입력 오류: " + strings.Join(parts, ", ")
}

// Field는 칸 하나짜리 입력 오류를 만든다.
func Field(field, message string) *ValidationError {
	return &ValidationError{Fields: map[string]string{field: message}}
}

// Body는 오류 봉투를 계약서 타입으로 만든다.
func Body(code, message string) apigen.Error {
	var e apigen.Error
	e.Error.Code = code
	e.Error.Message = message
	return e
}

// Validation은 입력 오류 400 봉투다. fields에 칸별 문구가 들어간다.
func Validation(ve *ValidationError) apigen.Error {
	e := Body("INVALID_INPUT", "입력한 내용을 다시 확인해 주세요")
	fields := ve.Fields
	e.Error.Fields = &fields
	return e
}

// Unauthorized는 세션 없음 401 봉투다. 인증 미들웨어와 handler가 같은 모양을 쓴다.
func Unauthorized() apigen.Error {
	return Body("UNAUTHORIZED", "로그인이 필요해요")
}

// NotFound는 404 봉투다. 남의 것도 같은 응답 — 있는지조차 알려 주지 않는다 (ARCHITECTURE 2절).
func NotFound() apigen.Error {
	return Body("NOT_FOUND", "찾을 수 없는 기록이에요")
}
