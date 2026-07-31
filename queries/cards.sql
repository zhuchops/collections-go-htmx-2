-- name: CreateCard :one
INSERT INTO cards (user_id, collection_id, title, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCard :one
SELECT * FROM cards
WHERE id = $1 AND user_id = $2;

-- name: GetCards :many
SELECT * FROM cards
WHERE user_id = $1 AND collection_id = $2;

-- name: UpdateCard :exec
UPDATE cards
SET title = $3, description = $4
WHERE id = $1 AND user_id=$2;

-- name: DeleteCard :exec
DELETE FROM cards
WHERE id=$1 AND user_id=$2;