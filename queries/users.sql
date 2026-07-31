-- name: CreateUser :one
INSERT INTO users (email, username, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUsername :one
UPDATE users
SET username = $2
WHERE id = $1
RETURNING username;

-- name: UpdateEmail :one
UPDATE users
SET email = $2
WHERE id = $1
RETURNING email;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;