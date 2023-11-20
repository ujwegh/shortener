-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE shortened_urls
(
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    short_url    VARCHAR UNIQUE NOT NULL,
    original_url VARCHAR        NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE shortened_urls;
-- +goose StatementEnd
