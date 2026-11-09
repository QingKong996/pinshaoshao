-- +goose Up
ALTER TABLE shops
    ADD COLUMN name VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN description TEXT NOT NULL DEFAULT '';
UPDATE shops SET name = username WHERE name = '';

CREATE TABLE product_categories(
    id INT8 GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shop_id INT8 NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (shop_id, name)
);

ALTER TABLE products
    ADD COLUMN category_id INT8 REFERENCES product_categories(id) ON DELETE SET NULL;
CREATE INDEX idx_products_category_id ON products(category_id);

-- +goose Down
DROP INDEX idx_products_category_id;
ALTER TABLE products DROP COLUMN category_id;
DROP TABLE product_categories;
ALTER TABLE shops DROP COLUMN description, DROP COLUMN name;
