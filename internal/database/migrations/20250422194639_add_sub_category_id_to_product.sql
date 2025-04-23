-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
ALTER TABLE products
DROP COLUMN IF EXISTS sub_category;

ALTER TABLE products
ADD COLUMN sub_category_id UUID NOT NULL REFERENCES product_sub_category (id);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
ALTER TABLE products
DROP COLUMN IF EXISTS sub_category_id;
