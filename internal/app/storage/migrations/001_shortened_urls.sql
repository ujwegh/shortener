-- +goose Up
-- +goose StatementBegin
CREATE TABLE shortened_urls
(
    uuid         UUID PRIMARY KEY,
    short_url    VARCHAR UNIQUE NOT NULL,
    original_url VARCHAR        NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE shortened_urls;
-- +goose StatementEnd
