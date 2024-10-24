-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    UNIQUE(email)
);

CREATE INDEX idx_users_email ON users(email);

-- +migrate Down
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE users;
