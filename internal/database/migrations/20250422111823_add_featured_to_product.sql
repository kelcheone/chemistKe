-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
ALTER TABLE products
ADD COLUMN featured BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
ALTER TABLE products
DROP COLUMN featured;
