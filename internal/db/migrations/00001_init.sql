-- 첫 마이그레이션: 테이블은 아직 없다. goose가 이 DB를 관리하기 시작했다는 기록만 남긴다 (PLAN M1).
-- 실제 테이블은 M2(users · sessions)부터 새 파일로 추가한다. push한 뒤에는 이 파일을 고치지 않는다 (CLAUDE.md).
-- 빈 파일 대신 아무 일도 안 하는 문장을 둔다: 문장 0개를 goose가 어떻게 다루는지에 기대지 않으려고.

-- +goose Up
SELECT 1;

-- +goose Down
SELECT 1;
