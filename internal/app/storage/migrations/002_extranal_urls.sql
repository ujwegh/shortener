-- +goose Up
-- +goose StatementBegin
ALTER TABLE shortened_urls
    add column if not exists correlation_id varchar;
CREATE UNIQUE INDEX IF NOT EXISTS shortened_urls_correlation_id_idx ON shortened_urls (correlation_id)
    WHERE correlation_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS shortened_urls_correlation_id_idx;
ALTER TABLE shortened_urls
    drop column if exists correlation_id;
-- +goose StatementEnd
