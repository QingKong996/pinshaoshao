-- +goose Up
CREATE TABLE user_addresses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_name VARCHAR(50) NOT NULL CHECK (btrim(receiver_name) <> ''),
    receiver_phone VARCHAR(32) NOT NULL CHECK (btrim(receiver_phone) <> ''),
    shipping_address TEXT NOT NULL CHECK (btrim(shipping_address) <> ''),
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_user_addresses_default
    ON user_addresses(user_id)
    WHERE is_default;
CREATE INDEX idx_user_addresses_user_updated_at
    ON user_addresses(user_id, updated_at DESC, id DESC);

-- +goose Down
DROP INDEX idx_user_addresses_user_updated_at;
DROP INDEX idx_user_addresses_default;
DROP TABLE user_addresses;
