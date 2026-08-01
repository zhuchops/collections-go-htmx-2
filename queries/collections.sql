-- name: CreateCollection :one
INSERT INTO collections (user_id, title)
VALUES ($1, $2)
RETURNING *;

-- name: GetCollection :one
SELECT * FROM collections
WHERE id = $1 AND user_id = $2;

-- name: GetCollectionsByUserId :many
SELECT * FROM collections
WHERE user_id = $1;

-- name: DeleteCollection :exec
DELETE FROM collections
WHERE id = $1 AND user_id = $2;