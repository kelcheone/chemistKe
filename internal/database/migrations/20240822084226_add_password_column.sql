-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
ALTER TABLE users DROP COLUMN password;
