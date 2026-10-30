# CBT 저널

인지행동치료(CBT)의 **사고 기록(Thought Record)** 을 매일 짧게 연습하는 웹앱이다.
떠오른 생각을 그 자리에서 붙잡고, 나중에 근거를 따져 보며 더 균형 잡힌 생각으로 바꾸는 연습을 돕는다.

> 이 앱은 자기 관찰 · 연습 도구이며 **치료를 대체하지 않는다.**

## 무엇을 하나
- **체크인** — 지금 감정과 강도(0 ~ 100)만 10초 안에 남긴다.
- **사고 기록** — 상황 → 감정 → 자동 사고 → 인지 왜곡 → 근거 · 반대 근거 → 대안 사고 → 감정 다시 매기기. 휴대폰으로 앞부분만 적어 두고 PC에서 이어 써도 된다(자동 임시저장).
- **통계** — 감정 추이, 사고 기록 전 · 후 변화, 자주 나오는 인지 왜곡.
- (예정) 리마인더 알림, 주간 리포트.

자세한 범위는 [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md).

## 상태
**v2 재작성 중.** 2026-07의 v1(Express + React)을 버리고 새 스택으로 다시 만들고 있다. 지금은 로그인까지 되는 빈 앱이고, 다음은 첫 배포(M3)다. v1 코드는 태그 [`v1-legacy`](https://github.com/sagming40/cbt-journal/tree/v1-legacy).

어디까지 왔는지는 [docs/PLAN.md](docs/PLAN.md) 2장에서 본다(이 README에는 진행 상황을 적지 않는다).

## 스택
| 영역 | 기술 |
|---|---|
| 백엔드 | Go (`net/http`, pgx, sqlc, goose, oapi-codegen), asynq |
| 저장소 | PostgreSQL (영구 데이터), Valkey (세션 캐시 · 요청 제한 · 초안 · 작업 큐) |
| 프론트엔드 | SvelteKit SPA (Svelte 5, TypeScript), shadcn-svelte, Tailwind CSS |
| 배포 단위 | 실행 파일 하나 (프론트를 `go:embed`로 포함, `serve` / `worker`) |

왜 이 스택인지는 [docs/decisions/](docs/decisions/README.md).

## 실행
> Windows PowerShell 기준이다. 자주 쓰는 명령은 `Taskfile.yml`에 묶여 있다 — `task`만 치면 목록이 나온다.
> `task`는 `.env`를 알아서 읽으므로 터미널에 따로 불러오지 않는다.
> 포트: Postgres `5436` · Valkey `6380` · API `8080` · 프론트 dev `5176` (ARCHITECTURE 6절).

### 처음 한 번 (PC마다)
```powershell
# 1. 도구 (Docker Desktop · Node 24는 따로 설치)
winget install --id GoLang.Go -e
winget install --id pnpm.pnpm -e
winget install --id Task.Task -e
winget install --id GolangCI.golangci-lint -e   # 버전은 .github/workflows/ci.yml의 version과 맞춘다

# 2. 비밀값 파일 만들기 → .env를 열어 POSTGRES_PASSWORD와 DATABASE_URL 안의 비밀번호를 같은 값으로 바꾼다
Copy-Item .env.example .env

# 3. Postgres · Valkey 띄우고, Go · 프론트 의존성 받고, 테이블 만들기
docker compose up -d
task setup
task migrate

# 4. 내 계정 만들기 (가입이 꺼져 있어서 명령으로 만든다. 비밀번호는 실행 중에 물어본다 — 12자 이상)
task user -- create --email 내이메일 --nickname 내닉네임
```
- 계정과 기록은 PC마다 따로다(DB가 기기마다 따로 — CLAUDE.md "기기 둘"). 두 PC에서 같은 계정을 쓰려면 각각 만든다. 실제로 쓰는 계정은 배포(M3) 뒤 서버에 하나만 둔다.
- 비밀번호는 영문 · 숫자 · `-` · `_`만 쓰는 게 편하다. `@` `:` `/` `?` `#` `%`가 있으면 `DATABASE_URL` 안에서 URL 인코딩해야 한다.
- Postgres 비밀번호는 **볼륨을 처음 만들 때 한 번만** 정해진다. 나중에 `.env`의 `POSTGRES_PASSWORD`만 바꾸면 DB에는 반영되지 않는다(바꾸려면 볼륨을 지우고 다시 만든다 — 아래 "멈추기 · 지우기").

### 매일 시작할 때
```powershell
git pull --ff-only
docker compose up -d
task migrate    # 다른 PC에서 마이그레이션이 늘었으면 적용
task dev        # API(air, 8080) + 프론트(Vite, 5176)를 한 터미널에. 끌 때는 Ctrl+C
```
- 화면은 `http://localhost:5176` (`/api`는 8080으로 넘어간다). Go 파일을 저장하면 air가 다시 빌드 · 재시작한다(Windows라 로그에 `Process Exit with Code: 1`이 찍히는 건 정상).
- `git pull`로 `go.mod` · `go.sum` · `web/pnpm-lock.yaml`이 바뀌었으면 `task setup`을 먼저 한다.

### 확인 · 검사
```powershell
task test        # Go 테스트(통합 테스트는 Docker Desktop이 켜져 있어야 돈다 — 꺼져 있으면 건너뜀) + 프론트 Vitest
task lint        # golangci-lint(govet · staticcheck · gosec …) + 프론트 타입 검사 · Prettier · ESLint
task fmt         # 형식 자동 맞추기 (gofmt · goimports · Prettier) — lint가 형식 오류를 내면
curl.exe -i http://127.0.0.1:8080/api/health   # 200 {"postgres":"ok","status":"ok","valkey":"ok"}, 하나라도 죽으면 503

go mod tidy      # Go import를 추가 · 삭제한 뒤 go.mod · go.sum 정리
```
main에 push하면 GitHub Actions(`.github/workflows/ci.yml`)가 Go lint · test, 프론트 check · lint, `task build`를 돌린다. 결과는 GitHub 저장소의 Actions 탭.

### 실행 파일 하나로 빌드 · 실행
```powershell
task build       # 프론트 빌드 → internal/webdist/dist로 복사 → bin/app.exe
task serve       # 화면은 http://127.0.0.1:8080 (Vite 필요 없음). 끌 때는 Ctrl+C
```
프론트를 고쳤으면 `task build`를 다시 해야 실행 파일에 들어간다. 개발 중에는 `task dev`(5176)로 보면 된다.

### 코드 생성 (`api/openapi.yaml` · `internal/db/queries/*.sql`을 고친 뒤)
```powershell
task gen         # → internal/apigen (Go) + internal/db/dbgen (sqlc, Docker로 실행) + web/src/lib/api/schema.d.ts (TS)
go build ./...   # 새 경로를 아직 구현하지 않았으면 여기서 "missing method"로 깨진다 (일부러)
task lint        # 프론트에서 계약서와 맞지 않는 호출이 있으면 타입 오류

# shadcn-svelte 부품 추가 → web/src/lib/components/ui/
Set-Location web; pnpm dlx shadcn-svelte@latest add <부품 이름>; Set-Location ..
```
생성된 파일은 직접 고치지 않고, 생성 결과도 커밋한다 (CLAUDE.md "코드 규칙"). CI가 `task gen` 뒤 바뀐 파일이 없는지 확인한다.

### 마이그레이션 (goose)
```powershell
task migrate -- status   # 파일별 pending / applied
task migrate             # 남은 것 전부 적용 (= up)
task migrate -- down     # 마지막 하나만 되돌림
```
- 새 파일은 `internal/db/migrations/`에 다음 번호로 만든다: `00003_checkins.sql`. 파일 안에 `-- +goose Up` / `-- +goose Down` 표시가 있어야 한다.
- **push한 마이그레이션 파일은 고치지 않는다.** 바꿀 게 있으면 새 번호로 만든다. 스키마는 고친 PC에서만 만든다 (CLAUDE.md "기기 둘").

### Postgres · Valkey 직접 보기
```powershell
docker compose ps                     # 둘 다 healthy인지
docker compose logs -f postgres       # 로그 (Ctrl+C로 나옴)
docker compose exec postgres psql -U cbt -d cbt_journal                      # SQL 콘솔 (\q로 나옴)
docker compose exec postgres psql -U cbt -d cbt_journal -c "select * from goose_db_version;"
docker compose exec valkey valkey-cli                                        # Valkey 콘솔 (exit로 나옴)
docker compose exec valkey valkey-cli ping                                   # PONG
```
`.env`에서 사용자 · DB 이름을 바꿨다면 `-U` · `-d`도 맞춘다.

### 멈추기 · 지우기
```powershell
docker compose stop        # 컨테이너만 멈춤 (데이터 남음)
docker compose down        # 컨테이너 삭제 (볼륨 · 데이터 남음)
docker compose down -v     # 볼륨까지 삭제 = 로컬 DB · Valkey 데이터가 전부 사라진다. 다시 up → task migrate
```

### 자주 막히는 것
| 증상 | 원인 · 해결 |
|---|---|
| `password authentication failed` (28P01) | `DATABASE_URL` 안의 비밀번호가 `POSTGRES_PASSWORD`와 다르다 |
| `설정을 읽지 못했다: APP_ENV: 값이 없다` | `.env`가 없거나 키가 빠졌다. `task` 없이 `go run`으로 직접 실행했다면 `.env`가 안 읽힌다 — `task` 명령을 쓴다 |
| `.env`를 고쳤는데 그대로다 | 터미널에 같은 이름의 환경변수가 이미 있으면 `task`가 `.env`로 덮어쓰지 않는다. 새 터미널에서 다시 |
| `dial tcp ... :5436 ... refused` | 컨테이너가 꺼져 있다 → `docker compose up -d` |
| 엉뚱한 DB에 붙는다 | 포트가 `5432`(직접 설치한 Postgres)로 돼 있다. compose는 `5436` |
| psql 결과 뒤에 `(END)`가 뜨고 멈춤 | 페이저다. `q`로 나온다. 한 줄만 보려면 `-tAc "…"` |
| `bind: ... 8080` 또는 `Port 5176 is in use` | 이미 다른 `task dev` · `task serve`가 떠 있다(Ctrl+C 뒤 남은 프로세스일 수도 있다). 그 터미널을 끄거나 작업 관리자에서 `app.exe` · `node.exe`를 끝낸다 |
| 8080에 "프론트 빌드가 없다" (503) | 프론트를 빌드 · 복사하지 않고 `go build` 했다 → `task build` |
| 화면에 "연결이 안 돼" | Go 서버가 꺼져 있다. `web`에서 `pnpm dev`만 띄우면 `/api`가 안 된다 → `task dev` |
| VS Code가 `layout.css`의 `@theme` · `@apply`에 경고 | Tailwind CSS IntelliSense 확장 + `"files.associations": { "*.css": "tailwindcss" }` 설정. `web/.vscode/settings.json`은 VS Code로 `web` 폴더를 직접 열 때만 읽힌다 |
| 로그인이 "잠시 뒤에 다시 시도해 줘" | 요청 제한(같은 이메일 실패 5번 · 같은 IP 20번 / 15분). 기다리거나, 개발 중이면 `docker compose exec valkey valkey-cli --scan --pattern "rl:*"`로 보고 `del`로 지운다 |
| 비밀번호를 잊었다 | `task user -- reset-password --email 내이메일` (그 계정의 세션도 전부 끊긴다) |
| "로그인 유지"를 안 했는데 브라우저를 닫았다 열어도 로그인돼 있다 | Chrome이 실제로 안 꺼졌다. `chrome://settings/system`의 "Chrome 종료 후에도 백그라운드 앱 계속 실행"을 끈다. 쿠키는 정상(개발자 도구 Application → Cookies에서 `session`의 Expires가 "Session") |
| `task test`에서 통합 테스트가 `SKIP` | Docker Desktop이 꺼져 있다. 켜고 다시. CI에서는 건너뛰지 않고 실패한다 |

## 문서
| 문서 | 내용 |
|---|---|
| [REQUIREMENTS](docs/REQUIREMENTS.md) | 무엇을 만들고 무엇을 만들지 않나 |
| [ARCHITECTURE](docs/ARCHITECTURE.md) | 코드 구조, 요청 · 인증 흐름, 개발 환경 |
| [DATA_MODEL](docs/DATA_MODEL.md) | PostgreSQL 테이블 · Valkey 키와 설계 이유 |
| [DESIGN](docs/DESIGN.md) | 화면 설계, 색 · 테마 ([목업](docs/mockups/README.md)) |
| [PLAN](docs/PLAN.md) | 단계별 작업과 완료 기준, 진행 상황 |
| [SERVER](docs/SERVER.md) | 운영 서버 접속 · 세팅 기록 · 자주 쓰는 명령 (런북) |
| [DEVLOG](docs/DEVLOG.md) | 개발 일지 |
| [decisions/](docs/decisions/README.md) | 결정 기록 (ADR) |
| [CLAUDE.md](CLAUDE.md) | 이 저장소에서 작업할 때 지키는 규칙 |

## 만든 사람
사공민규 · 1인 취미 프로젝트
