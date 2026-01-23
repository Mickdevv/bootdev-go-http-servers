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

-- name: GetUserById :one
SELECT id, email, created_at, updated_at, is_chirpy_red  from users where id = $1;

-- name: GetUserProfileByEmail :one
SELECT id, email, created_at, updated_at, is_chirpy_red  from users where email = $1;

-- name: GetUserForAuthByEmail :one
SELECT id, email, created_at, updated_at, hashed_password, is_chirpy_red from users where email = $1;

-- name: GetAllUsers :many
SELECT email from users;

-- name: UpdateUser :one
UPDATE users SET updated_at = NOW(), email = $2, hashed_password = $3 WHERE id = $1 RETURNING *;

-- name: UpgradeUserChirpyRed :one
UPDATE users SET is_chirpy_red = $2, updated_at = NOW() WHERE id = $1 RETURNING *;
