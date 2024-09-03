-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE authors(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bio TEXT NOT NULL,
    avatar VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    user_id UUID REFERENCES users(id)
);

CREATE TABLE categories(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE content(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    published_date TIMESTAMP NOT NULL,
    updated_date TIMESTAMP NOT NULL,
    cover_image VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    slug VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(255) NOT NULL,
    author_id UUID REFERENCES authors(id),
    category_id UUID REFERENCES categories(id)
);

CREATE INDEX content_author_id_index ON content(author_id);
CREATE INDEX content_category_id_index ON content(category_id);
CREATE INDEX content_slug_index ON content(slug);
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
DROP TABLE content;
DROP TABLE authors;
DROP TABLE categories;
