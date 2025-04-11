-- +migrate Up
CREATE TABLE session (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES usr(id) ON DELETE CASCADE,
    UNIQUE (token)
);

-- +migrate Down
DROP TABLE session;