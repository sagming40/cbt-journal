# 개발 일지 (v2)

> 겪은 일 · 막힌 일 · 그때 배운 것을 적는다. **최신이 위.** 지난 기록은 고치지 않고 위에 덧붙인다.
> 로드맵 · 진행 상황은 [PLAN.md](PLAN.md), 결정의 이유는 [decisions/](decisions/README.md). v1 일지는 [archive/DEVLOG_v1.md](archive/DEVLOG_v1.md).

## 현재 상태
| 항목 | 내용 |
|---|---|
| 단계 | M1 개발 환경 · 뼈대 — 회사 PC 쪽 전부 끝, 집 PC 확인만 남음 (PLAN 2장) |
| 끝난 것 | M0 전부. M1: 개념 설명, 회사 PC 도구 설치, v1 코드 삭제, `.gitattributes`, `compose.yaml`, Go 모듈 · `app serve` · `app migrate` · config · slog, `/api/health`(Postgres · Valkey 확인), README "실행", `api/openapi.yaml` + oapi-codegen strict 서버, SvelteKit SPA(shadcn-svelte · 토큰 테마 · openapi-fetch · health 화면), `go:embed` + SPA fallback(실행 파일 하나로 화면), Taskfile · air, golangci-lint, GitHub Actions(첫 실행 초록) |
| 진행 중 | 없음 |
| 다음에 할 일 | 집 PC에서 `task dev` 확인 → M1 완료 처리 → M2 인증 시작(개념 설명부터: 세션 · argon2id · sqlc) |
| 기기별로 할 일 | **집 PC**: README "실행 → 처음 한 번" 그대로 (도구 설치 Go · pnpm · Task · golangci-lint → `.env` 만들기(두 비밀번호 맞추기) → `docker compose up -d` → `task setup` → `task migrate`) → `task dev` · `task lint` 확인. VS Code `*.css` → tailwindcss 연결도 |

## 작성 규칙
- 세션(작업 단위)이 끝날 때 날짜 블록을 맨 위에 추가하고, 위 "현재 상태"를 고친다.
- 막혔던 일은 해결됐어도 남긴다. 형식: **증상 / 원인 / 해결 / 배운 것**.
- 구현 중에 설계를 바꿨으면 `[설계]` 항목으로 이유를 적고, 해당 설계 문서의 변경 이력도 고친다.
- 블록 끝에 **이번 세션 커밋**과 **다음에 할 일**을 꼭 쓴다.

---

## 2026-10-30 · M1 시작 ─ 도구 설치 · v1 삭제 · compose (회사 PC)

### 한 일
- 개념 설명: Go(패키지 · 에러 값 · 인터페이스), Valkey와 Redis의 관계, SvelteKit SPA · SPA fallback, goose.
- 회사 PC에 winget으로 Go 1.27.0 · pnpm 12.15.0 · Task 3.54.0 설치 (사용자 직접).
- 타이핑 학습 모드로 v1 코드 삭제 · `.gitignore` 재작성 · `.gitattributes` · `compose.yaml` · `.env.example` 작성.
- 확인: 두 컨테이너 healthy, PostgreSQL 18.6 · Valkey 9.1.3, 포트는 `127.0.0.1`에만 열림, AOF 켬 — TTL 600초 키가 Valkey 재시작 뒤에도 남음(ttl 595).
- 질문 정리: "Alembic을 쓸 수 있나" → 쓸 수는 있지만 Python 런타임이 필요해 실행 파일 하나 배포가 깨지고, sqlc가 `.py` 마이그레이션을 못 읽어 스키마 원본이 둘이 된다. goose 유지.
- 외부 API 후보 검토: 기록 내용을 보내지 않는 것만 고름 → REQUIREMENTS A10(유출 비밀번호 거절 — L로 적었다가 사용자 결정으로 M, PLAN M2에 추가) · X3(날씨) · 5.3절 메일 발송 API. 사고 기록 · 인지 왜곡 전용 공개 API는 찾아봤지만 쓸 만한 게 없다(있는 것은 기록 내용을 통째로 보내는 방식이라 5.1절 위반).

- 타이핑 학습 모드로 Go 첫 코드: `go.mod`, `internal/config`(필수 값 모아서 한 번에 거부, 에러에 비밀값 없음), `internal/server`(`GET /api/health`), `cmd/app`(서브커맨드 표 · slog · 우아한 종료). 확인: 테스트 config 4 · server 1 함수 통과, GET 200 / POST 405, Ctrl+C → "서버 종료" 로그.
- 타이핑 학습 모드로 goose · DB 연결: 의존성 pgx · go-redis · goose(사용자 `go get`), `internal/db`(embed 마이그레이션 · pgxpool), `internal/cache`, `app migrate up|down|status`, health가 확인별 ok/error · 503. 확인: migrate pending → applied → down → up, `goose_db_version` 2행(0 · 1), health 200 → Valkey stop 503 → start 200.
- 지금 쓰는 터미널 명령을 README "실행"에 정리했다(사용자 요청). Taskfile이 생기면 같이 고친다.
- 타이핑 학습 모드로 `api/openapi.yaml` 첫 판 + oapi-codegen(`go get -tool`로 go.mod에 고정, strict · std-http). 경로 등록을 생성 코드에 맡기고 `var _ apigen.StrictServerInterface = (*handler)(nil)`로 컴파일 시 확인. 확인: `operationId`를 바꿔 재생성하면 `go build`가 "missing method GetHealthz"로 깨짐 → 되돌리면 통과, server 테스트 4 케이스, curl 200 / 405.
- [설계] `openapi-typescript`는 이번 항목에서 빼고 SvelteKit 항목에서 연결: 아직 `web/`이 없어 설치할 곳이 없다. PLAN 항목에 표시.
- 안내한 `go get github.com/oapi-codegen/runtime`은 필요 없었다: health처럼 파라미터가 없는 경로만 있으면 생성 코드가 runtime을 import하지 않아서 `go mod tidy`가 지웠다. 경로 · 쿼리 파라미터가 생기는 M2 이후에 다시 들어온다.
- 타이핑 학습 모드로 SvelteKit SPA: `sv create`(minimal · TS · prettier · eslint · tailwindcss), adapter-static, shadcn-svelte 1.7(vega · lucide) Button · Card, `layout.css`를 DESIGN 6장 토큰으로, Inter → Pretendard(npm), mode-watcher, openapi-typescript + openapi-fetch, health 화면. 확인: `pnpm check` 0 · `pnpm lint` 통과 · `pnpm build` 성공, 브라우저에서 ok/ok → 라이트 ↔ 다크 → Valkey stop 시 error(아이콘 + 굵은 글자) → Go 종료 시 "연결이 안 돼" → 다시 켜면 ok/ok.
- [설계] 카드 모서리가 DESIGN의 12px이 아니라 14px: shadcn 1.7은 모서리를 곱셈(`md = radius × 0.8`, `xl = × 1.4`)으로 계산한다. `--radius` 10px로 Button(`rounded-md`) 8px을 정확히 맞추고, Card(`rounded-xl`)는 14px로 둔다. 화면을 만들 때 감싸는 부품에서 맞출지 본다(DESIGN 6장 "모양"은 그대로).
- `openapi-typescript` 7.13이 TypeScript `^5`를 피어로 원하는데 프로젝트는 TS 6(`sv` 기본). 생성 도구일 뿐이고 `pnpm gen:api` 결과가 정상이라 그대로 둔다. TS를 낮추면 SvelteKit 3 쪽이 어긋날 위험이 더 크다.
- 타이핑 학습 모드로 `go:embed` + SPA fallback: `internal/webdist`(`all:dist`, `.gitkeep`), `GET /` SPA 핸들러, 오류 봉투 첫 판. 확인: TestSPA 11 케이스, `bin/app.exe serve` 하나로 8080에서 화면 ok/ok · `/checkin`은 서버 200 + SvelteKit 404 화면, `/api/nope` JSON 404, 새로 checkout한 트리(dist에 `.gitkeep`만)에서 `go build` · `go vet` 통과.
- [설계] SPA를 메서드 없는 `/`가 아니라 `GET /`로 등록: `/`로 열면 Go mux가 `POST /api/health`를 경로가 맞는 `/`로 보내 405가 404로 바뀐다. 대가로 없는 `/api/*`에 GET이 아닌 메서드는 405. ARCHITECTURE 1절.
- `.env`는 Go가 읽지 않는다. Taskfile을 만들기 전까지는 PowerShell 한 줄로 터미널에 불러왔고, 이제 Task의 `dotenv`가 읽는다.
- 타이핑 학습 모드로 Taskfile · air: `setup` · `dev`(air + Vite 동시, `output: prefixed`) · `gen` · `test` · `lint` · `build`(빌드 → embed 복사 → `bin/app`) · `serve` · `migrate`. 확인: 환경변수를 지운 터미널에서 `task migrate -- status` 성공, `task build` 뒤 `git status` 깨끗, `task serve` health 200, `task dev` 저장 시 재빌드 · Ctrl+C로 오류 없이 종료. README "실행"을 `task` 중심으로 다시 썼다.
- 타이핑 학습 모드로 golangci-lint(v2.15.0, winget) + GitHub Actions. 첫 실행에서 지적 3개: 테스트의 가짜 비밀번호 G101(테스트 파일의 G101만 제외), 테스트 헬퍼 `httptest.NewRequest`의 noctx(`NewRequestWithContext`로). 확인: `task lint` 0 issues, Actions 첫 실행(run 41001543315) Go · 프론트 · 빌드 세 작업 모두 success — Taskfile의 Linux용 embed 명령과 `.gitkeep` 전략이 깨끗한 Linux에서 처음 검증됐다.
- golangci-lint는 `go get -tool` 대신 winget: 공식 문서가 비권장(검사기 의존성이 `go.mod`와 부딪힐 수 있음). 대신 PC와 `ci.yml`의 버전을 같은 번호로 맞춘다(CLAUDE.md "자주 쓰는 명령").

### 막혔던 일
**`git add --renormalize .` 뒤에도 `.gitattributes`가 Untracked**
- 증상: renormalize를 했는데 `.gitattributes`만 stage되지 않았다.
- 원인: `--renormalize`는 이미 추적 중인 파일만 다시 검사한다. 새 파일은 대상이 아니다.
- 해결: `git add .gitattributes`를 따로. 규칙 자체는 작업 폴더에 파일만 있어도 적용돼 renormalize 결과는 맞았다(바뀐 파일 0개).
- 배운 것: 규칙 파일은 "적용"과 "커밋"이 별개다.

**`psql` 결과 뒤에 `(END)`가 떠서 멈춘 것처럼 보임**
- 원인: `docker compose exec`가 터미널을 붙여 주면 psql이 결과를 페이저(less)로 보여 준다.
- 해결: `q`로 나온다. 한 줄 확인이면 `-tAc` 옵션으로 페이저 없이 값만 받는다.

**`migrate`가 `password authentication failed` (28P01)**
- 증상: `app migrate` · `serve`가 모두 비밀번호 인증 실패. 그런데 `docker compose exec postgres psql`은 잘 된다.
- 원인: `.env`의 `POSTGRES_PASSWORD`는 바꿨는데, 앱 설정 줄을 `.env.example`에서 복사하면서 `DATABASE_URL` 안의 비밀번호는 예시값으로 남았다. 컨테이너 안에서 붙는 psql은 Postgres 이미지가 비밀번호 없이 믿어 줘서(trust) 차이가 안 드러났다.
- 해결: `DATABASE_URL`의 비밀번호를 `POSTGRES_PASSWORD`와 맞추고 `.env`를 다시 불러왔다. 값은 대화에 꺼내지 않고 "같은가"만 비교해서 찾았다.
- 배운 것: 같은 비밀번호가 `.env`에 두 번 들어간다(compose용 · 앱용). README "처음 한 번"에 둘을 맞추라고 적었다. pgx 에러에는 비밀번호가 안 나온다(사용자 · DB 이름만).

**`localhost`가 `[::1]`부터 시도해서 매번 한 번 거절됨**
- 원인: Windows는 `localhost`를 IPv6(`::1`)부터 푸는데 compose는 포트를 `127.0.0.1`에만 열었다.
- 해결: `.env.example`의 `DATABASE_URL` · `VALKEY_URL`을 `127.0.0.1`로.

**안내한 SvelteKit 구조가 실제와 달랐다 (SvelteKit 3)**
- 증상: `svelte.config.js`가 없고, 전역 CSS가 `src/app.css`가 아니라 `src/routes/layout.css`, 별칭이 `$lib`가 아니라 `#lib`.
- 원인: `sv create`가 SvelteKit 3을 설치했다. adapter 설정이 `vite.config.ts`로 옮겨 왔고, 별칭은 Node 표준 subpath imports(`package.json`의 `imports`)로 바뀌었다.
- 해결: 생성 직후 파일을 열어 보고 안내를 고쳤다. shadcn-svelte 1.7 init은 `layout.css` · `#lib`를 스스로 찾았다. import는 확장자까지 적는다(`#lib/...js`).
- 배운 것: 처음 쓰는 도구는 생성기를 돌린 뒤 결과를 먼저 열어 보고 다음 안내를 쓴다. CLAUDE.md "함정"에 적었다.

**`pnpm lint`가 생성된 shadcn Button에서 실패**
- 증상: `ui/button/button.svelte` 63:3 `svelte/no-navigation-without-resolve`.
- 원인: 범용 Button은 `href`를 그대로 쓴다. 링크를 `resolve()`로 만드는 건 쓰는 쪽의 일이다.
- 해결: `eslint.config.js`의 전역 `ignores`에 `ui/` · `schema.d.ts`를 넣었다(`.prettierignore`와 같은 목록). 우리 화면 코드에는 규칙이 그대로 걸린다.

**Tailwind 확장이 있는데도 `layout.css`에 `Unknown at rule` 경고**
- 원인: `sv`가 만든 `web/.vscode/settings.json`(`*.css` → tailwindcss)은 VS Code로 `web` 폴더를 직접 열 때만 읽힌다. 저장소 맨 위를 열어 두면 안 읽힌다.
- 해결: 사용자 설정 또는 저장소 맨 위 `.vscode/settings.json`(커밋 안 됨)에 같은 줄. README "자주 막히는 것"에 적었다.

**`/../../go.mod` 테스트가 200이 아니라 307**
- 원인: Go `ServeMux`는 `..`이 섞인 경로를 핸들러까지 보내지 않고 정리된 주소로 307 리다이렉트한다. 안내한 테스트가 이 동작을 빠뜨리고 200을 기대했다.
- 해결: 307 + `Location: /go.mod`를 확인하는 테스트로 바꿨다. 경로 조작은 mux와 `path.Clean`이 두 겹으로 막는다.

**`task dev`에서 air가 `exclude_dir`의 `internal/webdist/dist`를 무시함**
- 원인: air는 폴더 경로를 운영체제 구분자(Windows는 `\`)로 만들어 `exclude_dir`와 글자 그대로 비교한다(air 소스 `isExcludeDir`). 한 단계 이름(`web` · `tmp`)만 맞았다.
- 해결: `"internal\\webdist\\dist"`(TOML에서 `\`는 `\\`). 같은 때 `build.bin` 지원 중단 경고 → `entrypoint = ["./tmp/app.exe", "serve"]`.

**`task dev`를 Ctrl+C로 끄면 깨진 Y/N 질문 + `exit status 58`**
- 원인 1: `pnpm dev`가 Windows에서 `vite.cmd`(배치 파일)를 거쳐, cmd가 "일괄 작업을 끝내시겠습니까 (Y/N)?"를 물었다(CP949라 Task 출력에서 깨짐).
- 원인 2: 배치를 건너뛰어도 Ctrl+C로 끝난 node의 종료 코드 `0xC000013A`의 마지막 바이트 `0x3A` = 58을 Task가 실패로 읽었다.
- 해결: `dev:web`은 `node node_modules/vite/bin/vite.js dev`로 직접 띄우고 `ignore_error: true`. air 쪽 `Process Exit with Code: 1`(재시작 때 강제 종료)은 Windows에서 정상.

**`webdist.go`에 staticcheck SA9009 경고**
- 원인: 설명 주석이 `// go:embed는 …`으로 시작해서 "공백 때문에 무시된 지시문"으로 보였다(안내한 주석 문장 탓).
- 해결: `// embed 지시문은 …`으로. 다른 `.go` 파일에 `// go:` 같은 주석이 없는지 검색해 진짜 지시문 3개(`//go:embed` 2, 생성 코드 `//go:build` 1)만 있는 걸 확인했다.

**참고: go-redis가 자기 로그를 따로 찍는다**
- Valkey를 끄면 `redis: ... pool.go:770: ... failed to dial` 줄이 slog 형식이 아니라 그대로 나온다. go-redis 내부 로거라서다. 기록 내용은 없어서 지금은 그대로 두고, 운영 로그를 JSON으로 모을 때(M3) 정리할지 본다.

### 이번 세션 커밋
`2d629e5` v1 삭제 · .gitignore / `05e2870` .gitattributes / `276f6af` compose · .env.example / `4494d7a` 문서 갱신 / `2df9652` · `15fcd2f` 외부 API 후보 · A10을 M2로 / `b3df9b8` Go 뼈대 / `5107443` 문서 / `84d4481` goose · health DB 확인 / `6600b4b` README "실행" · 문서 / `33b5e79` openapi · oapi-codegen / `d1d33b2` 문서 / `84646b0` SvelteKit SPA / `73af1fb` 문서 / `93d0e2f` go:embed · SPA fallback / `2517d5b` 문서 / `58b63e4` Taskfile · air / `bb9e918` 주석 SA9009 / `a7bf28b` 문서 / `7b59c92` golangci-lint / `82cafa6` GitHub Actions / (이 커밋) 문서 갱신

### 다음에 할 일
- 집 PC: README "처음 한 번" → `task dev` · `task lint` (M1 완료 기준의 마지막)
- M2 인증: 개념 설명(세션 쿠키 · argon2id · sqlc)부터
- 집 PC: 위 "기기별로 할 일"

---

## 2026-10-29 · 문서 전체 검토 (M0)

### 한 일
- 문서를 다 쓴 뒤 한꺼번에 다시 읽고 서로 대조했다. 불일치 8 · 설계 구멍 6 · 보완 제안 7을 찾아, 보완 1개(`__Host-` 쿠키 접두사)만 빼고 반영했다. 그대로 둔 것: "7칸 / 8단계" 용어, 수정 모드 자동 저장 없음, 계정별 로그인 제한으로 남이 15분 잠글 수 있는 점.

### [설계] 첫 배포를 M6 → M3(인증 바로 뒤)으로
- M4 · M5(옛 M3 · M4)의 완료 기준이 "휴대폰으로 10초 체크인", "휴대폰 → PC 이어 쓰기"인데 배포는 M6이었다. 개발 서버에 휴대폰을 붙이면 `Secure` 쿠키가 `http://IP주소`에서 저장되지 않고, 회사 Wi-Fi는 기기 간 통신을 막는 경우가 많다.
- 인터넷에 여는 건 인증이 있어야 안전하므로 M2 바로 뒤에 빈 앱으로 배포하고, 이후 단계마다 올려서 확인한다. 실제 기록을 쓰기 시작하는 건 그대로 M6. 대안이었던 "개발 중 터널"은 개발 서버를 인터넷에 여는 거라 매번 켜고 끄는 관리가 필요해서 뺐다 (사용자 결정).

### [설계] 검토에서 고친 설계 구멍
- **비밀번호를 바꿔도 다른 세션이 최대 10분 살아 있었다**: DB 세션만 지우고 Valkey 세션 캐시(TTL 10분)는 안 지우는 설계였다. 세션을 지우는 모든 경로(로그아웃 · 비밀번호 변경 · 다른 기기 로그아웃 · 탈퇴)에서 `RETURNING token_hash`로 캐시도 지운다.
- **프록시 뒤 IP**: 배포하면 요청 제한이 모든 사용자를 프록시 IP 하나로 본다 → `TRUSTED_PROXIES` 규칙(ARCHITECTURE 3.3절).
- **로그인 화면이 가입 허용 여부를 알 방법이 없었다** → 공개 설정 `GET /api/config`.
- **D2 "평균 감소율"**: 전 강도 0이면 나눗셈이 안 되고, "감소"는 평가로 읽힌다 → 평균 변화(점). D1 추이는 사고 기록의 "전" 강도만.
- **회사 PC에서 30일 로그인 유지**: 감정 기록을 자리 비운 사이 볼 수 있다 → "로그인 유지" 선택(기본 해제, 해제 시 서버 기한 12시간) + 다른 기기 모두 로그아웃.

### 이번 세션 커밋
(이 커밋) 문서 검토 반영

### 다음에 할 일
- M1 시작: 개념 설명(Go 기초 · Valkey · SvelteKit SPA) → 도구 설치(사용자 직접, 두 PC)

---

## 2026-10-29 · v2 재설계 (M0)

### 한 일
- v1 코드 · 문서를 다시 읽고 문제점 정리 → 전면 재작성 결정 (0001). v1 코드는 태그 `v1-legacy`, 문서는 `archive/`.
- 스택 결정: Go (0002) · SvelteKit SPA + shadcn-svelte (0003) · PostgreSQL (0004) · Valkey (0005), AI 코치 보류 (0006).
- 문서 구조를 새로 잡고 REQUIREMENTS · ARCHITECTURE · DATA_MODEL · DESIGN · PLAN 작성.
- 목업 4장(컬러 시안 · 체크인 · 사고 기록 모바일/데스크톱), 색은 B · 차분한 남색으로 확정.

### [설계] 사용 기기 기준을 "붙잡기는 휴대폰, 정리는 어디서든"으로
- 처음 문서에는 "대부분 휴대폰에서 쓴다"를 확인 없이 가정해서 적었다. 실제로는 회사 PC 앞에 있는 시간이 길다.
- 사고 기록을 성격으로 나눴다: 1 ~ 3단계(상황 · 감정 · 자동 사고)는 떠오른 순간에 붙잡아야 해서 휴대폰, 4 ~ 8단계는 가라앉은 뒤 따져 보는 게 잘 돼서 PC · 휴대폰 둘 다.
- 이미 설계한 서버 쪽 초안 저장(Valkey)으로 기기를 옮겨 이어 쓰기가 그대로 된다 → 새로 만들 것은 [나중에 이어 쓰기] 버튼과 데스크톱 "앞에서 쓴 답" 패널 정도.

### 막혔던 일
**목업이 다크 화면에서 글자가 검게 나옴**
- 증상: 컬러 시안의 다크 절반에서 본문 글자가 배경과 같은 검은색.
- 원인: `color: var(--text)`를 `body`에만 걸었다. CSS 변수는 화면 틀(`.half`)마다 다르게 심었지만, `color`는 `body`에서 **이미 계산된 값**(라이트 색)으로 상속된다.
- 해결: 테마를 심는 틀(`[data-theme]`)에 `color: var(--text)`를 다시 건다.
- 배운 것: 변수를 요소마다 바꿔 끼우는 구조에서는, 그 변수를 쓰는 속성도 그 요소에서 다시 선언해야 한다. 실제 앱도 테마 클래스를 `<html>`에 거는 방식이면 문제없지만, 한 화면에 두 테마를 섞을 땐 같은 함정이 있다.

**`npx http-server`가 설치를 묻는다**
- 증상: 목업을 보려고 `npx http-server`를 쳤더니 "Need to install the following packages"가 나왔다.
- 원인: `npx`는 없는 패키지를 npm 캐시에 임시로 받아 실행한다. 프로젝트 · 전역에 설치되는 것은 아니다.
- 해결: 받아도 되고, conda 환경이면 `python -m http.server 5199`로 받지 않고 띄운다. 목업 README에 둘 다 적음.

### 이번 세션 커밋
`387c9cd` 결정 기록 · REQUIREMENTS / `3c6debe` ARCHITECTURE / `5103e58` DATA_MODEL / `95779fd` DESIGN / `d677587` 사용 기기 기준 / `13ba562` 목업 / `0dc20c5` 색 확정 / `998e3d6` PLAN · DEVLOG / `1122489` CLAUDE.md / (이 커밋) README

### 다음에 할 일
- M1: 개념 설명(Go 기초 · Valkey · SvelteKit SPA) → 도구 설치(사용자 직접)
