# 0004. 메인 DB는 PostgreSQL로 한다

- 상태: 채택
- 날짜: 2026-10-29

## 맥락
- 데이터가 "사용자 → 기록 → 감정 · 인지 왜곡"으로 이어지는 **관계형**이다. 문서형 · 키-값 DB(MongoDB, DynamoDB)는 맞지 않는다.
- 최종 후보는 SQLite와 PostgreSQL 두 개였다.
  - SQLite 쪽 논리: Go 바이너리 하나 + DB 파일 하나로 운영 부담이 0에 가깝다. 이 앱 규모에서 "쓰기는 한 번에 하나" 제약은 병목이 아니다.
  - PostgreSQL 쪽 논리: 개발자가 가장 많이 쓰는 DB(2025 Stack Overflow 설문 1위)라 제대로 익혀 둘 가치가 있고, `timestamptz` · `jsonb` · 동시 쓰기 · 관리형 호스팅을 그대로 쓸 수 있다. 다중 사용자 공개([0001](0001-rewrite-v2.md))로 가는 길도 이쪽이 곧다.

## 결정
PostgreSQL을 쓴다. 드라이버 `pgx`, 쿼리 `sqlc`, 마이그레이션 `goose`.

- 시각은 전부 `timestamptz`(UTC 저장). 사용자 시간대는 설정으로 두고 일별 집계 때 변환한다.
- 스키마의 원본은 **마이그레이션 SQL 파일**이다. 문서(DATA_MODEL)에는 관계도와 설계 이유만 쓴다.

## 검토한 대안
- **SQLite (+ Litestream 백업)**: 운영은 가장 가볍다. 위 이유로 사용자가 PostgreSQL을 골랐다.
- **MySQL / MariaDB**: 가능하지만 이 앱에서 PostgreSQL보다 나은 점이 없다.

## 감수할 것
- DB 서버를 따로 운영 · 백업해야 한다. 로컬은 docker compose로 띄운다.
- 통합 테스트에 실제 PostgreSQL이 필요하다(`testcontainers-go`).
