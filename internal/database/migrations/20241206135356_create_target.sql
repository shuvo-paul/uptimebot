-- +migrate Up
CREATE TABLE target (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    status TEXT,
    enabled BOOLEAN DEFAULT true,
    interval FLOAT NOT NULL,
    changed_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE target;