# archive — 옛 문서 보관함

> **지금은 쓰지 않지만 남겨 두는 문서**를 모은 폴더다. 보관만 하고 고치지 않는다(옮길 때 링크 경로만 맞춘다).
> 지금 문서는 [CLAUDE.md](../../CLAUDE.md) "문서 지도", 진행 상황은 [PLAN.md](../PLAN.md).

## v2 개발 일지 — 끝난 단계
마일스톤이 끝나면 그 기간의 [DEVLOG](../DEVLOG.md) 날짜 블록을 여기로 옮긴다. DEVLOG에는 단계마다 한 줄 요약과 링크만 남는다.
두 단계에 걸친 블록은 뒤 단계 파일에, 지난 단계 일을 다음 단계 기간에 한 블록은 그 기간 단계 파일에 둔다.

| 파일 | 단계 | 기간 |
|---|---|---|
| [DEVLOG_M0.md](DEVLOG_M0.md) | M0 설계 문서 · 목업 | 2026-10-29 |
| [DEVLOG_M1.md](DEVLOG_M1.md) | M1 개발 환경 · 뼈대 | 2026-10-30 |
| [DEVLOG_M2.md](DEVLOG_M2.md) | M2 인증 (M1 · M2 집 PC 확인 포함) | 2026-10-30 |
| [DEVLOG_M3.md](DEVLOG_M3.md) | M3 첫 배포 | 2026-10-30 ~ 11-01 |
| [DEVLOG_M4.md](DEVLOG_M4.md) | M4 체크인 (스타일 D2 · PC 배치 · 회사 PC 서버 준비 포함) | 2026-11-01 ~ 02 |

## v1 문서
전면 재작성(2026-10-29) 전, v1(Express + React + PostgreSQL) 시절 문서다. v1 코드는 git 태그 `v1-legacy`에서 볼 수 있다. 재작성 결정은 [decisions/0001](../decisions/0001-rewrite-v2.md) 참고.

| 파일 | 내용 |
|---|---|
| `CBT_Journal_Feature_List.md` | v1 기능 정의서 |
| `CBT_Journal_ERD.md` | v1 ERD (테이블 4개) |
| `CBT_Journal_API_Specification.md` | v1 API 명세 (엔드포인트 8개) |
| `DEVLOG_v1.md` | v1 개발 일지 (2026-07-15 ~ 07-20) |
