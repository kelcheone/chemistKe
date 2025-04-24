-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
ALTER TABLE product_category
ALTER COLUMN slug
SET
    NOT NULL,
    ADD CONSTRAINT unique_category_slug UNIQUE (slug);

ALTER TABLE product_sub_category
ALTER COLUMN slug
SET
    NOT NULL,
    ADD CONSTRAINT unique_sub_category_slug UNIQUE (slug);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
