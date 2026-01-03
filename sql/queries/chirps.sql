-- name: CreateChirp :one
INSERT INTO chirps (id, user_id, created_at, updated_at, body)
VALUES (
	gen_random_UUID(),
	$1,
	NOW(),
	NOW(),
	$2
)
RETURNING *;

-- name: GetChirp :one
SELECT * FROM chirps WHERE id = $1; 
-- name: GetAllChirps :many
SELECT * FROM chirps; 

