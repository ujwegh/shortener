-- +goose Up
-- +goose StatementBegin

create table if not exists user_urls
(
    uuid               uuid not null,
    shortened_url_uuid uuid not null references shortened_urls (uuid) on delete cascade,
    unique (uuid, shortened_url_uuid)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists user_urls;

-- +goose StatementEnd
