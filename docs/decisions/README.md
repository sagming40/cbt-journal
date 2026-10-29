# decisions — 결정 기록 (ADR)

"왜 이렇게 정했나"를 남기는 곳이다. 결정 하나 = 파일 하나.

- 파일 이름: `NNNN-영문-slug.md` (번호는 4자리, 순서대로. Windows 도구 호환을 위해 파일 이름은 영문)
- 지난 결정은 고치지 않는다. 결정을 뒤집으면 **새 파일**을 만들고, 옛 파일의 상태를 `대체됨 → NNNN`으로만 바꾼다
- 형식: 상태 · 날짜 / 맥락 / 결정 / 검토한 대안 / 감수할 것

| 번호 | 결정 | 상태 |
|---|---|---|
| [0001](0001-rewrite-v2.md) | v1을 버리고 전면 재작성, "CBT 연습 도구"로 방향 전환 | 채택 |
| [0002](0002-backend-go.md) | 백엔드는 Go | 채택 |
| [0003](0003-frontend-sveltekit-spa.md) | 프론트엔드는 SvelteKit SPA + shadcn-svelte | 채택 |
| [0004](0004-db-postgresql.md) | 메인 DB는 PostgreSQL | 채택 |
| [0005](0005-valkey.md) | Valkey를 붙이고 Postgres와 역할을 나눈다 | 채택 |
| [0006](0006-ai-coach-deferred.md) | AI 코치는 핵심 기능 뒤 옵트인으로 보류 | 채택 |
