-- +goose Up
CREATE TYPE shipment_status AS ENUM (
    'pending_pickup',
    'picked_up',
    'in_transit',
    'at_transit_hub',
    'out_for_delivery',
    'delivered'
);

CREATE TABLE shipments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shipment_no UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    sub_order_id BIGINT NOT NULL UNIQUE REFERENCES sub_orders(id),
    order_id BIGINT NOT NULL REFERENCES orders(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    shop_id BIGINT NOT NULL REFERENCES shops(id),
    provider VARCHAR(32) NOT NULL CHECK (btrim(provider) <> ''),
    provider_shipment_no UUID NOT NULL UNIQUE,
    tracking_no VARCHAR(100) NOT NULL CHECK (btrim(tracking_no) <> ''),
    carrier VARCHAR(64) NOT NULL CHECK (btrim(carrier) <> ''),
    status shipment_status NOT NULL,
    item_summary VARCHAR(200) NOT NULL CHECK (btrim(item_summary) <> ''),
    receiver_name VARCHAR(100) NOT NULL CHECK (btrim(receiver_name) <> ''),
    receiver_phone VARCHAR(32) NOT NULL CHECK (btrim(receiver_phone) <> ''),
    receiver_address TEXT NOT NULL CHECK (btrim(receiver_address) <> ''),
    origin_city VARCHAR(64) NOT NULL DEFAULT '',
    destination_city VARCHAR(64) NOT NULL DEFAULT '',
    current_location TEXT NOT NULL DEFAULT '',
    shipped_at TIMESTAMPTZ NOT NULL,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((status = 'delivered') = (delivered_at IS NOT NULL))
);

CREATE TABLE shipment_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shipment_id BIGINT NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    event_id UUID NOT NULL UNIQUE,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    status shipment_status NOT NULL,
    event_type VARCHAR(64) NOT NULL CHECK (btrim(event_type) <> ''),
    description TEXT NOT NULL CHECK (btrim(description) <> ''),
    location TEXT NOT NULL CHECK (btrim(location) <> ''),
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (shipment_id, sequence)
);

CREATE INDEX idx_shipments_shop_created_at
    ON shipments(shop_id, created_at DESC, id DESC);
CREATE INDEX idx_shipments_user_order
    ON shipments(user_id, order_id, created_at DESC);
CREATE INDEX idx_shipment_events_shipment
    ON shipment_events(shipment_id, sequence);

-- +goose Down
DROP INDEX idx_shipment_events_shipment;
DROP INDEX idx_shipments_user_order;
DROP INDEX idx_shipments_shop_created_at;
DROP TABLE shipment_events;
DROP TABLE shipments;
DROP TYPE shipment_status;
