-- +goose Up
CREATE TYPE mock_shipment_status AS ENUM (
    'pending_outbound',
    'pending_pickup',
    'picked_up',
    'in_transit',
    'at_transit_hub',
    'out_for_delivery',
    'delivered'
);

CREATE TYPE mock_logistics_notification_status AS ENUM (
    'pending',
    'delivering',
    'succeeded',
    'exhausted'
);

CREATE TABLE mock_shipments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shipment_no UUID NOT NULL UNIQUE,
    merchant_order_no VARCHAR(64) NOT NULL UNIQUE CHECK (btrim(merchant_order_no) <> ''),
    request_hash VARCHAR(64) NOT NULL CHECK (request_hash ~ '^[0-9a-f]{64}$'),
    tracking_no VARCHAR(64) NOT NULL UNIQUE CHECK (tracking_no ~ '^[A-Z]{2,4}[0-9]{12}$'),
    carrier VARCHAR(64) NOT NULL CHECK (btrim(carrier) <> ''),
    status mock_shipment_status NOT NULL DEFAULT 'pending_outbound',
    item_summary VARCHAR(200) NOT NULL CHECK (btrim(item_summary) <> ''),
    sender_name VARCHAR(100) NOT NULL CHECK (btrim(sender_name) <> ''),
    sender_phone VARCHAR(32) NOT NULL CHECK (btrim(sender_phone) <> ''),
    sender_address VARCHAR(500) NOT NULL CHECK (btrim(sender_address) <> ''),
    receiver_name VARCHAR(100) NOT NULL CHECK (btrim(receiver_name) <> ''),
    receiver_phone VARCHAR(32) NOT NULL CHECK (btrim(receiver_phone) <> ''),
    receiver_address VARCHAR(500) NOT NULL CHECK (btrim(receiver_address) <> ''),
    origin_city VARCHAR(64) NOT NULL CHECK (btrim(origin_city) <> ''),
    destination_city VARCHAR(64) NOT NULL CHECK (btrim(destination_city) <> ''),
    notify_url TEXT NOT NULL CHECK (btrim(notify_url) <> ''),
    current_location TEXT NOT NULL CHECK (btrim(current_location) <> ''),
    dispatched_at TIMESTAMPTZ,
    picked_up_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (updated_at >= created_at),
    CHECK (
        (status = 'pending_outbound'
            AND dispatched_at IS NULL
            AND picked_up_at IS NULL
            AND delivered_at IS NULL)
        OR (status = 'pending_pickup'
            AND dispatched_at IS NOT NULL
            AND picked_up_at IS NULL
            AND delivered_at IS NULL)
        OR (status = 'picked_up'
            AND dispatched_at IS NOT NULL
            AND picked_up_at IS NOT NULL
            AND delivered_at IS NULL)
        OR (status IN ('in_transit', 'at_transit_hub', 'out_for_delivery')
            AND dispatched_at IS NOT NULL
            AND picked_up_at IS NOT NULL
            AND delivered_at IS NULL)
        OR (status = 'delivered'
            AND dispatched_at IS NOT NULL
            AND picked_up_at IS NOT NULL
            AND delivered_at IS NOT NULL)
    )
);

CREATE TABLE mock_logistics_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    shipment_id BIGINT NOT NULL REFERENCES mock_shipments(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL CHECK (sequence > 0),
    status mock_shipment_status NOT NULL,
    event_type VARCHAR(64) NOT NULL CHECK (btrim(event_type) <> ''),
    description TEXT NOT NULL CHECK (btrim(description) <> ''),
    location TEXT NOT NULL CHECK (btrim(location) <> ''),
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (shipment_id, sequence)
);

CREATE TABLE mock_logistics_notifications (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    shipment_id BIGINT NOT NULL REFERENCES mock_shipments(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    target_url TEXT NOT NULL CHECK (btrim(target_url) <> ''),
    payload JSONB NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    status mock_logistics_notification_status NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts INTEGER NOT NULL CHECK (max_attempts > 0),
    next_attempt_at TIMESTAMPTZ NOT NULL,
    locked_until TIMESTAMPTZ,
    last_attempt_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (shipment_id, event_type),
    CHECK (
        (status IN ('pending', 'delivering') AND delivered_at IS NULL)
        OR (status = 'succeeded' AND delivered_at IS NOT NULL)
        OR (status = 'exhausted' AND delivered_at IS NULL AND last_error IS NOT NULL)
    )
);

CREATE TABLE mock_auth_accounts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE CHECK (btrim(username) <> ''),
    password_hash TEXT NOT NULL CHECK (btrim(password_hash) <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE mock_auth_sessions (
    token_hash BYTEA PRIMARY KEY CHECK (octet_length(token_hash) = 32),
    account_id BIGINT NOT NULL REFERENCES mock_auth_accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    CHECK (expires_at > created_at)
);

CREATE INDEX idx_mock_logistics_events_shipment
    ON mock_logistics_events(shipment_id, sequence);
CREATE INDEX idx_mock_logistics_notifications_due
    ON mock_logistics_notifications(status, next_attempt_at, id)
    WHERE status IN ('pending', 'delivering');
CREATE INDEX idx_mock_logistics_notifications_shipment
    ON mock_logistics_notifications(shipment_id, id);
CREATE INDEX idx_mock_auth_sessions_account
    ON mock_auth_sessions(account_id);
CREATE INDEX idx_mock_auth_sessions_expiry
    ON mock_auth_sessions(expires_at);

-- +goose Down
DROP INDEX idx_mock_auth_sessions_expiry;
DROP INDEX idx_mock_auth_sessions_account;
DROP INDEX idx_mock_logistics_notifications_shipment;
DROP INDEX idx_mock_logistics_notifications_due;
DROP INDEX idx_mock_logistics_events_shipment;
DROP TABLE mock_auth_sessions;
DROP TABLE mock_auth_accounts;
DROP TABLE mock_logistics_notifications;
DROP TABLE mock_logistics_events;
DROP TABLE mock_shipments;
DROP TYPE mock_logistics_notification_status;
DROP TYPE mock_shipment_status;
