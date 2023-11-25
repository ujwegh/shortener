-- +goose Up
-- +goose StatementBegin

alter table shortened_urls
    add column is_deleted boolean not null default false;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table shortened_urls
    drop column is_deleted;

-- +goose StatementEnd
