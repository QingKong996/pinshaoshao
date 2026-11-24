-- +goose Up
CREATE TYPE mock_payment_status AS ENUM ('pending', 'paid', 'closed');

CREATE TYPE mock_payment_notification_status AS ENUM (
    'pending',
    'delivering',
    'succeeded',
    'exhausted'
);

CREATE TABLE mock_payment_orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    payment_no UUID NOT NULL UNIQUE,
    merchant_order_no VARCHAR(64) NOT NULL UNIQUE CHECK (btrim(merchant_order_no) <> ''),
    request_hash VARCHAR(64) NOT NULL CHECK (request_hash ~ '^[0-9a-f]{64}$'),
    status mock_payment_status NOT NULL DEFAULT 'pending',
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    currency CHAR(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    subject VARCHAR(200) NOT NULL CHECK (btrim(subject) <> ''),
    notify_url TEXT NOT NULL CHECK (btrim(notify_url) <> ''),
    return_url TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CHECK (expires_at > created_at),
    CHECK (
        (status = 'pending' AND paid_at IS NULL AND closed_at IS NULL)
        OR (status = 'paid' AND paid_at IS NOT NULL AND closed_at IS NULL)
        OR (status = 'closed' AND paid_at IS NULL AND closed_at IS NOT NULL)
    )
);

CREATE TABLE mock_payment_notifications (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    payment_id BIGINT NOT NULL REFERENCES mock_payment_orders(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    target_url TEXT NOT NULL CHECK (btrim(target_url) <> ''),
    payload JSONB NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    status mock_payment_notification_status NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts INTEGER NOT NULL CHECK (max_attempts > 0),
    next_attempt_at TIMESTAMPTZ NOT NULL,
    locked_until TIMESTAMPTZ,
    last_attempt_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (payment_id, event_type),
    CHECK (
        (status IN ('pending', 'delivering') AND delivered_at IS NULL)
        OR (status = 'succeeded' AND delivered_at IS NOT NULL)
        OR (status = 'exhausted' AND delivered_at IS NULL AND last_error IS NOT NULL)
    )
);

CREATE INDEX idx_mock_payment_orders_pending_expiry
    ON mock_payment_orders(expires_at, id)
    WHERE status = 'pending';
CREATE INDEX idx_mock_payment_notifications_due
    ON mock_payment_notifications(status, next_attempt_at, id)
    WHERE status IN ('pending', 'delivering');
CREATE INDEX idx_mock_payment_notifications_payment
    ON mock_payment_notifications(payment_id, id);

-- +goose Down
DROP INDEX idx_mock_payment_notifications_payment;
DROP INDEX idx_mock_payment_notifications_due;
DROP INDEX idx_mock_payment_orders_pending_expiry;
DROP TABLE mock_payment_notifications;
DROP TABLE mock_payment_orders;
DROP TYPE mock_payment_notification_status;
DROP TYPE mock_payment_status;
