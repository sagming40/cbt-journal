-- 감정 마스터와 체크인 (DATA_MODEL 3.3 · 3.4 · 3.8 · 3.9절, REQUIREMENTS C1 · C2 · C4).

-- +goose Up
CREATE TABLE emotions (
    -- 마스터라 몇십 개뿐이고 밖에 보여도 괜찮아서 uuid 대신 작은 순번 (DATA_MODEL 1절)
    id          smallint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- NULL = 시스템 기본 감정, 값이 있으면 사용자 정의(X2, 지금은 안 씀)
    user_id     uuid REFERENCES users (id) ON DELETE CASCADE,
    name        text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 20),
    sort_order  smallint NOT NULL,
    -- 기록이 참조 중이면 지울 수 없으니(RESTRICT) 안 쓰게 된 감정은 숨긴다 (DATA_MODEL 3.3절)
    archived_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    -- 기본 감정끼리도(user_id가 둘 다 NULL) 이름이 겹치지 않게
    UNIQUE NULLS NOT DISTINCT (user_id, name)
);

-- 기본 감정 10개 (DATA_MODEL 3.3절 seed). 화면 순서 = sort_order
INSERT INTO emotions (name, sort_order) VALUES
    ('불안', 1), ('두려움', 2), ('슬픔', 3), ('무기력', 4), ('분노', 5),
    ('짜증', 6), ('수치심', 7), ('죄책감', 8), ('외로움', 9), ('좌절', 10);

CREATE TABLE checkins (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id     uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- 감정을 느낀 때. UTC로 저장하고 "그날"은 사용자 시간대로 계산한다 (DATA_MODEL 3.4절)
    occurred_at timestamptz NOT NULL DEFAULT now(),
    -- 한 줄 메모(선택). NULL 대신 ''로 통일 (DESIGN 4.2절 최대 200자)
    note        text NOT NULL DEFAULT '' CHECK (char_length(note) <= 200),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE checkin_emotions (
    checkin_id uuid NOT NULL REFERENCES checkins (id) ON DELETE CASCADE,
    -- 기록이 쓰는 감정은 지우지 못한다 → archived_at으로 숨긴다 (DATA_MODEL 3.9절)
    emotion_id smallint NOT NULL REFERENCES emotions (id) ON DELETE RESTRICT,
    -- 앱 검증과 별개로 DB가 마지막 방어선 (DATA_MODEL 1절)
    intensity  smallint NOT NULL CHECK (intensity BETWEEN 0 AND 100),
    -- 한 체크인에 같은 감정은 한 번만
    PRIMARY KEY (checkin_id, emotion_id)
);

-- 오늘 체크인 · 목록 · 통계 기간 조회 (DATA_MODEL 3.8절)
CREATE INDEX checkins_user_occurred_idx ON checkins (user_id, occurred_at DESC);

-- +goose Down
DROP TABLE checkin_emotions;
DROP TABLE checkins;
DROP TABLE emotions;
