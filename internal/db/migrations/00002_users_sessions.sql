-- 사용자와 로그인 세션 (DATA_MODEL 3.1 · 3.2절, ARCHITECTURE 4절).

-- +goose Up
CREATE TABLE users (
    -- uuidv7(): PostgreSQL 18 내장. 시간순이라 인덱스가 덜 흩어지고, 순번처럼 전체 개수를 짐작하게 하지 않는다 (DATA_MODEL 1절)
    id            uuid PRIMARY KEY DEFAULT uuidv7(),
    -- 앱이 소문자로 정규화해서 넣는다. DB는 마지막 방어선으로 같은지 확인한다 (citext 확장을 쓰지 않으려고)
    email         text NOT NULL UNIQUE CHECK (email = lower(email) AND char_length(email) BETWEEN 3 AND 254),
    -- argon2id PHC 문자열 ($argon2id$v=19$m=…,t=…,p=…$salt$hash). 파라미터가 같이 들어 있어 나중에 올려도 옛 해시를 검증할 수 있다
    password_hash text NOT NULL CHECK (password_hash LIKE '$argon2id$%'),
    nickname      text NOT NULL CHECK (char_length(nickname) BETWEEN 1 AND 30),
    -- IANA 시간대 이름. "하루"의 경계를 이걸로 계산한다 (REQUIREMENTS A4)
    timezone      text NOT NULL DEFAULT 'Asia/Seoul',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id           uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id      uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- 쿠키의 토큰 원문이 아니라 SHA-256 해시(32바이트)만 둔다. DB가 새도 세션을 훔칠 수 없게 (ARCHITECTURE 4절)
    token_hash   bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    -- 로그인 유지(A1): 기한이 30일 / 12시간으로 갈린다 (DATA_MODEL 3.2절)
    remember     boolean NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL
);

-- 비밀번호 변경 · 다른 기기 로그아웃 때 그 사용자의 세션을 찾는다 (DATA_MODEL 3.8절). token_hash는 UNIQUE라 인덱스가 이미 있다
CREATE INDEX sessions_user_id_idx ON sessions (user_id);

-- +goose Down
DROP TABLE sessions;
DROP TABLE users;
