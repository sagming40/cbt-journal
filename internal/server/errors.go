package server

import (
	"encoding/json"
	"net/http"
)

// writeError는 오류 응답을 한 가지 모양으로 쓴다 (ARCHITECTURE 3.2절):
// {"error": {"code": "NOT_FOUND", "message": "…"}}
// 입력 오류의 fields와 나머지 코드(400 · 401 · 409 …)는 M2에서 더한다.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}
