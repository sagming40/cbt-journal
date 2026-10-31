-- name: CreateUser :one
INSERT INTO users (email, password_hash, nickname)
VALUES (@email, @password_hash, @nickname)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = @email;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = @id;

-- name: GetUserTimezone :one
-- "그날"의 경계를 계산할 때만 쓴다 — 비밀번호 해시까지 읽어 올 필요가 없다.
SELECT timezone FROM users WHERE id = @id;

-- name: UpdateUserProfile :one
UPDATE users
SET nickname = @nickname, timezone = @timezone, updated_at = now()
WHERE id = @id
RETURNING *;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = @password_hash, updated_at = now()
WHERE id = @id;
