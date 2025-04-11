-- +migrate Up
CREATE TABLE token (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    FOREIGN KEY(user_id) REFERENCES usr(id) ON DELETE CASCADE
);

CREATE INDEX idx_token_token ON token(token);

-- +migrate Down
DROP INDEX idx_token_token;
DROP TABLE token;