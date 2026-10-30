-- 세션은 토큰 해시로 찾는다. 사용자 단위 작업(다른 기기 로그아웃 등)은 user_id 조건을 같이 건다 (ARCHITECTURE 2절).

-- name: CreateSession :exec
INSERT INTO sessions (user_id, token_hash, remember, expires_at)
VALUES (@user_id, @token_hash, @remember, @expires_at);

-- name: GetSession :one
-- 만료된 세션은 없는 것과 같다.
SELECT user_id, remember, last_seen_at, expires_at
FROM sessions
WHERE token_hash = @token_hash AND expires_at > now();

-- name: TouchSession :exec
-- 기한 연장 (DATA_MODEL 3.2절). 요청마다가 아니라 last_seen_at이 일정 시간 지났을 때만 부른다.
UPDATE sessions
SET last_seen_at = now(), expires_at = @expires_at
WHERE token_hash = @token_hash;

-- name: DeleteSession :many
-- 지운 해시를 돌려받아 Valkey sess: 키도 지운다 (ARCHITECTURE 4절).
DELETE FROM sessions WHERE token_hash = @token_hash
RETURNING token_hash;

-- name: DeleteOtherSessions :many
-- 비밀번호 변경 · 다른 기기 모두 로그아웃(A8): 지금 세션만 남긴다.
DELETE FROM sessions WHERE user_id = @user_id AND token_hash <> @keep_token_hash
RETURNING token_hash;

-- name: DeleteUserSessions :many
-- 관리 명령 reset-password: 그 사용자의 세션을 전부 끊는다.
DELETE FROM sessions WHERE user_id = @user_id
RETURNING token_hash;

-- name: DeleteExpiredUserSessions :exec
-- 로그인할 때 그 사용자의 만료된 세션을 같이 치운다 (전체 정리는 M7 worker).
DELETE FROM sessions WHERE user_id = @user_id AND expires_at <= now();
