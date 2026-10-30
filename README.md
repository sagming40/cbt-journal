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
**v2 재작성 중.** 2026-07의 v1(Express + React)을 버리고 새 스택으로 다시 만들고 있다. 지금은 개발 환경 · 뼈대 단계(M1)다. v1 코드는 태그 [`v1-legacy`](https://github.com/sagming40/cbt-journal/tree/v1-legacy).

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
> Windows PowerShell 기준이다. Taskfile(M1)이 생기면 `task …` 명령으로 줄어들고, 그때 이 절을 고친다.
> 포트: Postgres `5436` · Valkey `6380` · API `8080` (ARCHITECTURE 6절).

### 처음 한 번 (PC마다)
```powershell
# 1. 도구 (Docker Desktop · Node 24는 따로 설치)
winget install --id GoLang.Go -e
winget install --id pnpm.pnpm -e
winget install --id Task.Task -e

# 2. 비밀값 파일 만들기 → .env를 열어 POSTGRES_PASSWORD와 DATABASE_URL 안의 비밀번호를 같은 값으로 바꾼다
Copy-Item .env.example .env

# 3. Postgres · Valkey 띄우고, Go 의존성 받고, 테이블 만들기
docker compose up -d
go mod download
go run ./cmd/app migrate up
```
- 비밀번호는 영문 · 숫자 · `-` · `_`만 쓰는 게 편하다. `@` `:` `/` `?` `#` `%`가 있으면 `DATABASE_URL` 안에서 URL 인코딩해야 한다.
- Postgres 비밀번호는 **볼륨을 처음 만들 때 한 번만** 정해진다. 나중에 `.env`의 `POSTGRES_PASSWORD`만 바꾸면 DB에는 반영되지 않는다(바꾸려면 볼륨을 지우고 다시 만든다 — 아래 "멈추기 · 지우기").

### 매일 시작할 때
```powershell
git pull --ff-only
docker compose up -d

# .env를 이 터미널에 불러온다 (터미널을 새로 열 때마다, .env를 고칠 때마다)
Get-Content .env | Where-Object { $_ -match '^\s*[^#\s][^=]*=' } | ForEach-Object { $k, $v = $_ -split '=', 2; Set-Item "env:$($k.Trim())" $v.Trim() }

go run ./cmd/app migrate up    # 다른 PC에서 마이그레이션이 늘었으면 적용
go run ./cmd/app serve         # 끌 때는 Ctrl+C
```
`git pull`로 `go.mod` · `go.sum`이 바뀌었으면 `go mod download`를 먼저 한다.

### 확인 · 검사
```powershell
curl.exe -i http://127.0.0.1:8080/api/health   # 200 {"postgres":"ok","status":"ok","valkey":"ok"}, 하나라도 죽으면 503

go fmt ./...      # 형식 맞추기
go vet ./...      # 흔한 실수 검사
go test ./...     # 테스트
go mod tidy       # import를 추가 · 삭제한 뒤 go.mod · go.sum 정리
```

### 마이그레이션 (goose)
```powershell
go run ./cmd/app migrate status   # 파일별 pending / applied
go run ./cmd/app migrate up       # 남은 것 전부 적용
go run ./cmd/app migrate down     # 마지막 하나만 되돌림
```
- 새 파일은 `internal/db/migrations/`에 다음 번호로 만든다: `00002_users.sql`. 파일 안에 `-- +goose Up` / `-- +goose Down` 표시가 있어야 한다.
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
docker compose down -v     # 볼륨까지 삭제 = 로컬 DB · Valkey 데이터가 전부 사라진다. 다시 up → migrate up
```

### 자주 막히는 것
| 증상 | 원인 · 해결 |
|---|---|
| `password authentication failed` (28P01) | `DATABASE_URL` 안의 비밀번호가 `POSTGRES_PASSWORD`와 다르다. `.env`를 고친 뒤 **불러오기 한 줄을 다시** 실행한다 |
| `설정을 읽지 못했다: APP_ENV: 값이 없다` | 이 터미널에 `.env`를 안 불러왔다 |
| `dial tcp ... :5436 ... refused` | 컨테이너가 꺼져 있다 → `docker compose up -d` |
| 엉뚱한 DB에 붙는다 | 포트가 `5432`(직접 설치한 Postgres)로 돼 있다. compose는 `5436` |
| psql 결과 뒤에 `(END)`가 뜨고 멈춤 | 페이저다. `q`로 나온다. 한 줄만 보려면 `-tAc "…"` |
| `bind: ... 8080` 시작 실패 | 이미 다른 `serve`가 떠 있다. 그 터미널에서 Ctrl+C |

## 문서
| 문서 | 내용 |
|---|---|
| [REQUIREMENTS](docs/REQUIREMENTS.md) | 무엇을 만들고 무엇을 만들지 않나 |
| [ARCHITECTURE](docs/ARCHITECTURE.md) | 코드 구조, 요청 · 인증 흐름, 개발 환경 |
| [DATA_MODEL](docs/DATA_MODEL.md) | PostgreSQL 테이블 · Valkey 키와 설계 이유 |
| [DESIGN](docs/DESIGN.md) | 화면 설계, 색 · 테마 ([목업](docs/mockups/README.md)) |
| [PLAN](docs/PLAN.md) | 단계별 작업과 완료 기준, 진행 상황 |
| [DEVLOG](docs/DEVLOG.md) | 개발 일지 |
| [decisions/](docs/decisions/README.md) | 결정 기록 (ADR) |
| [CLAUDE.md](CLAUDE.md) | 이 저장소에서 작업할 때 지키는 규칙 |

## 만든 사람
사공민규 · 1인 취미 프로젝트
