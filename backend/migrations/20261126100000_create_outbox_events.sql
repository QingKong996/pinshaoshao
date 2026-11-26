-- +goose Up
CREATE TABLE outbox_events (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_key TEXT NOT NULL UNIQUE CHECK (btrim(event_key) <> ''),
    topic TEXT NOT NULL CHECK (btrim(topic) <> ''),
    aggregate_type TEXT NOT NULL CHECK (btrim(aggregate_type) <> ''),
    aggregate_id BIGINT NOT NULL CHECK (aggregate_id > 0),
    payload JSONB NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'published', 'dead')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_until TIMESTAMPTZ,
    last_error TEXT,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK ((status = 'published') = (published_at IS NOT NULL))
);

CREATE INDEX idx_outbox_events_due
    ON outbox_events(status, available_at, id)
    WHERE status IN ('pending', 'processing');

-- +goose Down
DROP INDEX idx_outbox_events_due;
DROP TABLE outbox_events;
