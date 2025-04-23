-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
-- first removethe category COLUMN
ALTER TABLE products
DROP COLUMN IF EXISTS category;

ALTER TABLE products
ADD COLUMN category_id UUID NOT NULL REFERENCES product_category (id);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
ALTER TABLE products
DROP COLUMN category_id;
