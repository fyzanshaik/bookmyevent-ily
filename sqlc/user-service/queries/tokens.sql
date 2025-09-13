-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    token,
    user_id,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET 
    revoked_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE token = $1;

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET 
    revoked_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: CleanupExpiredTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < CURRENT_TIMESTAMP OR revoked_at IS NOT NULL;