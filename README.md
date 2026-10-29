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
**v2 재작성 중.** 2026-07의 v1(Express + React)을 버리고 새 스택으로 다시 만들고 있다. 지금은 설계 문서 단계이고, `client/` · `server/`는 v1 코드라 다음 단계(M1)에서 지운다. v1은 태그 [`v1-legacy`](https://github.com/sagming40/cbt-journal/tree/v1-legacy).

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
> M1(개발 환경 · 뼈대)에서 채운다. 필요한 도구: Go, Node 24 + pnpm, Task, Docker Desktop.

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
