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
| `app serve` | API + 프론트 정적 파일. `/api/*`가 아닌 경로는 `index.html`로 돌려준다(SPA fallback). 없는 `/api/*` GET은 JSON 404 |
| `app worker` | 작업 큐 처리 + 주기 작업(리마인더 확인, 주간 리포트) |
| `app migrate up\|down\|status` | goose 마이그레이션 |
| `app user create\|reset-password` | 계정 만들기 · 비밀번호 재설정 (가입이 막힌 동안의 관리 수단, REQUIREMENTS A9) |

- 라우팅: API 경로는 `openapi.yaml`에서 생성한 코드가 등록하고, 프론트는 **`GET /`** 하나로 받는다(가장 덜 구체적인 패턴이라 API가 먼저 고른다). 있는 파일은 그 파일, 없는 경로 · 폴더는 `index.html`, `/api/`로 시작하는데 여기까지 온 GET은 JSON 404다. 메서드 없이 `/`로 열지 않는 이유: `POST /api/health` 같은 요청이 405 대신 404가 된다. 대신 없는 `/api/*`에 GET이 아닌 메서드는 405다.
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
1. 요청 id 부여 + `slog` 접근 로그 (경로 · 상태 코드 · 걸린 시간 · 사용자 id만. **본문은 남기지 않는다** — REQUIREMENTS 5.1절). 가장 바깥에 둬야 패닉으로 끝난 요청도 500으로 로그에 남는다
2. 패닉 복구 → 500 봉투
3. `http.CrossOriginProtection` — 교차 출처 쓰기 요청 차단 (CSRF)
4. 세션 읽기 → 로그인한 사용자를 context에 넣음
5. 요청 제한 (로그인 · 가입 등 지정한 경로만)
6. 인증 필요 경로는 사용자가 없으면 401
7. handler

**기본은 잠금**이다. 인증 없이 열리는 경로는 로그인 · 가입 · 헬스체크 · 공개 설정(`GET /api/config` — 가입 허용 여부처럼 로그인 화면이 알아야 하는 값만)을 명시적으로 연다.

### 3.2 응답 규칙
- JSON 필드는 `snake_case`, 시각은 RFC 3339 UTC(`...Z`).
- 오류는 한 가지 모양: `{"error": {"code": "NOT_FOUND", "message": "…", "fields": {"…": "…"}}}` (`fields`는 입력 오류일 때만).
- 400 입력 오류 / 401 비로그인 / 403 막힌 기능(가입 꺼짐) / 404 없음 · 남의 것 / 409 충돌 / 429 요청 제한 / 500 서버 오류.

### 3.3 실제 접속 IP
요청 제한 · 로그는 접속한 사람의 IP를 쓴다. 리버스 프록시 뒤에서는 모든 요청의 `RemoteAddr`가 프록시 하나로 보여서, 그대로 쓰면 모든 사용자가 같은 제한을 나눠 갖는다.
- `TRUSTED_PROXIES`(환경변수)에 적힌 주소에서 온 요청일 때만 `X-Forwarded-For`의 **오른쪽 끝 값**(그 프록시가 직접 붙인 값)을 쓴다. 그 밖에는 `RemoteAddr`.
- 헤더의 왼쪽 값들은 클라이언트가 마음대로 넣을 수 있으므로 믿지 않는다.
- 개발 환경은 `TRUSTED_PROXIES`를 비워 둔다(프록시 없음).

## 4. 인증 흐름

```
로그인   이메일·비밀번호 ─▶ argon2id 검증 ─▶ 무작위 토큰 32바이트 생성
         ─▶ Postgres sessions에 토큰의 SHA-256 해시 + 로그인 유지 여부 저장 ─▶ 쿠키 session=<토큰>
매 요청  쿠키 토큰 ─▶ 해시 ─▶ Valkey 조회 ─(없으면)▶ Postgres 조회 후 Valkey에 TTL로 캐시
로그아웃 sessions 행 삭제 + Valkey 키 삭제 + 쿠키 만료
```

- 쿠키: `HttpOnly` · `Secure` · `SameSite=Lax`. 브라우저는 `http://localhost`를 안전한 출처로 쳐서 개발 중에도 `Secure` 쿠키가 동작한다.
- DB에는 토큰 원문이 아니라 **해시**만 둔다(DB가 새도 세션을 훔칠 수 없게).
- **로그인 유지**(REQUIREMENTS A1): 고르면 쿠키에 `Max-Age`를 줘서 브라우저를 닫아도 남고, 안 고르면 `Max-Age` 없는 세션 쿠키라 브라우저를 닫으면 사라진다. 서버 쪽 기한도 다르게 둔다(DATA_MODEL 3.2절).
- 없는 이메일이어도 고정된 가짜 해시로 argon2id 검증을 한 번 돌린다. 응답 문구뿐 아니라 걸리는 시간으로도 가입 여부가 드러나지 않게.
- **세션을 지울 때는 Valkey 캐시도 같이 지운다**: 로그아웃 · 비밀번호 변경 · 다른 기기 모두 로그아웃 · 탈퇴. Postgres에서 `DELETE … RETURNING token_hash`로 지운 해시를 받아 `sess:` 키를 지운다. 안 지우면 캐시 TTL(10분) 동안 옛 세션이 살아 있다.
- 세션 기한과 연장 방식은 DATA_MODEL 3.2절.
- 회원가입은 `ALLOW_SIGNUP`이 꺼져 있으면 403. 첫 계정은 `app user create`로 만든다.
- **유출 비밀번호 검사**(REQUIREMENTS A10): 새 비밀번호를 정하는 곳은 전부 `auth` 패키지의 같은 검사기(인터페이스)를 거친다. SHA-1 해시 앞 5글자로 Pwned Passwords range API를 부르고 응답 목록과 나머지 35글자를 서버에서 비교한다. 짧은 타임아웃을 두고, 실패 · 타임아웃이면 경고 로그(비밀번호 · 해시는 남기지 않음)만 남기고 통과시킨다. 테스트에서는 가짜 검사기 또는 가짜 HTTP 서버를 꽂는다.
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
- **휴대폰 확인은 배포 서버로 한다**(PLAN M3 이후). 개발 서버에 휴대폰을 붙이려 하지 않는다 — `Secure` 쿠키는 `http://IP주소`에서 저장되지 않는다.
- 자주 쓰는 명령은 `Taskfile.yml`에 모은다 (`task dev` · `task test` · `task lint` · `task gen` · `task migrate`). 정확한 목록은 CLAUDE.md.

## 7. 빌드
1. `web/`에서 `pnpm build` → `web/build/` (정적 파일)
2. 결과를 `internal/webdist/dist/`로 복사 → `//go:embed all:dist` (`go:embed`는 자기 폴더 아래만 넣을 수 있어서 복사한다. `all:`이 없으면 밑줄로 시작하는 `_app/`이 빠진다)
3. `go build ./cmd/app` → 실행 파일 하나

- `internal/webdist/dist/`에는 `.gitkeep`만 커밋한다. 프론트를 빌드하지 않은 PC · CI에서도 `go build` · `go test`가 되고, 그 상태로 화면 주소에 들어오면 503 안내를 준다.
- 캐시: `_app/immutable/`(파일 이름에 내용 해시)은 1년 immutable, `index.html`은 `no-cache`.

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
| 2026-10-29 | 문서 검토 반영: 미들웨어 순서(로그를 패닉 복구 바깥으로), `app user` 서브커맨드, 공개 설정 `GET /api/config`, 403 추가, 3.3절 실제 접속 IP(`TRUSTED_PROXIES`), 4절 로그인 유지 · 가짜 해시 · 세션 삭제 시 Valkey 캐시 삭제, 6절 휴대폰 확인은 배포 서버로 |
| 2026-10-30 | 4절 유출 비밀번호 검사(A10) 추가 |
| 2026-10-30 | 1절 라우팅(`GET /`로 SPA, 없는 `/api/*` GET은 JSON 404), 7절 embed 세부(`dist/` · `all:` · `.gitkeep` · 캐시) |
