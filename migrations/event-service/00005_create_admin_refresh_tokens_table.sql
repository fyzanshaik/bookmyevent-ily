-- +goose Up
-- +goose StatementBegin
CREATE TABLE admin_refresh_tokens (
    token TEXT PRIMARY KEY,
    admin_id UUID NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin
        FOREIGN KEY(admin_id)
        REFERENCES admins(admin_id)
        ON DELETE CASCADE
);

-- Performance indexes for admin refresh token operations
CREATE INDEX idx_admin_refresh_tokens_admin_id ON admin_refresh_tokens(admin_id);
CREATE INDEX idx_admin_refresh_tokens_expires ON admin_refresh_tokens(expires_at) WHERE revoked_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admin_refresh_tokens;
-- +goose StatementEnd

