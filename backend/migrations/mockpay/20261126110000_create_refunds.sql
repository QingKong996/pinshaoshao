-- +goose Up
CREATE TYPE mock_refund_status AS ENUM ('succeeded');

CREATE TABLE mock_payment_refunds (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    refund_no UUID NOT NULL UNIQUE,
    provider_refund_no UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    payment_id BIGINT NOT NULL UNIQUE REFERENCES mock_payment_orders(id) ON DELETE CASCADE,
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    reason VARCHAR(200) NOT NULL CHECK (btrim(reason) <> ''),
    status mock_refund_status NOT NULL DEFAULT 'succeeded',
    refunded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE mock_payment_refunds;
DROP TYPE mock_refund_status;
