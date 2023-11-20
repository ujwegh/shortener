-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS original_urls_unique_idx ON shortened_urls (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS original_urls_unique_idx;
-- +goose StatementEnd
