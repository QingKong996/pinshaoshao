-- +goose Up
CREATE TYPE payment_status AS ENUM ('pending', 'succeeded', 'closed');

CREATE TABLE payments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    payment_no UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    order_id BIGINT NOT NULL UNIQUE,
    provider VARCHAR(32) NOT NULL CHECK (btrim(provider) <> ''),
    provider_transaction_id VARCHAR(128),
    status payment_status NOT NULL DEFAULT 'pending',
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (order_id, amount_cents)
        REFERENCES orders(id, total_amount_cents),
    UNIQUE (provider, provider_transaction_id),
    CHECK ((status = 'succeeded') = (paid_at IS NOT NULL))
);

CREATE INDEX idx_payments_pending ON payments(updated_at) WHERE status = 'pending';

-- +goose Down
DROP INDEX idx_payments_pending;
DROP TABLE payments;
DROP TYPE payment_status;
