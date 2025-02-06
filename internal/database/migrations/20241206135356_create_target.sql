-- +migrate Up
CREATE TABLE target (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    status VARCHAR(50),
    enabled BOOLEAN DEFAULT true,
    interval FLOAT NOT NULL,
    status_changed_at TEXT,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE target;