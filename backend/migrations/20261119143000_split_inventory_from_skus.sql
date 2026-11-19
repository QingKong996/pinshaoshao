-- +goose Up
CREATE TABLE inventory_items (
    sku_id INT8 PRIMARY KEY REFERENCES product_skus(id) ON DELETE RESTRICT,
    total_stock INTEGER NOT NULL DEFAULT 0 CHECK (total_stock >= 0),
    frozen_stock INTEGER NOT NULL DEFAULT 0 CHECK (frozen_stock >= 0),
    version BIGINT NOT NULL DEFAULT 0 CHECK (version >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT inventory_stock_bounds CHECK (frozen_stock <= total_stock)
);

INSERT INTO inventory_items (sku_id, total_stock, frozen_stock)
SELECT id, stock, 0
FROM product_skus;

ALTER TABLE product_skus DROP COLUMN stock;

-- +goose Down
ALTER TABLE product_skus
    ADD COLUMN stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0);

UPDATE product_skus ps
SET stock = ii.total_stock
FROM inventory_items ii
WHERE ii.sku_id = ps.id;

DROP TABLE inventory_items;
