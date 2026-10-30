package server

import (
	"encoding/json"
	"net/http"
)

// writeError는 오류 응답을 한 가지 모양으로 쓴다 (ARCHITECTURE 3.2절):
// {"error": {"code": "NOT_FOUND", "message": "…", "fields": {"칸": "문구"}}}
// fields는 입력 오류일 때만. message는 화면에 그대로 보여 줄 수 있게 반말이다 (DESIGN 8장).
func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	body := map[string]any{"code": code, "message": message}
	if len(fields) > 0 {
		body["fields"] = fields
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": body})
}
