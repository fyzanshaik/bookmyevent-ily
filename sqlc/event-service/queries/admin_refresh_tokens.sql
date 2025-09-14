-- Admin Refresh Token Management Queries

-- CREATE ADMIN REFRESH TOKEN
-- name: CreateAdminRefreshToken :one
INSERT INTO admin_refresh_tokens (
    token, admin_id, expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- GET ADMIN REFRESH TOKEN (for validation)
-- name: GetAdminRefreshToken :one
SELECT * FROM admin_refresh_tokens
WHERE token = $1 AND expires_at > CURRENT_TIMESTAMP AND revoked_at IS NULL;

-- REVOKE ADMIN REFRESH TOKEN (logout)
-- name: RevokeAdminRefreshToken :exec
UPDATE admin_refresh_tokens
SET 
    revoked_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE token = $1;

-- REVOKE ALL ADMIN TOKENS (security action)
-- name: RevokeAllAdminTokens :exec
UPDATE admin_refresh_tokens
SET 
    revoked_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE admin_id = $1 AND revoked_at IS NULL;

-- CLEANUP EXPIRED ADMIN TOKENS (maintenance)
-- name: CleanupExpiredAdminTokens :exec
DELETE FROM admin_refresh_tokens
WHERE expires_at < CURRENT_TIMESTAMP OR revoked_at IS NOT NULL;
