-- 인지 왜곡 마스터와 사고 기록 (DATA_MODEL 3.3 · 3.5 · 3.8 · 3.9절, REQUIREMENTS T1 · T5 · T6).
-- 작성 중 초안은 여기 없다 — Valkey에 있다(DATA_MODEL 4.3절). 그래서 상태 컬럼이 없다.

-- +goose Up
CREATE TABLE distortions (
    id          smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        text NOT NULL UNIQUE CHECK (char_length(name) BETWEEN 1 AND 40),
    -- 문구는 사용자가 직접 쓴다. 비어 있으면 화면은 이름만 보여 준다 (DATA_MODEL 3.3절)
    description text,
    example     text,
    sort_order  smallint NOT NULL,
    archived_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- 인지 왜곡 11개 (DATA_MODEL 3.3절 seed, v1과 같음). 화면 순서 = sort_order
INSERT INTO distortions (name, sort_order) VALUES
    ('흑백논리', 1), ('과도한 일반화', 2), ('정신적 필터', 3), ('긍정적인 것 무시하기', 4),
    ('성급한 결론(독심술/예언자적 오류)', 5), ('확대와 축소', 6), ('감정적 추론', 7),
    ('당위적 진술', 8), ('낙인찍기', 9), ('개인화', 10), ('파국화', 11);

CREATE TABLE thought_records (
    id                  uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id             uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- 사고가 일어난 때(T6). UTC로 저장하고 "그날"은 사용자 시간대로 계산한다
    occurred_at         timestamptz NOT NULL,
    -- 글자 수는 service가 먼저 보고, DB는 마지막 방어선 (DATA_MODEL 3.5절)
    situation           text NOT NULL CHECK (char_length(situation) BETWEEN 1 AND 1000),
    automatic_thought   text NOT NULL CHECK (char_length(automatic_thought) BETWEEN 1 AND 1000),
    -- 근거는 못 찾는 날도 있다. NULL 대신 ''로 통일
    evidence_for        text NOT NULL DEFAULT '' CHECK (char_length(evidence_for) <= 2000),
    evidence_against    text NOT NULL DEFAULT '' CHECK (char_length(evidence_against) <= 2000),
    alternative_thought text NOT NULL CHECK (char_length(alternative_thought) BETWEEN 1 AND 1000),
    -- 이 기록을 만든 초안. 같은 초안을 두 번 완료해도 기록이 하나이게 (DATA_MODEL 3.5절)
    source_draft_id     uuid,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, source_draft_id)
);

CREATE TABLE record_emotions (
    record_id        uuid NOT NULL REFERENCES thought_records (id) ON DELETE CASCADE,
    emotion_id       smallint NOT NULL REFERENCES emotions (id) ON DELETE RESTRICT,
    -- 고른 감정마다 전 · 후를 전부 매긴다 (DATA_MODEL 3.5절)
    intensity_before smallint NOT NULL CHECK (intensity_before BETWEEN 0 AND 100),
    intensity_after  smallint NOT NULL CHECK (intensity_after BETWEEN 0 AND 100),
    PRIMARY KEY (record_id, emotion_id)
);

CREATE TABLE record_distortions (
    record_id     uuid NOT NULL REFERENCES thought_records (id) ON DELETE CASCADE,
    distortion_id smallint NOT NULL REFERENCES distortions (id) ON DELETE RESTRICT,
    PRIMARY KEY (record_id, distortion_id)
);

-- 그날의 기록 · 목록 · 통계 기간 조회 (DATA_MODEL 3.8절)
CREATE INDEX thought_records_user_occurred_idx ON thought_records (user_id, occurred_at DESC);

-- +goose Down
DROP TABLE record_distortions;
DROP TABLE record_emotions;
DROP TABLE thought_records;
DROP TABLE distortions;
