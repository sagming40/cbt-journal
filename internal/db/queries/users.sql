-- name: CreateUser :one
INSERT INTO users (email, password_hash, nickname)
VALUES (@email, @password_hash, @nickname)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = @email;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = @id;

-- name: UpdateUserProfile :one
UPDATE users
SET nickname = @nickname, timezone = @timezone, updated_at = now()
WHERE id = @id
RETURNING *;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = @password_hash, updated_at = now()
WHERE id = @id;
