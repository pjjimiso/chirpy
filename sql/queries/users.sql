-- name: CreateUser :one
INSERT INTO users(
	id, 
	created_at, 
	updated_at, 
	email,
	hashed_passwords
)
VALUES (
	gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2
)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: TruncateUsers :exec
TRUNCATE TABLE users CASCADE;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserCredentials :exec
UPDATE users
SET email = $1, hashed_passwords = $2, updated_at = NOW()
WHERE id = $3;

-- name: UpdateUserAddChirpyRed :exec
UPDATE users
SET is_chirpy_red = TRUE 
WHERE id = $1;

