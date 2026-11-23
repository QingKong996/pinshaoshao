-- +goose Up
CREATE TYPE order_status AS ENUM (
    'pending_payment',
    'paid',
    'partially_shipped',
    'shipped',
    'completed',
    'cancelled'
);

CREATE TYPE sub_order_status AS ENUM (
    'pending_payment',
    'awaiting_shipment',
    'shipped',
    'completed',
    'cancelled'
);

CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    checkout_key UUID NOT NULL,
    request_hash VARCHAR(64) NOT NULL CHECK (request_hash ~ '^[0-9a-f]{64}$'),
    status order_status NOT NULL DEFAULT 'pending_payment',
    total_amount_cents BIGINT NOT NULL DEFAULT 0 CHECK (total_amount_cents >= 0),
    receiver_name VARCHAR(50) NOT NULL CHECK (btrim(receiver_name) <> ''),
    receiver_phone VARCHAR(32) NOT NULL CHECK (btrim(receiver_phone) <> ''),
    shipping_address TEXT NOT NULL CHECK (btrim(shipping_address) <> ''),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, checkout_key),
    UNIQUE (id, total_amount_cents),
    CHECK ((status IN ('pending_payment', 'cancelled')) = (paid_at IS NULL))
);

CREATE TABLE sub_orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    shop_id BIGINT NOT NULL REFERENCES shops(id),
    shop_name VARCHAR(50) NOT NULL,
    status sub_order_status NOT NULL DEFAULT 'pending_payment',
    total_amount_cents BIGINT NOT NULL DEFAULT 0 CHECK (total_amount_cents >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (order_id, shop_id)
);

CREATE TABLE order_item_snapshots (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sub_order_id BIGINT NOT NULL REFERENCES sub_orders(id),
    product_id BIGINT NOT NULL,
    sku_id BIGINT NOT NULL,
    product_name VARCHAR(100) NOT NULL,
    sku_specs JSONB NOT NULL CHECK (jsonb_typeof(sku_specs) = 'object'),
    unit_price_cents BIGINT NOT NULL CHECK (unit_price_cents >= 0),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    line_amount_cents BIGINT GENERATED ALWAYS AS (unit_price_cents * quantity) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (sub_order_id, sku_id)
);

CREATE INDEX idx_orders_user_created_at ON orders(user_id, created_at DESC, id DESC);
CREATE INDEX idx_orders_user_pending_created_at ON orders(user_id, created_at) WHERE status = 'pending_payment';
CREATE INDEX idx_sub_orders_shop_status ON sub_orders(shop_id, status, created_at DESC);

-- +goose Down
DROP INDEX idx_sub_orders_shop_status;
DROP INDEX idx_orders_user_pending_created_at;
DROP INDEX idx_orders_user_created_at;
DROP TABLE order_item_snapshots;
DROP TABLE sub_orders;
DROP TABLE orders;
DROP TYPE sub_order_status;
DROP TYPE order_status;
