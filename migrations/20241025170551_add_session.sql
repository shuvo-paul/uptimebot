-- +migrate Up
CREATE TABLE sessions (
    user_id INT NOT NULL,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE (token)
);

-- +migrate Down
DROP TABLE sessions;