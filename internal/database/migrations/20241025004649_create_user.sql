-- +migrate Up
CREATE TABLE user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    UNIQUE(email)
);

CREATE INDEX idx_user_email ON user(email);

-- +migrate Down
DROP INDEX IF EXISTS idx_user_email;
DROP TABLE user;
