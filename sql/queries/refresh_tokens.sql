-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, revoked_at, expires_at, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT * from refresh_tokens where token = $1; 

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = $2, updated_at = $2
WHERE token = $1
RETURNING *;
