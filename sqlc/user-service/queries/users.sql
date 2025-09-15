-- name: CreateUser :one
INSERT INTO users (
    email,
    phone_number,
    name,
    password_hash
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND is_active = true;

-- name: GetUserByID :one
SELECT * FROM users
WHERE user_id = $1 AND is_active = true;

-- name: UpdateUser :one
UPDATE users
SET 
    name = COALESCE($2, name),
    phone_number = COALESCE($3, phone_number),
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND is_active = true
RETURNING *;

-- name: DeactivateUser :exec
UPDATE users
SET 
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET
    password_hash = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1;

-- name: CheckUserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND is_active = true) as exists;