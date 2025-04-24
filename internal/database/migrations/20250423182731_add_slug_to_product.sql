-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
ALTER TABLE products
ADD COLUMN slug VARCHAR(255) UNIQUE;

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
ALTER TABLE products
DROP COLUMN slug;
