-- +goose Up
-- +goose StatementBegin
SELECT
    'up SQL query';

-- +goose StatementEnd
CREATE TABLE product_reviews (
    id UUID NOT NULL DEFAULT gen_random_uuid () PRIMARY KEY,
    product_id UUID NOT NULL,
    user_id UUID,
    title TEXT NOT NULL,
    rating INT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        -- foreign key constraint
        FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
        FOREIGN KEY (user_id) REFERENCES users (id)
);

-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
DROP TABLE product_reviews;
