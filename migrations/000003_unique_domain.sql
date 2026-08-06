-- +goose Up
ALTER TABLE projects ADD CONSTRAINT unique_domain UNIQUE (domain);

-- +goose Down
ALTER TABLE projects DROP CONSTRAINT IF EXISTS unique_domain;
