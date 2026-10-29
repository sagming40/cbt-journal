# 0002. 백엔드는 Go로 만든다

- 상태: 채택
- 날짜: 2026-10-29

## 맥락
- 사용자는 Express(v1), FastAPI, Django를 이미 써 봤고 Python 백엔드에 질렸다. 취미 프로젝트라 **새로 배우는 재미**가 선택 기준에 들어간다.
- 동시에 "제대로" 만들고 싶다: 테스트·린트·마이그레이션·구조화 로그가 갖춰져야 한다.
- 회사를 다니며 하는 사이드 프로젝트라 운영 부담이 작아야 한다.

## 결정
Go로 만든다. 프레임워크 없이 표준 `net/http`를 쓰고, 꼭 필요한 곳에만 도구를 붙인다.

| 영역 | 선택 |
|---|---|
| HTTP | 표준 `net/http` (1.22+ 경로 · 메서드 라우팅) |
| DB | `pgx` + `sqlc`(SQL → 타입 있는 Go 코드), 마이그레이션 `goose` |
| API 계약 | OpenAPI 스펙을 먼저 쓰고 `oapi-codegen`으로 서버 인터페이스 생성 |
| 인증 | 직접 구현: DB 세션 + 세션 쿠키, argon2id, `http.CrossOriginProtection` |
| 작업 큐 | `asynq` (Valkey 기반, [0005](0005-valkey.md)) |
| 로그 · 테스트 · 린트 | `log/slog` · `testing` + `testcontainers-go` · `golangci-lint` |

빌드 결과는 **바이너리 하나**다. `serve`(API + 프론트 정적 파일)와 `worker`(알림 · 리포트) 서브커맨드로 역할을 나눈다.

## 검토한 대안
- **Kotlin + Spring Boot**: 탄탄하지만 무겁고 사이드 프로젝트에선 보일러플레이트가 크다.
- **Rust + Axum**: 가장 재밌지만 진도가 느려 앱 완성이 멀어진다.
- **Elixir + Phoenix LiveView**: 독특하지만 생태계가 작다.
- **TypeScript(Hono · Next.js)**: 생산성은 최고지만 Express에서 크게 안 벗어난다.

## 감수할 것
- Better Auth 같은 TS 인증 라이브러리를 못 쓴다 → 인증을 직접 구현한다(공부거리로 본다).
- 프론트와 타입 공유가 끊긴다 → OpenAPI 스펙 하나에서 Go 서버와 TS 클라이언트를 같이 생성해 메운다.
