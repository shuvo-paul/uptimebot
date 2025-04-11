-- +migrate Up
CREATE TABLE notifier (
    id SERIAL PRIMARY KEY,
    target_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    config JSONB NOT NULL,
    FOREIGN KEY (target_id) REFERENCES target (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS notifier;
