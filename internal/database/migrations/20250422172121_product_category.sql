-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
CREATE TABLE product_category (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    slug VARCHAR(255) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW ()
);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
DROP TABLE product_category;
