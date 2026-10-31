-- 감정 마스터 · 체크인 (DATA_MODEL 3.3 · 3.4절).
-- 사용자 데이터를 읽고 쓰는 SQL은 전부 user_id를 조건에 넣는다 — 남의 기록은 0행 → 404 (ARCHITECTURE 2절).

-- name: ListEmotions :many
-- 기본 감정(user_id NULL) 중 숨기지 않은 것을 화면 순서대로.
SELECT id, name
FROM emotions
WHERE user_id IS NULL AND archived_at IS NULL
ORDER BY sort_order, id;

-- name: CreateCheckin :one
INSERT INTO checkins (user_id, occurred_at, note)
VALUES (@user_id, @occurred_at, @note)
RETURNING *;

-- name: AddCheckinEmotions :exec
-- 감정 · 강도 여러 개를 한 번에. 두 배열은 같은 순서 · 같은 길이다(서비스가 맞춘다).
-- SELECT 줄에 unnest를 둘 놓으면 Postgres가 같은 자리끼리 짝지어 펼친다(sqlc가 인자 둘짜리 unnest를 몰라서 이렇게 쓴다).
INSERT INTO checkin_emotions (checkin_id, emotion_id, intensity)
SELECT @checkin_id::uuid, unnest(@emotion_ids::smallint[]), unnest(@intensities::smallint[]);

-- name: GetCheckin :one
SELECT * FROM checkins
WHERE id = @id AND user_id = @user_id;

-- name: ListCheckinsBetween :many
-- [from, to) 범위. "그날"의 경계(사용자 시간대 0시 ~ 다음 날 0시)는 서비스가 UTC로 바꿔서 넘긴다 (DATA_MODEL 3.4절).
-- 하루를 시간순으로 읽게 오름차순 (DESIGN 4.1절 "오늘 목록").
SELECT * FROM checkins
WHERE user_id = @user_id AND occurred_at >= @from_time AND occurred_at < @to_time
ORDER BY occurred_at, id;

-- name: ListCheckinEmotions :many
-- 여러 체크인의 감정을 한 번에(체크인마다 따로 부르지 않게). user_id를 같이 걸어 남의 체크인 id가 섞여도 안 나온다.
SELECT ce.checkin_id, ce.emotion_id, e.name, ce.intensity
FROM checkin_emotions ce
JOIN checkins c ON c.id = ce.checkin_id
JOIN emotions e ON e.id = ce.emotion_id
WHERE c.user_id = @user_id AND ce.checkin_id = ANY(@checkin_ids::uuid[])
ORDER BY ce.checkin_id, e.sort_order, e.id;

-- name: UpdateCheckin :one
-- 통째로 고치기(C4). 나중에 저장한 쪽이 이긴다 (DATA_MODEL 3.4절).
UPDATE checkins
SET occurred_at = @occurred_at, note = @note, updated_at = now()
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: DeleteCheckinEmotions :exec
-- UpdateCheckin으로 주인을 확인한 같은 트랜잭션 안에서만 부른다.
DELETE FROM checkin_emotions WHERE checkin_id = @checkin_id;

-- name: DeleteCheckin :execrows
-- 감정 행은 CASCADE로 같이 지워진다 (DATA_MODEL 3.9절).
DELETE FROM checkins WHERE id = @id AND user_id = @user_id;
