-- +goose Up
CREATE TABLE links (
    code VARCHAR(16) PRIMARY KEY,
    url VARCHAR(2048) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE links;