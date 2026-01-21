-- +goose Up
CREATE TABLE IF NOT EXISTS telegram_integrations (
    user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    chat_id BIGINT NOT NULL, 
    username TEXT,          
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS telegram_link_tokens (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS telegram_link_tokens;
DROP TABLE IF EXISTS telegram_integrations;