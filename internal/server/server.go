// Package server는 HTTP 라우터를 조립한다 (ARCHITECTURE 2절).
// 미들웨어 · 오류 봉투(3.1 · 3.2절)는 M2에서 붙인다. 지금은 줄기가 통하는지만 본다.
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// New는 모든 경로를 등록한 핸들러 하나를 돌려준다.
// 비유: 건물 안내 데스크 — 주소(경로)를 보고 어느 방(핸들러)으로 보낼지 정한다.
func New(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	// 패턴 앞에 메서드를 적으면 다른 메서드로 온 요청에 405를 알아서 준다 (Go 1.22부터).
	mux.HandleFunc("GET /api/health", handleHealth(logger))
	return mux
}

// handleHealth는 지금은 "서버가 살아 있다"만 답한다.
// 다음 단계에서 Postgres · Valkey 연결 확인을 붙인다 (PLAN M1).
func handleHealth(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			// 응답을 쓰다 실패하면(클라이언트가 먼저 끊음 등) 더 할 수 있는 게 없어 기록만 남긴다.
			logger.Error("health 응답 쓰기 실패", "err", err)
		}
	}
}
