-- +migrate Up
CREATE TABLE notifiers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{}',
    FOREIGN KEY (site_id) REFERENCES sites (id) ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE IF EXISTS notifiers;
