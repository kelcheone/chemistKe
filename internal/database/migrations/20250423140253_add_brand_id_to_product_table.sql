-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
ALTER TABLE products
DROP COLUMN IF EXISTS brand;

ALTER TABLE products
ADD COLUMN brand_id UUID NOT NULL REFERENCES product_brand (id);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
ALTER TABLE products
DROP COLUMN brand_id;
