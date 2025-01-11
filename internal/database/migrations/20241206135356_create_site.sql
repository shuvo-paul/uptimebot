-- +migrate Up
CREATE TABLE sites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    status VARCHAR(50),
    enabled BOOLEAN DEFAULT true,
    interval FLOAT NOT NULL,
    status_changed_at TEXT,
    FOREIGN KEY (user_id) REFERENCES users (id)
);

-- +migrate Down
DROP TABLE sites;