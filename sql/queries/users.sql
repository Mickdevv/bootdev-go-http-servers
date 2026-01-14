-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_UUID(),
	NOW(),
	NOW(),
	$1,
	$2
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM  users;

-- name: GetUserProfileByEmail :one
SELECT id, email, created_at, updated_at  from users where email = $1;

-- name: GetUserForAuthByEmail :one
SELECT id, email, created_at, updated_at, hashed_password from users where email = $1;

-- name: GetAllUsers :many
SELECT email from users;
