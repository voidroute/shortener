-- +goose Up
ALTER TABLE links ADD COLUMN expires_at TIMESTAMPTZ NULL;

-- +goose Down
ALTER TABLE links DROP COLUMN expires_at;