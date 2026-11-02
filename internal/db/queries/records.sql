-- 인지 왜곡 마스터 · 사고 기록 (DATA_MODEL 3.3 · 3.5절).
-- 사용자 데이터를 읽고 쓰는 SQL은 전부 user_id를 조건에 넣는다 — 남의 기록은 0행 → 404 (ARCHITECTURE 2절).

-- name: ListDistortions :many
SELECT id, name, description, example
FROM distortions
WHERE archived_at IS NULL
ORDER BY sort_order, id;

-- name: CreateRecord :one
-- 같은 초안으로 다시 오면 아무것도 넣지 않는다(0행 → 서비스가 있던 기록을 돌려준다, DATA_MODEL 3.5절).
-- 먼저 찾아보고 넣으면 동시에 온 두 요청이 둘 다 "없다"고 보므로, UNIQUE 제약에 맡긴다.
INSERT INTO thought_records (
    user_id, occurred_at, situation, automatic_thought,
    evidence_for, evidence_against, alternative_thought, source_draft_id
)
VALUES (
    @user_id, @occurred_at, @situation, @automatic_thought,
    @evidence_for, @evidence_against, @alternative_thought, @source_draft_id
)
ON CONFLICT (user_id, source_draft_id) DO NOTHING
RETURNING *;

-- name: GetRecordByDraft :one
SELECT * FROM thought_records
WHERE user_id = @user_id AND source_draft_id = @source_draft_id;

-- name: AddRecordEmotions :exec
-- 세 배열은 같은 순서 · 같은 길이다(서비스가 맞춘다). 쓰는 법은 AddCheckinEmotions와 같다.
INSERT INTO record_emotions (record_id, emotion_id, intensity_before, intensity_after)
SELECT @record_id::uuid, unnest(@emotion_ids::smallint[]), unnest(@befores::smallint[]), unnest(@afters::smallint[]);

-- name: AddRecordDistortions :exec
INSERT INTO record_distortions (record_id, distortion_id)
SELECT @record_id::uuid, unnest(@distortion_ids::smallint[]);

-- name: GetRecord :one
SELECT * FROM thought_records
WHERE id = @id AND user_id = @user_id;

-- name: ListRecordsBetween :many
-- [from, to) 범위. 경계 계산은 ListCheckinsBetween과 같다 (DATA_MODEL 3.4절).
SELECT * FROM thought_records
WHERE user_id = @user_id AND occurred_at >= @from_time AND occurred_at < @to_time
ORDER BY occurred_at, id;

-- name: ListRecordEmotions :many
-- 여러 기록의 감정을 한 번에. user_id를 같이 걸어 남의 기록 id가 섞여도 안 나온다.
SELECT re.record_id, re.emotion_id, e.name, re.intensity_before, re.intensity_after
FROM record_emotions re
JOIN thought_records r ON r.id = re.record_id
JOIN emotions e ON e.id = re.emotion_id
WHERE r.user_id = @user_id AND re.record_id = ANY(@record_ids::uuid[])
ORDER BY re.record_id, e.sort_order, e.id;

-- name: ListRecordDistortions :many
SELECT rd.record_id, rd.distortion_id, d.name
FROM record_distortions rd
JOIN thought_records r ON r.id = rd.record_id
JOIN distortions d ON d.id = rd.distortion_id
WHERE r.user_id = @user_id AND rd.record_id = ANY(@record_ids::uuid[])
ORDER BY rd.record_id, d.sort_order, d.id;

-- name: UpdateRecord :one
-- 통째로 고치기(T5). 나중에 저장한 쪽이 이긴다 (DATA_MODEL 3.5절).
UPDATE thought_records
SET occurred_at = @occurred_at,
    situation = @situation,
    automatic_thought = @automatic_thought,
    evidence_for = @evidence_for,
    evidence_against = @evidence_against,
    alternative_thought = @alternative_thought,
    updated_at = now()
WHERE id = @id AND user_id = @user_id
RETURNING *;

-- name: DeleteRecordEmotions :exec
-- UpdateRecord로 주인을 확인한 같은 트랜잭션 안에서만 부른다.
DELETE FROM record_emotions WHERE record_id = @record_id;

-- name: DeleteRecordDistortions :exec
-- UpdateRecord로 주인을 확인한 같은 트랜잭션 안에서만 부른다.
DELETE FROM record_distortions WHERE record_id = @record_id;

-- name: DeleteRecord :execrows
-- 감정 · 왜곡 행은 CASCADE로 같이 지워진다 (DATA_MODEL 3.9절).
DELETE FROM thought_records WHERE id = @id AND user_id = @user_id;
