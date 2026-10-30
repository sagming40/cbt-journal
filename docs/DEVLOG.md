# 개발 일지 (v2)

> 겪은 일 · 막힌 일 · 그때 배운 것을 적는다. **최신이 위.** 지난 기록은 고치지 않고 위에 덧붙인다.
> 로드맵 · 진행 상황은 [PLAN.md](PLAN.md), 결정의 이유는 [decisions/](decisions/README.md). v1 일지는 [archive/DEVLOG_v1.md](archive/DEVLOG_v1.md).

## 현재 상태
| 항목 | 내용 |
|---|---|
| 단계 | M1 개발 환경 · 뼈대 진행 중 (PLAN 2장) |
| 끝난 것 | M0 전부. M1: 개념 설명, 회사 PC 도구 설치, v1 코드 삭제, `.gitattributes`, `compose.yaml`, Go 모듈 · `app serve` · `app migrate` · config · slog, `/api/health`(Postgres · Valkey 확인), README "실행", `api/openapi.yaml` + oapi-codegen strict 서버 |
| 진행 중 | 없음 |
| 다음에 할 일 | SvelteKit + `adapter-static` + Tailwind + shadcn-svelte, DESIGN 6장 토큰 테마, `openapi-typescript` 연결 (PLAN M1 여덟째 항목 + 일곱째의 남은 반) |
| 기기별로 할 일 | **집 PC**: README "실행 → 처음 한 번" 그대로 (도구 설치 → `.env` 만들기(두 비밀번호 맞추기) → `docker compose up -d` → `go mod download` → `migrate up`) |

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
- `.env`는 Go가 읽지 않는다. Taskfile을 만들기 전까지는 PowerShell 한 줄로 터미널에 불러온다(Taskfile에서 `dotenv`로 대체 예정).

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

**참고: go-redis가 자기 로그를 따로 찍는다**
- Valkey를 끄면 `redis: ... pool.go:770: ... failed to dial` 줄이 slog 형식이 아니라 그대로 나온다. go-redis 내부 로거라서다. 기록 내용은 없어서 지금은 그대로 두고, 운영 로그를 JSON으로 모을 때(M3) 정리할지 본다.

### 이번 세션 커밋
`2d629e5` v1 삭제 · .gitignore / `05e2870` .gitattributes / `276f6af` compose · .env.example / `4494d7a` 문서 갱신 / `2df9652` · `15fcd2f` 외부 API 후보 · A10을 M2로 / `b3df9b8` Go 뼈대 / `5107443` 문서 / `84d4481` goose · health DB 확인 / `6600b4b` README "실행" · 문서 / `33b5e79` openapi · oapi-codegen / (이 커밋) 문서 갱신

### 다음에 할 일
- SvelteKit + `adapter-static` + Tailwind + shadcn-svelte, DESIGN 6장 토큰 테마 · 라이트/다크, `openapi-typescript` 연결 (개념 설명부터)
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
