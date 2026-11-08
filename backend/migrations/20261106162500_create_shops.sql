-- +goose Up
CREATE TYPE shop_status AS ENUM (
    'active',
    'banned'
    );

CREATE TABLE shops(
    id INT8 GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(20) UNIQUE NOT NULL CHECK (
            char_length(username) >= 5
        ),
    password_hash TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    status shop_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()

);

CREATE TABLE shop_sessions(
    token_hash TEXT PRIMARY KEY,
    shop_id INT8 NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '7 days')
);
CREATE INDEX idx_shop_sessions_shop_id ON shop_sessions(shop_id);

-- +goose Down
DROP INDEX idx_shop_sessions_shop_id;
DROP TABLE shop_sessions;
DROP TABLE shops;
DROP TYPE shop_status;
