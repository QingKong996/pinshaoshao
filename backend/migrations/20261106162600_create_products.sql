-- +goose Up
CREATE TYPE product_status AS ENUM (
    'on_shelf',
    'off_shelf'
    );

CREATE TABLE products(
    id INT8 GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shop_id INT8 NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status product_status NOT NULL DEFAULT 'off_shelf',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_shop_id ON products(shop_id);

CREATE TABLE product_skus(
    id INT8 GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id INT8 NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    specs JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(specs) = 'object'),
    price_cents INT8 NOT NULL CHECK (price_cents >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (product_id, specs)
);

-- +goose Down
DROP TABLE product_skus;
DROP INDEX idx_products_shop_id;
DROP TABLE products;
DROP TYPE product_status;
