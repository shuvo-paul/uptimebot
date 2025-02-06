-- +migrate Up
CREATE TABLE session (
    user_id INT NOT NULL,
    token VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id),
    UNIQUE (token)
);

-- +migrate Down
DROP TABLE session;