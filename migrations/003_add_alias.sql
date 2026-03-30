-- +goose Up
ALTER TABLE links ADD COLUMN alias VARCHAR(16) NULL;
CREATE UNIQUE INDEX idx_links_alias ON links (alias) WHERE alias IS NOT NULL;

-- +goose Down
DROP INDEX idx_links_alias;
ALTER TABLE links DROP COLUMN alias;