-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
CREATE TABLE product_brand (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW (),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW ()
);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
DROP TABLE product_brand;
