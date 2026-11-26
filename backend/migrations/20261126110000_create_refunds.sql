-- +goose Up
CREATE TYPE refund_status AS ENUM ('pending', 'succeeded', 'failed');

CREATE TABLE refunds (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    refund_no UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    payment_id BIGINT NOT NULL UNIQUE REFERENCES payments(id),
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    reason VARCHAR(200) NOT NULL CHECK (btrim(reason) <> ''),
    status refund_status NOT NULL DEFAULT 'pending',
    provider_refund_no VARCHAR(128) UNIQUE,
    refunded_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((status = 'succeeded') = (refunded_at IS NOT NULL)),
    CHECK ((status = 'failed') = (failed_at IS NOT NULL)),
    CHECK (status <> 'failed' OR last_error IS NOT NULL)
);

CREATE INDEX idx_refunds_pending
    ON refunds(created_at, id)
    WHERE status = 'pending';

-- +goose Down
DROP INDEX idx_refunds_pending;
DROP TABLE refunds;
DROP TYPE refund_status;
