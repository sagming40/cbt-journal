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

- 라우팅: 맨 바깥은 경로 앞부분만 보고 `/api/`는 API mux(미들웨어 줄을 거침, 3.1절)로, 나머지는 프론트 mux(**`GET /`**만)로 보낸다. (메서드 없는 `/api/`와 `GET /`를 한 mux에 등록하면 Go가 "겹친다"며 거부해서 mux로 나누지 않는다.) API mux에는 `openapi.yaml`에서 생성한 코드가 경로를 등록하고, 거기 없는 `GET /api/*`는 JSON 404다. 프론트는 있는 파일은 그 파일, 없는 경로 · 폴더는 `index.html`. 메서드 없는 catch-all을 두지 않는 이유: `POST /api/health` 같은 요청이 405 대신 404가 된다. 대신 없는 `/api/*`에 GET이 아닌 메서드는 405다. 정적 파일 요청은 세션 조회 같은 API 미들웨어를 거치지 않는다.
- Postgres와 Valkey의 역할 분담 원칙은 [decisions/0005](decisions/0005-valkey.md). 한 줄 요약: **Valkey가 통째로 날아가도 잃는 건 작성 중 초안뿐이어야 한다.**

## 2. 디렉터리 구조

```
cmd/app/               진입점 (serve · worker · migrate)
internal/
  config/              환경변수 읽기 · 검증 (빠진 값이 있으면 시작하지 않는다)
  server/              라우터 조립, 미들웨어, 오류 응답 봉투
  auth/                비밀번호 해시, 세션, 요청 제한, 유출 비밀번호 검사 + 인증 API handler
  clientip/            실제 접속 IP 미들웨어 (3.3절)
  checkin/  record/    기능별 패키지 ─ handler(HTTP ↔ 도메인) + service(규칙 · 트랜잭션)
  stats/  reminder/
  report/  account/
  db/                  queries/*.sql(직접 쓰는 SQL) · migrations/*.sql(goose) · 연결 · 마이그레이션 실행
    dbgen/             sqlc 생성 코드 ★
  testutil/            통합 테스트용 Postgres · Valkey 컨테이너 (testcontainers)
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
- 생성된 `StrictServerInterface`는 타입 하나가 전부 구현해야 한다. 기능 패키지마다 자기 경로의 메서드를 가진 `Handler`를 만들고, `server`의 handler가 그것들을 **임베드**해서 합친다(메서드가 위로 올라와 인터페이스를 채운다).
- 기능 패키지끼리는 서로 import하지 않는다. 같이 써야 하는 건 service 인터페이스로 넘긴다.
- **사용자 데이터를 읽고 쓰는 SQL은 전부 `user_id`를 조건에 넣는다.** 남의 기록은 0행 → 404. 이 규칙은 통합 테스트로 검사한다(다른 사용자로 조회 · 수정 · 삭제 → 404).

## 3. 요청 흐름

### 3.1 미들웨어 순서 (`/api/*`)
1. 요청 id 부여(`X-Request-ID` 응답 헤더) + `slog` 접근 로그 (경로 · 상태 코드 · 걸린 시간 · 사용자 id만. **본문은 남기지 않는다** — REQUIREMENTS 5.1절). 가장 바깥에 둬야 패닉으로 끝난 요청도 500으로 로그에 남는다
2. 패닉 복구 → 500 봉투
3. 요청 본문 1MiB 제한 (REQUIREMENTS 5.2절)
4. `http.CrossOriginProtection` — 교차 출처 쓰기 요청 차단 (CSRF)
5. 실제 접속 IP를 context에 (3.3절)
6. 세션 읽기 → 로그인한 사용자를 context에 넣음, 기한 연장 시 쿠키 다시 발급 (4절)
7. 인증 필요 경로는 사용자가 없으면 401 — 생성 코드가 표시해 준 경로만 (아래 "기본 잠금")
8. handler

요청 제한은 미들웨어가 아니라 **handler 안에서** 한다. 계정별 제한은 요청 본문의 이메일을 알아야 하고, 실패했을 때만 세는 규칙이 경로마다 달라서다(4절).

**기본은 잠금**이다. `openapi.yaml` 전체에 `security: cookieAuth`를 걸고, 인증 없이 열 경로만 `security: []`로 연다: 로그인 · 가입 · 헬스체크 · 공개 설정(`GET /api/config` — 가입 허용 여부처럼 로그인 화면이 알아야 하는 값만). 생성 코드가 인증이 필요한 경로의 요청 context에 표시(`CookieAuthScopes`)를 넣어 주므로, 7번 미들웨어는 그 표시만 보고 401을 준다. 새 경로를 추가하면 따로 하지 않아도 잠긴다. 통합 테스트가 "공개 경로 말고는 전부 401"을 확인한다.

### 3.2 응답 규칙
- JSON 필드는 `snake_case`, 시각은 RFC 3339 UTC(`...Z`).
- 오류는 한 가지 모양: `{"error": {"code": "NOT_FOUND", "message": "…", "fields": {"…": "…"}}}` (`fields`는 입력 오류일 때만).
- 400 입력 오류 / 401 비로그인 · 로그인 실패 / 403 막힌 기능(가입 꺼짐) · 교차 출처 요청 / 404 없음 · 남의 것 / 409 충돌 / 413 본문 1MiB 초과 / 429 요청 제한(`Retry-After` 초) / 500 서버 오류.
- 지금 쓰는 `code`: `INVALID_INPUT`(fields 있음) · `INVALID_REQUEST`(JSON 해석 실패) · `UNAUTHORIZED` · `INVALID_CREDENTIALS` · `SIGNUP_DISABLED` · `CROSS_ORIGIN` · `NOT_FOUND` · `EMAIL_TAKEN` · `PAYLOAD_TOO_LARGE` · `RATE_LIMITED` · `INTERNAL`.
- 요청 본문을 해석하지 못한 경우(잘못된 JSON 등)와 handler 안에서 난 예상 못 한 오류도 같은 봉투로 준다(400 · 500). mux가 직접 주는 405만 글자 응답이다.

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
- 세션 기한과 연장 방식은 DATA_MODEL 3.2절. **로그인 유지 세션의 기한을 늘릴 때는 쿠키도 `Max-Age` 30일로 다시 내려준다.** 서버 기한만 늘리면 쿠키가 로그인 30일 뒤에 그대로 사라진다.
- 쿠키 이름은 `session`. 형식이 틀리거나 없는 세션의 쿠키가 오면 만료시키는 쿠키를 내려 지운다.
- 회원가입은 `ALLOW_SIGNUP`이 꺼져 있으면 403. 첫 계정은 `app user create`로 만든다.
- **요청 제한** (Valkey 카운터, DATA_MODEL 4.2절). 넘으면 429 + `Retry-After`(키의 남은 TTL).
  | 대상 | 기준 | 세는 것 |
  |---|---|---|
  | 로그인 · IP | 15분에 20회 | 모든 시도 |
  | 로그인 · 계정 | 15분에 5회 | 실패만. 성공하면 초기화 |
  | 가입 · IP | 1시간에 5회 | 모든 시도 |
  | 비밀번호 변경 · 사용자 | 15분에 5회 | "현재 비밀번호" 틀림만 |
  - 계정별 제한은 남이 일부러 틀려서 15분 잠글 수 있다(M0 검토에서 감수하기로 함). 제한 검사는 비밀번호 검증 **전에** 해서, 잠긴 동안은 맞는 비밀번호도 429다(맞았는지 알려 주지 않게).
- **Valkey가 응답하지 않을 때**: 세션은 Postgres에서 읽고, 요청 제한은 통과시킨다(로그만 남김). Postgres가 진실의 원천이고 Valkey는 없어도 돌아야 한다(decisions 0005). 감수할 점: Valkey가 잠깐 죽은 사이에 세션을 지우면 캐시 삭제가 실패해, Valkey가 돌아왔을 때 옛 캐시가 최대 10분 남을 수 있다(로그로 남김).
- 로그인할 때 그 사용자의 만료된 세션을 같이 지운다. 전체 정리는 worker가 생기면(M7) 하루 한 번.
- **유출 비밀번호 검사**(REQUIREMENTS A10): 새 비밀번호를 정하는 곳은 전부 `auth` 패키지의 같은 검사기(인터페이스)를 거친다. SHA-1 해시 앞 5글자로 Pwned Passwords range API를 부르고 응답 목록과 나머지 35글자를 서버에서 비교한다. 짧은 타임아웃을 두고, 실패 · 타임아웃이면 경고 로그(비밀번호 · 해시는 남기지 않음)만 남기고 통과시킨다. 테스트에서는 가짜 검사기 또는 가짜 HTTP 서버를 꽂는다.
- 프론트: API가 401을 주면 클라이언트 한 곳에서 로그인 상태를 지우고 `/login?next=지금주소`로 보낸다. 화면마다 따로 처리하지 않는다. **예외는 로그인 · 가입 요청의 401**(비밀번호 틀림)이다 — 그 화면에서 문구로 보여 준다. 안 빼면 로그인 화면에서 로그인 화면으로 튕긴다.
- `next`는 앱 안 경로만 받는다: `/`로 시작하고, `//`나 `/\`로 시작하지 않는 것. 브라우저는 `/\evil.com`을 `//evil.com`처럼 바깥 주소로 읽는다(DESIGN 3.1절).

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

## 7. 빌드 · 배포 · 운영

### 7.1 빌드
1. `web/`에서 `pnpm build` → `web/build/` (정적 파일)
2. 결과를 `internal/webdist/dist/`로 복사 → `//go:embed all:dist` (`go:embed`는 자기 폴더 아래만 넣을 수 있어서 복사한다. `all:`이 없으면 밑줄로 시작하는 `_app/`이 빠진다)
3. `go build ./cmd/app` → 실행 파일 하나

- `internal/webdist/dist/`에는 `.gitkeep`만 커밋한다. 프론트를 빌드하지 않은 PC · CI에서도 `go build` · `go test`가 되고, 그 상태로 화면 주소에 들어오면 503 안내를 준다.
- 캐시: `_app/immutable/`(파일 이름에 내용 해시)은 1년 immutable, `index.html`은 `no-cache`.

- 운영용은 `task build:linux` — 같은 순서에 `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`, 그리고 **커밋 번호를 실행 파일에 새긴다**(`-ldflags "-X main.version=<커밋>"`). `app version`과 `serve` 시작 로그에 나온다. 의존성이 전부 순수 Go(pgx · argon2)라 Windows에서 바로 Linux용을 만든다.

### 7.2 운영 구성
배포 환경: Vultr 서울 서버 한 대 + 도메인 `cbt.sagming.com`(Cloudflare, DNS만) — [decisions/0007](decisions/0007-deploy-vultr-seoul.md). 서버 접속 · 실제로 친 명령은 [SERVER.md](SERVER.md).

```
바깥 ─ 443 · 80 ─▶ Caddy (인증서 자동 · 보안 헤더 · http→https)
                     │ 127.0.0.1:8080
                     ▼
                   app serve  ← systemd 서비스 cbt, 로그인 불가 계정 cbt로 실행
                     │ 127.0.0.1만
                     ▼
              docker compose: Postgres(5432) · Valkey(6379, 비밀번호)
```

| 서버 위치 | 내용 |
|---|---|
| `/opt/cbt/releases/<커밋>/app` | 배포한 실행 파일. 최근 5개만 남긴다 |
| `/opt/cbt/current` | 지금 버전을 가리키는 링크. 되돌리기 = 링크만 바꾸고 재시작 |
| `/etc/cbt/cbt.env` | 앱 비밀값 · 설정(`root:cbt`, 640). systemd가 읽는다 |
| `/srv/cbt/compose.yaml` · `/srv/cbt/.env` | 운영 Postgres · Valkey와 그 비밀번호(600) |
| `/var/backups/cbt/` | 암호화된 DB 백업. 최근 14개 |
| `/usr/local/bin/cbtctl` | 운영 env를 읽고 `cbt` 계정으로 `app`을 실행하는 도우미 — `sudo cbtctl user create …` · `sudo cbtctl migrate status` |

운영 env: `APP_ENV=production` · `HTTP_ADDR=127.0.0.1:8080` · `TRUSTED_PROXIES=127.0.0.1` · `ALLOW_SIGNUP=false` · DB · Valkey 주소(비밀번호는 서버에서 무작위로 만든다).

저장소의 운영 파일: `deploy/`(운영 compose · Caddyfile · systemd 유닛 · `cbtctl` · 서버 쪽 배포 스크립트 · 백업 스크립트 · env 예시)와 `scripts/deploy.ps1`(PC 쪽). 서버에는 배포 때 필요한 것만 복사된다.

**Caddy가 하는 일**
- `cbt.sagming.com`만 받는다. IP로 들어오거나 다른 이름이면 앱을 보여 주지 않는다.
- 보안 헤더: `X-Content-Type-Options: nosniff` · `X-Frame-Options: DENY` · `Referrer-Policy: strict-origin-when-cross-origin` · HSTS(**처음엔 하루짜리**, 확인 뒤 1년). CSP는 화면이 생긴 뒤(M4~) 인라인 스크립트 해시와 함께 붙인다.
- 바깥에서 온 `X-Forwarded-For`는 버리고 자기가 새로 붙인다 → 3.3절 규칙으로 실제 접속 IP가 나온다. `Host`도 그대로 넘겨 교차 출처 차단(3.1절)이 동작한다.
- 접근 로그는 남기지 않는다(앱 로그로 충분하고, 남기는 곳을 늘리지 않는다 — REQUIREMENTS 5.1절). 인증서 등록 이메일은 넣지 않는다 — Let's Encrypt가 2025년에 만료 알림 메일을 없앴고 갱신은 Caddy가 자동으로 해서 쓸 곳이 없다(2026-10-31, 처음엔 `ACME_EMAIL` env로 빼려 했음). 저장소에 이메일이 들어갈 일도 없다.
- IP로 `https://`에 들어오면 그 IP용 인증서가 없어 연결 자체가 안 된다. `http://IP`는 Caddy가 빈 응답을 준다. 어느 쪽이든 앱은 안 보인다(따로 "그 밖의 이름" 블록을 두지 않는다 — 자동 http→https 전환 규칙과 엉킬 수 있어서).

### 7.3 배포 흐름 (`task deploy`, PC에서)
1. **확인**: 작업 트리가 깨끗하고, 지금 커밋이 `origin/main`과 같아야 한다(push 안 된 코드 · 커밋 안 된 변경은 거부). 두 PC가 서로 다른 버전을 덮어쓰지 않는다.
2. **빌드**: `task build:linux` → `bin/app-linux-amd64`(커밋 번호 포함).
3. **복사** 1번: 실행 파일 + 서버 쪽 배포 스크립트를 서버 임시 폴더로(`scp`).
4. **서버에서 실행** 1번(`ssh` → `sudo`, `deploy/release.sh <커밋>`): 아래를 한 번에 한다.
   1. `releases/<커밋>/`에 넣는다. 넣기 전에 실행 파일의 `app version`이 넘겨받은 커밋과 같은지 본다(엉뚱한 파일 · `-dirty` 빌드 거부)
   2. **배포 직전 DB 백업**(7.4절 백업 스크립트 `cbt-backup`). 백업이 실패하면 멈춘다. 백업 스크립트가 아직 없으면 **첫 배포(release가 하나도 없을 때)만** 건너뛴다 — 그때는 DB가 비어 있어 지킬 것이 없다
   3. 새 실행 파일로 `migrate up`(`cbt` 계정 · 운영 env). 실패하면 여기서 멈춘다 — 링크를 안 바꿨으니 옛 앱이 그대로 돈다
   4. `current` 링크 교체(새 링크를 만든 뒤 `mv`로 바꿔치기 — 링크가 없는 순간이 없다) → `systemctl restart cbt`
   5. `127.0.0.1:8080/api/health`가 200인지 30초까지 확인 — **실패하면 옛 링크로 되돌리고 재시작**, 실패로 끝낸다. 마이그레이션은 되돌리지 않는다(7.5절 규칙으로 옛 버전도 맞는다)
   6. 오래된 release 정리(지금 쓰는 것을 빼고 최근 5개 남김). 성공했을 때만
- 두 PC가 동시에 배포하면 서버 쪽 잠금(`flock`)으로 하나만 돈다. 나중 것은 바로 실패한다.
- 같은 커밋을 다시 배포하면 같은 폴더에 덮어쓰고 재시작한다(설정만 바꾼 뒤 다시 띄울 때).
- 열쇠 passphrase는 복사 · 실행에 한 번씩, sudo 비밀번호는 한 번 묻는다.

### 7.4 백업 · 복원
- 서버: systemd 타이머가 **하루 한 번** `pg_dump`(compose 안 Postgres) → **`age`로 암호화** → `/var/backups/cbt/`. 서버에는 **잠그는 열쇠(공개키)만** 있다.
- PC: `task backup:pull`이 최신 백업을 가져온다. **여는 열쇠(개인키)는 PC에만** — 만들자마자 비밀번호 관리자나 종이에 따로 보관한다. **잃으면 모든 백업을 못 연다.** 두 PC 모두 복원할 수 있게 같은 개인키를 둔다.
- 복원 연습: 백업을 풀어 로컬(또는 빈 DB)에 넣고 건수를 비교한다(PLAN M3 완료 기준).

### 7.5 운영 규칙
- **마이그레이션은 옛 버전과도 맞게 쓴다**: 컬럼 · 테이블 추가 위주. 지우기 · 이름 바꾸기는 새 버전이 안 쓰게 된 **다음 배포에**. 배포가 실패해 옛 버전으로 되돌려도 DB가 맞아야 한다.
- **첫 배포 순서는 DNS → Caddy**: `cbt` 레코드를 만들고 `nslookup cbt.sagming.com`으로 서버 IP가 나오는 걸 본 뒤 Caddy 사이트를 켠다. DNS 없이 켜면 인증서 발급이 계속 실패하고, Let's Encrypt가 실패가 반복된 이름을 한 시간 동안 막는다.
- **재부팅하면 저절로 살아나야 한다**: compose `restart: unless-stopped`, systemd `Restart=always`. 앱은 DB가 아직 안 떴으면 시작을 거부하고(1절 serve), systemd가 몇 초 뒤 다시 띄운다.
- **새벽 자동 재부팅은 아직 끈다**: 재부팅 뒤 자동 복구를 실제로 확인한 다음 켤지 정한다(그 전까지는 SERVER.md 5장 "재부팅이 필요한지"를 가끔 본다).
- 스왑 1GB(메모리 1GB에 네 프로세스), Docker 로그는 컨테이너마다 크기 제한.

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
| 2026-10-30 | M2 사전 검증: 1절 API mux 분리, 2절 `clientip/` · `db/dbgen/` · `testutil/` · handler 임베드, 3.1절 미들웨어 줄(본문 제한 · IP), 요청 제한은 handler에서, 기본 잠금을 openapi `security`로, 3.2절 429 `Retry-After` · 봉투 범위, 4절 쿠키 재발급 · 요청 제한 표 · Valkey 장애 · 만료 세션 정리 · 401 예외 · `next` 규칙 |
| 2026-10-30 | M2 구현 중: 1절 바깥 라우팅을 경로 앞부분으로(mux 패턴 충돌), 3.2절 413 · 오류 code 목록 |
| 2026-10-31 | 7절을 "빌드 · 배포 · 운영"으로 넓힘: 7.1 운영용 빌드(커밋 번호), 7.2 운영 구성(Caddy · systemd · compose · 서버 위치 · `cbtctl`), 7.3 배포 흐름, 7.4 백업(`age`, 개인키는 PC에만), 7.5 운영 규칙(옛 버전과 맞는 마이그레이션 · DNS 먼저 · 재부팅 자동 복구 · 자동 재부팅 보류). 사전 검증에서 시나리오를 돌려 찾은 것 반영 (사용자 결정: 결정 3개 추천안) |
| 2026-10-31 | 7.2절 Caddy: 인증서 등록 이메일을 넣지 않음(Let's Encrypt가 만료 알림 메일을 없앰 — `ACME_EMAIL` env 계획 철회), IP · 다른 이름 접속 처리 명시 |
| 2026-10-31 | 7.3절 서버 쪽 배포(`release.sh`) 세부: 커밋 · 실행 파일 일치 확인, 백업 스크립트가 없을 때는 첫 배포만 건너뜀, migrate 실패 시 링크를 안 바꿈, 링크 바꿔치기, 헬스체크 30초, 동시 배포 잠금, 같은 커밋 재배포 |
