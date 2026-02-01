-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN is_verified BOOLEAN DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS verification_tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS verificaion_tokens;
ALTER TABLE users DROP COLUMN is_verified;
-- +goose StatementEnd
