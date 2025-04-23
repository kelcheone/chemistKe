-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
CREATE TABLE product_sub_category (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name VARCHAR(255) NOT NULL,
    category_id UUID NOT NULL,
    description TEXT,
    FOREIGN KEY (category_id) REFERENCES product_category (id)
);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
DROP TABLE product_sub_category;
