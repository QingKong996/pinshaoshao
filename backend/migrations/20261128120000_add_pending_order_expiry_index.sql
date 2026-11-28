-- +goose Up
CREATE INDEX idx_orders_pending_due
    ON orders(created_at, id)
    WHERE status = 'pending_payment';

-- +goose Down
DROP INDEX idx_orders_pending_due;
