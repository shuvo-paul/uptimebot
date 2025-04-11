-- +migrate Up
CREATE TABLE usr (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(email)
);

CREATE INDEX idx_usr_email ON usr(email);

-- +migrate Down
DROP INDEX IF EXISTS idx_usr_email;
DROP TABLE usr;
