-- +migrate Up
CREATE TABLE notifier (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    config BLOB NOT NULL,
    FOREIGN KEY (target_id) REFERENCES target (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS notifier;
