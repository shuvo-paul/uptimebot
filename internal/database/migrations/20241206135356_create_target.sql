-- +migrate Up
CREATE TABLE target (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    status TEXT,
    enabled BOOLEAN DEFAULT true,
    interval FLOAT NOT NULL,
    changed_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES usr (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE target;