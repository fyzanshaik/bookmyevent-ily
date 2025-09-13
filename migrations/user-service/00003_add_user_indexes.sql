-- +goose Up
-- +goose StatementBegin
-- Index for fast email lookups (used in login)
CREATE INDEX idx_users_email ON users(email);

-- Index for phone number lookups (optional for user search)
CREATE INDEX idx_users_phone ON users(phone_number) WHERE phone_number IS NOT NULL;

-- Index for active users only (most queries filter by is_active)
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = true;

-- Index for refresh tokens by user_id (when revoking all user tokens)
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Index for active refresh tokens (not revoked) - we'll handle expiry in queries
CREATE INDEX idx_refresh_tokens_active ON refresh_tokens(expires_at, revoked_at) 
WHERE revoked_at IS NULL;

-- Index for expired/revoked tokens cleanup
CREATE INDEX idx_refresh_tokens_cleanup ON refresh_tokens(expires_at, revoked_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_active;
DROP INDEX IF EXISTS idx_refresh_tokens_cleanup;
-- +goose StatementEnd