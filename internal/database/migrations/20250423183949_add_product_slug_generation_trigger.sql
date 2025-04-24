-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION generate_product_slug()
RETURNS TRIGGER AS $$
BEGIN
    -- Ensure NEW.name is not null to avoid errors, maybe trim whitespace first
    IF NEW.name IS NULL THEN
        -- Use 'product-' prefix if name is null
        NEW.slug := 'product-' || right(replace(NEW.id::text, '-', ''), 12);
    ELSE
        -- Generate slug from name: lowercase, replace whitespace with hyphen
        NEW.slug := lower(regexp_replace(trim(NEW.name), '\s+', '-', 'g'));
        -- Append the unique part of the ID
        NEW.slug := NEW.slug || '-' || right(replace(NEW.id::text, '-', ''), 12);
        -- Clean up potential multiple dashes or leading/trailing dashes
        NEW.slug := regexp_replace(NEW.slug, '-{2,}', '-', 'g'); -- Replace multiple dashes with one
        NEW.slug := trim(BOTH '-' FROM NEW.slug); -- Remove leading/trailing dashes
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create trigger (separate statement)
CREATE TRIGGER product_slug_generation_trigger
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION generate_product_slug();


-- +goose Down
-- Drop trigger first as it depends on the function
DROP TRIGGER IF EXISTS product_slug_generation_trigger ON products;

-- Then drop the function
DROP FUNCTION IF EXISTS generate_product_slug();
