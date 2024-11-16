-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE authors ADD CONSTRAINT unique_user_id UNIQUE (user_id);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
ALTER TABLE authors DROP CONSTRAINT unique_user_id;
