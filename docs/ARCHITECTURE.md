# 아키텍처

> **코드가 어떻게 나뉘고, 요청이 어떻게 흐르는지** 적는 문서다. 디렉터리는 폴더 단위까지만 적는다(파일 목록은 코드를 본다).
> 스택을 고른 이유는 [decisions/](decisions/README.md), 테이블 · Valkey 키는 [DATA_MODEL.md](DATA_MODEL.md)에 있다.

## 1. 한눈에 보기

```
                        ┌────────────────────────────────────┐
 브라우저 (SvelteKit SPA)│  app serve                          │──▶ PostgreSQL  (영구 데이터 · 진실의 원천)
 ── /api/* ────────────▶│   API (net/http)                    │
 ── 그 밖의 경로 ───────▶│   정적 파일 (go:embed, SPA fallback)│──▶ Valkey      (세션 캐시 · 요청 제한 ·
                        └────────────────────────────────────┘                 통계 캐시 · 초안 · 작업 큐)
                        ┌────────────────────────────────────┐                     ▲
                        │  app worker (asynq)                 │─────────────────────┘
                        │   리마인더 웹 푸시 · 주간 리포트    │──▶ PostgreSQL
                        └────────────────────────────────────┘
```

- 실행 파일은 **하나**(`app`)다. 서브커맨드로 역할을 고른다.

| 서브커맨드 | 하는 일 |
|---|---|
| `app serve` | API + 프론트 정적 파일. `/api/*`가 아닌 경로는 `index.html`로 돌려준다(SPA fallback) |
| `app worker` | 작업 큐 처리 + 주기 작업(리마인더 확인, 주간 리포트) |
| `app migrate up\|down\|status` | goose 마이그레이션 |

- Postgres와 Valkey의 역할 분담 원칙은 [decisions/0005](decisions/0005-valkey.md). 한 줄 요약: **Valkey가 통째로 날아가도 잃는 건 작성 중 초안뿐이어야 한다.**

## 2. 디렉터리 구조

```
cmd/app/               진입점 (serve · worker · migrate)
internal/
  config/              환경변수 읽기 · 검증 (빠진 값이 있으면 시작하지 않는다)
  server/              라우터 조립, 미들웨어, 오류 응답 봉투
  auth/                비밀번호 해시, 세션, 요청 제한 (인터페이스 + Postgres/Valkey 구현)
  checkin/  record/    기능별 패키지 ─ handler(HTTP ↔ 도메인) + service(규칙 · 트랜잭션)
  stats/  reminder/
  report/  account/
  db/                  queries/*.sql(직접 쓰는 SQL) · migrations/*.sql(goose) · sqlc 생성 코드
  cache/               Valkey 연결, 키 이름 함수
  jobs/                asynq 작업 정의 · 핸들러 · 스케줄
  apigen/              oapi-codegen 생성 코드 ★
  webdist/             프론트 빌드 결과를 embed하는 패키지
api/openapi.yaml       API 계약의 원본
web/                   SvelteKit 프로젝트
  src/lib/api/         openapi-typescript 생성 타입 ★ + openapi-fetch 클라이언트
  src/lib/components/ui/  shadcn-svelte 생성 부품 ★
  src/lib/components/  직접 만든 부품
  src/routes/          화면
compose.yaml           로컬 Postgres · Valkey
Taskfile.yml           자주 쓰는 명령 모음
```

★ = **생성 코드. 직접 고치지 않는다.** 원본(`openapi.yaml`, `queries/*.sql`)을 고치고 `task gen`으로 다시 만든다. 생성 코드도 커밋한다(빌드할 때 생성 도구가 없어도 되게).

**패키지 규칙**
- 기능별 패키지 안에서 `handler → service → db(sqlc)` 한 방향으로만 부른다. 별도 repository 계층은 두지 않는다(sqlc가 그 역할).
- 기능 패키지끼리는 서로 import하지 않는다. 같이 써야 하는 건 service 인터페이스로 넘긴다.
- **사용자 데이터를 읽고 쓰는 SQL은 전부 `user_id`를 조건에 넣는다.** 남의 기록은 0행 → 404. 이 규칙은 통합 테스트로 검사한다(다른 사용자로 조회 · 수정 · 삭제 → 404).

## 3. 요청 흐름

### 3.1 미들웨어 순서 (`/api/*`)
1. 패닉 복구 → 500 봉투
2. 요청 id 부여 + `slog` 접근 로그 (경로 · 상태 코드 · 걸린 시간 · 사용자 id만. **본문은 남기지 않는다** — REQUIREMENTS 5.1절)
3. `http.CrossOriginProtection` — 교차 출처 쓰기 요청 차단 (CSRF)
4. 세션 읽기 → 로그인한 사용자를 context에 넣음
5. 요청 제한 (로그인 · 가입 등 지정한 경로만)
6. 인증 필요 경로는 사용자가 없으면 401
7. handler

**기본은 잠금**이다. 인증 없이 열리는 경로는 로그인 · 가입 · 헬스체크만 명시적으로 연다.

### 3.2 응답 규칙
- JSON 필드는 `snake_case`, 시각은 RFC 3339 UTC(`...Z`).
- 오류는 한 가지 모양: `{"error": {"code": "NOT_FOUND", "message": "…", "fields": {"…": "…"}}}` (`fields`는 입력 오류일 때만).
- 400 입력 오류 / 401 비로그인 / 404 없음 · 남의 것 / 409 충돌 / 429 요청 제한 / 500 서버 오류.

## 4. 인증 흐름

```
로그인   이메일·비밀번호 ─▶ argon2id 검증 ─▶ 무작위 토큰 32바이트 생성
         ─▶ Postgres sessions에 토큰의 SHA-256 해시 저장 ─▶ 쿠키 session=<토큰>
매 요청  쿠키 토큰 ─▶ 해시 ─▶ Valkey 조회 ─(없으면)▶ Postgres 조회 후 Valkey에 TTL로 캐시
로그아웃 sessions 행 삭제 + Valkey 키 삭제 + 쿠키 만료
```

- 쿠키: `HttpOnly` · `Secure` · `SameSite=Lax`. 브라우저는 `http://localhost`를 안전한 출처로 쳐서 개발 중에도 `Secure` 쿠키가 동작한다.
- DB에는 토큰 원문이 아니라 **해시**만 둔다(DB가 새도 세션을 훔칠 수 없게).
- 세션 기한과 연장 방식, 비밀번호 변경 시 다른 세션 정리는 DATA_MODEL에서 정한다.
- 회원가입은 `ALLOW_SIGNUP`이 꺼져 있으면 403. 첫 계정은 `app` 명령으로 만든다(PLAN에서 정함).
- 프론트: API가 401을 주면 클라이언트 한 곳에서 로그인 상태를 지우고 `/login?next=지금주소`로 보낸다. 화면마다 따로 처리하지 않는다.

## 5. 백그라운드 작업 (`app worker`)
- 큐: `asynq` (Valkey). 작업은 **멱등**하게 만든다(같은 작업이 두 번 돌아도 알림 · 리포트가 두 번 생기지 않게).
- 주기 작업: asynq 스케줄러가 짧은 간격으로 "지금 보낼 리마인더 · 만들 리포트"를 Postgres에서 찾아 작업을 넣는다. 사용자 시간대 계산은 이 단계에서 한다.
- 웹 푸시: VAPID 키는 환경변수. 구독 정보는 Postgres. 만료된 구독(410 응답)은 지운다.
- 상세 설계는 리마인더 · 리포트 단계에서 이 절에 채운다.

## 6. 개발 환경

| 무엇 | 어디서 | 주소 |
|---|---|---|
| PostgreSQL | docker compose | `localhost:5436` |
| Valkey | docker compose | `localhost:6380` |
| API (`air`로 핫리로드) | 호스트 | `localhost:8080` |
| 프론트 dev 서버 | 호스트 (`pnpm dev`) | `localhost:5176` — `/api`를 8080으로 프록시 |

- 포트를 기본값에서 비켜 둔 이유: 두 PC 모두 PostgreSQL이 직접 설치돼 있어 5432가 차 있고, 다른 프로젝트 compose가 5433 · 5434 · 5435를 쓴다.
- 앱은 컨테이너 밖에서 돌린다. Windows에서 컨테이너 안 핫리로드는 파일 감지가 느리고 불안정하다.
- 개발 중에는 프록시 덕분에 같은 출처라 CORS 설정이 필요 없다. 운영에서도 같은 바이너리가 둘 다 주므로 마찬가지다.
- 필요한 도구: Go, Node 24 LTS + pnpm, Docker Desktop(WSL2), Task. Python은 쓰지 않는다.
- 자주 쓰는 명령은 `Taskfile.yml`에 모은다 (`task dev` · `task test` · `task lint` · `task gen` · `task migrate`). 정확한 목록은 CLAUDE.md.

## 7. 빌드
1. `web/`에서 `pnpm build` → 정적 파일
2. 결과를 `internal/webdist/`로 복사 → `go:embed`
3. `go build ./cmd/app` → 실행 파일 하나

배포 환경은 아직 정하지 않았다(첫 배포 단계에서 결정 기록으로 남긴다).

## 8. 테스트
| 층 | 도구 | 무엇을 |
|---|---|---|
| Go 단위 | `testing` | 계산 · 검증 규칙 (강도 범위, 통계 집계, 시간대 경계) |
| Go 통합 | `testing` + `testcontainers-go` | 실제 Postgres · Valkey를 띄워 API를 끝까지 호출. 소유자 검사(남의 것 404), 인증 · 요청 제한 |
| 프론트 단위 | Vitest | 폼 검증, 표시 포맷 |
| E2E | Playwright | 핵심 흐름만: 로그인 → 체크인 → 사고 기록 작성 · 임시저장 → 목록 · 대시보드 |

통합 테스트는 Docker가 켜져 있어야 돈다.

---

## 변경 이력
| 날짜 | 내용 |
|---|---|
| 2026-10-29 | v2 첫 작성 |
