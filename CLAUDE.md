# CLAUDE.md

이 저장소에서 일하는 Claude가 지키는 공통 규칙이다.
세션끼리는 기억을 공유하지 않고 기기 · 계정도 둘이라서(아래 "기기 둘") **저장소가 공용 노트**다. 이 파일과 `docs/`가 기억 역할을 한다. 기기와 상관없이 지킬 규칙은 Claude 메모리가 아니라 전부 이 파일에 적는다.

CBT 저널 v2: 인지행동치료의 사고 기록을 매일 짧게 연습하는 웹앱. 1인 취미 프로젝트다(혼자 쓰기로 시작, 나중에 공개 가능).
Go(`net/http` · pgx · sqlc · goose) + PostgreSQL + Valkey(asynq) / SvelteKit SPA(Svelte 5 · TypeScript · shadcn-svelte · Tailwind), 실행 파일 하나로 배포.

## 문서 지도 (질문 하나 = 문서 하나)
| 질문 | 문서 |
|---|---|
| 지금 어디까지 왔고 바로 다음에 뭘 해? | [`docs/DEVLOG.md`](docs/DEVLOG.md) "현재 상태" |
| 무엇을 어떤 순서로 만들어? 완료 기준은? | [`docs/PLAN.md`](docs/PLAN.md) |
| 기능 범위는? 하지 않을 것은? | [`docs/REQUIREMENTS.md`](docs/REQUIREMENTS.md) (기능 번호는 4장) |
| 코드가 어떻게 나뉘고 요청이 어떻게 흘러? 인증은? | [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) |
| 무엇을 어디에 저장하고 왜? (Postgres · Valkey) | [`docs/DATA_MODEL.md`](docs/DATA_MODEL.md) |
| 화면은 어떻게 생겼고 왜? 색은? | [`docs/DESIGN.md`](docs/DESIGN.md) (색 토큰은 6장), 그림은 [`docs/mockups/`](docs/mockups/README.md) |
| 왜 이 스택 · 이 방식이야? | [`docs/decisions/`](docs/decisions/README.md) |
| API 모양은? | `api/openapi.yaml` (M1부터. 마크다운 API 명세는 따로 쓰지 않는다) |
| 겪은 문제와 해결은? | [`docs/DEVLOG.md`](docs/DEVLOG.md) 날짜 기록 (최신이 위) |
| 이게 뭐고 어떻게 실행해? | [`README.md`](README.md) |
| v1은? | [`docs/archive/`](docs/archive/README.md), 코드는 태그 `v1-legacy` |

**문서 규칙**
- 같은 내용을 두 곳에 쓰지 않는다. 필요하면 링크나 "ARCHITECTURE 4절"처럼 절 번호로 가리킨다.
- 설계가 바뀌면 **설계 문서를 코드보다 먼저** 고친다. 구현 중에 바꿨으면 그 문서 변경 이력 + DEVLOG `[설계]` 항목에 이유를 남긴다.
- 결정을 새로 하거나 뒤집으면 `docs/decisions/`에 새 파일. 지난 결정 파일은 상태만 `대체됨 → NNNN`으로 바꾼다.
- 스키마 컬럼의 원본은 마이그레이션 파일이다. DATA_MODEL에 컬럼 표를 옮겨 적지 않는다.
- 문서의 장 · 절 번호를 바꾸면 그 번호를 부르는 자리를 검색해서 같이 고친다. 오기를 고칠 때도 다른 문서에 같은 오기가 있는지 검색한다.
- 새 기능 아이디어는 바로 구현하지 않고 REQUIREMENTS에 우선순위(L)로 적는다.

## 시작할 때
```powershell
git status              # 커밋 안 된 변경이 있으면 먼저 사용자에게 알린다
git switch main
git pull --ff-only      # 다른 기기에서 push한 것 받기. 실패하면 멈추고 원인 보고
```
받은 커밋에 따라 (M1 이후):
- `go.mod` · `go.sum`이 바뀜 → `go mod download`
- `web/pnpm-lock.yaml`이 바뀜 → `web`에서 `pnpm install --frozen-lockfile`
- `internal/db/migrations/`가 바뀜 → `task migrate`

그다음 DEVLOG "현재 상태"와 최신 블록의 "다음에 할 일"을 읽고, 하려는 일에 맞는 문서를 위 지도에서 골라 본다.

## 끝날 때
- DEVLOG: 맨 위에 날짜 블록 추가, "현재 상태" 표 갱신. 형식은 DEVLOG 위쪽 "작성 규칙"을 따른다. 막혔던 일은 해결됐어도 남긴다.
- PLAN: 끝낸 항목 체크, 2장 진행 현황 표 갱신. 계획에 없던 작업은 `*(계획에 없던 작업)*`을 붙여 추가한다.
- 커밋하고 **push까지** 한다. push 안 하면 다른 기기에서 이어갈 수 없다.
- 세션이 중간에 끊길 것 같으면 `wip:` 커밋으로 push해 둔다. 다음 세션에서 정식 커밋으로 정리할지는 사용자에게 묻는다(`main`이라 고쳐 쓰려면 force push가 필요해서).

## 작업 방식
- **기본은 자동**: Claude가 코드를 쓰고, 실행 · 검증하고, 커밋 · push까지 한다.
- **사용자가 직접 하는 것** (Claude는 명령 · 코드와 "왜 필요한지"를 보여 주고, 사용자가 친 뒤 결과를 확인한다):
  - 사용자가 "직접 칠게"라고 할 때 — **타이핑 학습 모드**. Claude는 파일에 쓰지 않고, 비유를 곁들인 주석을 단 코드를 보여 준 뒤 "코드 검수!"로 보여 주면 오타 · 버그를 리뷰한다. 수정 지시는 "파일 위치 + 지금 코드 + 바꿀 코드"로.
  - 도구 설치(Go, pnpm, Task 등 PC에 까는 것)와 새 의존성 추가(`go get`, `pnpm add`) — 처음 쓰는 스택이라 무엇이 들어오는지 사용자가 직접 보고 친다.
  - 그 밖에 사용자가 "이건 내가 쳐 볼게"라고 한 명령.
- 학습 포인트가 큰 새 기능은 시작 전에 "내가 바로 할까, 직접 칠래?"를 한 줄로 물어봐도 된다. 이미 정해 줬으면 묻지 않는다.
- **새 도구 · 개념이 처음 나오는 단계는 설명부터** (PLAN 1장). 설명은 짧게, 일상 비유를 섞어서, 이 프로젝트에서 어디에 쓰는지까지.
- "검수만" 요청이면 검수만 하고 다음으로 넘어가지 않는다.

**구현 전 사전 검증 (자동 작업일 때 필수)**
1. 설계와 진행 순서를 정한다.
2. 구현을 마쳤다고 가정하고 머릿속으로 끝까지 돌려 본다: 정상 · 실패 · 엣지 케이스(빈 값 · 0건 · 범위 밖 · 비로그인 · **남의 데이터** · 시간대 경계 · 두 기기 동시 편집)를 하나씩 따라가며 문서 · 규칙과 대조한다. 구멍을 찾으면 설계를 고치고 다시 돌린다.
3. 설계한 순서대로 구현하고, 2에서 따라간 시나리오를 테스트로 옮긴다.
4. 보고: 무엇을 어떻게 구현했는지, 사전 검증에서 찾은 것 · 바꾼 것, 검증 결과(관찰값).

**검증**: "확인함"이 아니라 관찰값(테스트 건수, 응답 코드 · 본문, 화면에서 본 것)을 남긴다. 확인 못 한 것은 커밋 본문과 답변에 그렇게 적는다. 사전 검증은 실제 실행 확인을 대신하지 않는다.

## 브랜치 · 커밋 · push
- **`main` 하나만 쓴다.** 브랜치 · PR을 만들지 않는다.
- force push, `reset --hard`, `rebase` 같은 히스토리 재작성은 사용자가 명시적으로 요청할 때만.
- 검증이 실패했거나 끝나지 않은 작업은 커밋하지 않고 상태를 알린다(`wip:` 예외는 위 "끝날 때").
- **분리 기준: "따로 되돌리고 싶을 수 있는가"** — 성격이 다르면(feat / fix / docs) 나눈다. 백엔드 · 프론트를 같이 고쳤으면 나눈다. 문서 커밋은 코드 커밋과 따로.
- 메시지는 **Summary + Description**:
  ```text
  feat(M2): 세션 쿠키 로그인

  - sessions 테이블에 토큰 SHA-256 해시만 저장 (DATA_MODEL 3.2절)
  - Valkey 세션 캐시 10분, 로그아웃 시 키 삭제
  - 통합 테스트 8개 통과 (go test ./internal/auth/...)
  ```
  - Summary: `종류(M{n}): 한국어 요약`. 종류는 `feat` · `fix` · `refactor` · `test` · `docs` · `chore` · `ci`. `{n}`은 지금 단계(PLAN 2장). M0 문서 작업은 `docs:`.
  - Description: `- ` 목록으로 무엇을 · 왜, 확인한 결과. 동작 변경이 없으면 그렇게 적는다.

## 업로드 금지 · 데이터 취급
- 이 프로젝트 내용은 **Artifact, Claude Docs, 클라우드 세션, 원격 실행 · 예약 작업, 그 밖의 외부 서비스(Notion · Google Drive 등)에 올리지 않는다.** 결과물은 로컬 파일 또는 답변으로만 준다. 목업도 로컬 HTML로만 만든다.
- 밖으로 나가는 건 사용자 개인 GitHub 저장소(`sagming40/cbt-journal`) push뿐이다.
- **실제 기록 데이터**(배포 뒤 쌓이는 감정 · 생각 · 메모)를 대화 · 커밋 · 로그 · 테스트 · 목업에 넣지 않는다. 운영 DB를 조회해야 하면 건수 · id처럼 내용 없는 값만 본다. 예시 데이터는 지어낸다.
- 비밀값(`.env`, VAPID 키, DB 비밀번호)은 커밋 · 대화 · 로그에 남기지 않는다.

## 코드 규칙 (어기지 말 것)
- **사용자 데이터 SQL에는 전부 `user_id` 조건.** 남의 것은 404(없는 것과 같은 응답). 새 API마다 "남의 데이터 404" 통합 테스트를 붙인다. (ARCHITECTURE 2절)
- **기본 잠금**: 인증 없이 열리는 경로는 로그인 · 가입 · 헬스체크만, 명시적으로 연다. (ARCHITECTURE 3.1절)
- **로그에 기록 내용을 남기지 않는다.** id · 경로 · 상태 코드 · 걸린 시간만. 요청 본문 로깅 금지. (REQUIREMENTS 5.1절)
- **Postgres가 진실의 원천, Valkey는 휘발성.** Valkey가 날아가도 잃는 건 초안뿐이어야 한다. TTL 없는 키를 만들지 않는다(`asynq:*` 제외). 키를 새로 만들면 DATA_MODEL 4.2절 표에 먼저 추가한다. (decisions 0005)
- **생성 코드는 고치지 않는다**: `internal/apigen/`, sqlc 결과, `web/src/lib/api/` 생성 타입, `web/src/lib/components/ui/`. 원본(`openapi.yaml` · `queries/*.sql`)을 고치고 `task gen`. 생성 결과도 커밋한다.
- **시각은 UTC로 저장**, "하루"는 사용자 시간대(`users.timezone`)로 계산한다.
- **색은 DESIGN 6장 토큰으로만.** 임의 hex 금지. 빨강 · 초록으로 감정 · 강도 · 변화를 표시하지 않는다. 오류도 빨강이 아니라 아이콘 + 글자.
- **주석**: 한국어로, "왜 이렇게 했나"를 쓰고 필요하면 일상 비유를 곁들인다. 문서를 근거로 삼은 줄에는 `(DATA_MODEL 4.3절)`처럼 출처를 단다.
- 백엔드 동작을 바꾸면 테스트도 같이 바꾼다.

## 기기 둘
| 기기 | Claude 계정 | 비고 |
|---|---|---|
| **회사 PC** (주 작업 기기) | 회사 계정 | Docker Desktop(WSL2), PostgreSQL 18 직접 설치(5432 사용 중), Node 24, Miniconda(기본 Python 없음) |
| 집 PC | 개인 계정 | Docker, PostgreSQL 18 직접 설치(5432 사용 중), Node 24 |

- 코드는 GitHub로만 주고받는다. `.env`, `node_modules/`, DB · Valkey 데이터는 기기마다 따로다(`.env.example` 참고).
- DB 구조는 goose로 맞춘다: `git pull` 뒤 `task migrate`. 마이그레이션 파일은 **스키마를 고친 기기에서만** 만들고, 이미 push한 마이그레이션은 고치지 않는다(새 파일로).
- 도구 · 의존성을 추가했으면 다른 기기에서 할 일을 DEVLOG "현재 상태"의 "기기별로 할 일"에 적는다.

## 자주 쓰는 명령
포트: Postgres `5436` · Valkey `6380` · API `8080` · 프론트 dev `5176` (ARCHITECTURE 6절)

```powershell
docker compose up -d     # Postgres · Valkey
task setup               # Go · 프론트 의존성 (처음 · pull 뒤)
task dev                 # API(air, 8080) + 프론트 dev 서버(5176)
task gen                 # openapi.yaml → Go · TS 생성 코드 (sqlc는 M2에서 추가)
task test                # Go 테스트 (Vitest는 프론트 테스트가 생기면 추가)
task lint                # go vet + svelte-check · Prettier · ESLint (golangci-lint는 M1 다음 항목)
task migrate             # goose up (task migrate -- status | down)
task build               # 프론트 빌드 → embed 복사 → bin/app 하나
task serve               # bin/app serve
```
> `task`는 `.env`를 dotenv로 읽는다. `go run`으로 직접 실행하면 `.env`가 안 읽힌다. 명령 · 자주 막히는 것의 사람용 설명은 README "실행". Taskfile을 고치면 이 절과 README "실행"을 같이 고친다.

## 함정 (고치기 전에 확인)
- **5432는 이미 차 있다**: 두 PC 모두 PostgreSQL을 직접 설치해 둬서 기본 포트를 쓴다. compose Postgres는 `5436`. 접속 문자열 포트를 헷갈리면 엉뚱한 DB(직접 설치한 쪽)에 붙는다.
- **줄바꿈은 LF**: `.gitattributes`가 모든 글자 파일을 `eol=lf`로 정한다(PC의 `core.autocrlf`보다 우선). 에디터가 새 파일을 CRLF로 저장해도 커밋할 때 LF로 들어간다. `*.ps1` · `*.cmd` · `*.bat`만 CRLF.
- **Postgres 18 이미지 볼륨은 `/var/lib/postgresql`**: 17까지처럼 `.../data`에 묶는 예전 자료를 따라 하지 않는다(18 이미지의 `PGDATA`는 `/var/lib/postgresql/18/docker`라서 데이터가 볼륨 밖에 쌓이거나 시작이 실패한다).
- **SvelteKit 3 구조는 예전 자료와 다르다**: `svelte.config.js`가 없고 adapter는 `web/vite.config.ts`의 `sveltekit({ adapter })` 안에 있다. 별칭은 `$lib`가 아니라 `#lib`(package.json `imports`)이고 import에 확장자를 붙인다(`#lib/api/client.js`, `#lib/components/ui/button/index.js`). 전역 CSS는 `src/routes/layout.css`.
- **생성 코드는 포매터 · 린터에서도 뺀다**: `web/.prettierignore`와 `web/eslint.config.js`의 `ignores`에 `src/lib/components/ui/` · `src/lib/api/schema.d.ts`. 새 생성 위치가 생기면 두 곳에 같이 추가한다.
- **`npx`는 패키지를 임시로 받는다**: 없는 패키지면 설치를 묻는다(프로젝트에는 안 생김). 사용자에게 미리 말하고 쓴다.
- **회사 PC에는 기본 Python이 없다**: conda 환경이 켜진 터미널에서만 `python`이 된다. 이 프로젝트 빌드 · 실행에는 Python이 필요 없다.
- **휴대폰 확인은 배포 서버로**: 개발 서버(`http://IP주소`)에서는 `Secure` 쿠키가 저장되지 않아 휴대폰 로그인이 안 된다. 휴대폰이 필요한 확인은 M3(첫 배포) 이후 배포 서버에서 한다 (ARCHITECTURE 6절).
- **세션을 지우면 Valkey 캐시도 지운다**: 빠뜨리면 로그아웃 · 비밀번호 변경 뒤에도 옛 세션이 최대 10분 살아 있다 (ARCHITECTURE 4절).
- **목업은 로컬 서버로 연다**: `file://`로 열면 `shared.js` · `shared.css`를 못 불러온다 (`docs/mockups/README.md`).

## 하지 말 것
- `.env`를 커밋하지 않는다.
- access 토큰 · 세션 토큰을 `localStorage`에 넣지 않는다(세션은 `HttpOnly` 쿠키만).
- 이미 push한 마이그레이션 파일을 고치지 않는다.
- 업로드 금지 대상(위 절)에 이 프로젝트 내용을 올리지 않는다.

## 말투
대화 말투는 전역 설정(`~/.claude/CLAUDE.md`)을 따른다. 문서 · 주석 · 커밋 메시지는 이 저장소의 문체를 따른다(문서는 "~한다" 평서문, 화면 문구는 반말 — DESIGN 8장).
