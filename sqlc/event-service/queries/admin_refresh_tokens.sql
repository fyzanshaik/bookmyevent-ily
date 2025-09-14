
-- name: CreateAdminRefreshToken :one
INSERT INTO admin_refresh_tokens (
    token, admin_id, expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetAdminRefreshToken :one
SELECT * FROM admin_refresh_tokens
WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL;

-- name: RevokeAdminRefreshToken :exec
UPDATE admin_refresh_tokens
SET
    revoked_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE token = $1;

-- name: RevokeAllAdminTokens :exec
UPDATE admin_refresh_tokens
SET
    revoked_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE admin_id = $1 AND revoked_at IS NULL;

-- name: CleanupExpiredAdminTokens :exec
DELETE FROM admin_refresh_tokens
WHERE expires_at < CURRENT_TIMESTAMP OR revoked_at IS NOT NULL;
