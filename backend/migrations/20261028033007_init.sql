-- +goose Up
CREATE TYPE user_status AS ENUM (
    'active',
    'banned'
    );

CREATE TABLE users(
    id INT8 GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username VARCHAR(20) UNIQUE NOT NULL CHECK (
            char_length(username) >= 5
        ),
    password_hash TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()

);

CREATE TABLE sessions(
    token_hash TEXT PRIMARY KEY,
    user_id INT8 NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '7 days')
);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- +goose Down
DROP INDEX idx_sessions_user_id;
DROP TABLE sessions;
DROP TABLE users;
DROP TYPE user_status;