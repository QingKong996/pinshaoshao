-- +goose Up
CREATE INDEX idx_sub_orders_shop_created_at
    ON sub_orders(shop_id, created_at DESC, id DESC);

-- +goose Down
DROP INDEX idx_sub_orders_shop_created_at;
