-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email,hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
    
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users;

-- name: SaveChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetChirps :many
SELECT *
FROM chirps
ORDER BY created_at ASC;

-- name: GetChirp :one
SELECT *
FROM chirps
WHERE id = $1;

-- name: GetUser :one
SELECT *
FROM users
WHERE email = $1;

-- name: AddRefreshToken :one
INSERT INTO refresh_tokens(token,created_at,updated_at,user_id,expires_at,revoked_at)
VALUES(
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT user_id
FROM refresh_tokens
WHERE token = $1 AND revoked_at IS NULL AND expires_at > NOW();

-- name: RevokeUserRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1;

-- name: UpdateEmailandPassword :one
UPDATE users
SET email = $1, hashed_password = $2
WHERE id = $3
RETURNING *;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;