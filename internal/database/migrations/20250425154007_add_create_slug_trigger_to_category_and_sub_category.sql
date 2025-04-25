-- +goose Up
-- +goose StatementBegin

-- Create function for category slug
CREATE OR REPLACE FUNCTION set_category_slug()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.slug IS NULL OR TRIM(NEW.slug) = '' THEN
    NEW.slug := LOWER(REPLACE(NEW.name, ' ', '-'));
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for category
CREATE TRIGGER create_category_slug
BEFORE INSERT ON product_category
FOR EACH ROW
EXECUTE FUNCTION set_category_slug();

-- Create function for subcategory slug
CREATE OR REPLACE FUNCTION set_sub_category_slug()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.slug IS NULL OR TRIM(NEW.slug) = '' THEN
    NEW.slug := LOWER(REPLACE(NEW.name, ' ', '-'));
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for subcategory
CREATE TRIGGER create_sub_category_slug
BEFORE INSERT ON product_sub_category
FOR EACH ROW
EXECUTE FUNCTION set_sub_category_slug();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS create_category_slug ON product_category;
DROP FUNCTION IF EXISTS set_category_slug();

DROP TRIGGER IF EXISTS create_sub_category_slug ON product_sub_category;
DROP FUNCTION IF EXISTS set_sub_category_slug();

-- +goose StatementEnd
