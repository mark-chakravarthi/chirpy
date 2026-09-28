-- name: CreateChirp :one
INSERT INTO chirps (id, body, user_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
WHERE user_id = COALESCE(sqlc.narg('author_id'), user_id)
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps WHERE id = $1;