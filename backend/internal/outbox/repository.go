package outbox

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Claim(ctx context.Context, limit int, lease time.Duration) ([]Event, error) {
	if limit <= 0 {
		return nil, nil
	}

	const query = `WITH due AS (
    SELECT id
    FROM outbox_events
    WHERE (status = 'pending' AND available_at <= now())
       OR (status = 'processing' AND locked_until < now())
    ORDER BY id
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
UPDATE outbox_events AS events
SET status = 'processing',
    attempts = events.attempts + 1,
    locked_until = now() + make_interval(secs => $2),
    updated_at = now()
FROM due
WHERE events.id = due.id
RETURNING events.id, events.event_key, events.topic, events.aggregate_type,
          events.aggregate_id, events.payload, events.attempts`

	rows, err := r.pool.Query(ctx, query, limit, int(lease.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("claim outbox events: %w", err)
	}
	defer rows.Close()

	events := make([]Event, 0, limit)
	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.ID,
			&event.Key,
			&event.Topic,
			&event.AggregateType,
			&event.AggregateID,
			&event.Payload,
			&event.Attempts,
		); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read outbox events: %w", err)
	}
	return events, nil
}

func (r *Repository) MarkPublished(ctx context.Context, eventID int64, attempts int) error {
	const query = `UPDATE outbox_events
SET status = 'published',
    published_at = now(),
    locked_until = NULL,
    last_error = NULL,
    updated_at = now()
WHERE id = $1 AND attempts = $2 AND status = 'processing'`
	if _, err := r.pool.Exec(ctx, query, eventID, attempts); err != nil {
		return fmt.Errorf("mark outbox event published: %w", err)
	}
	return nil
}

func (r *Repository) MarkFailed(
	ctx context.Context,
	eventID int64,
	attempts int,
	maxAttempts int,
	availableAt time.Time,
	cause string,
	permanent bool,
) error {
	const query = `UPDATE outbox_events
SET status = CASE WHEN $2 OR attempts >= $3 THEN 'dead' ELSE 'pending' END,
    available_at = $4,
    locked_until = NULL,
    last_error = $5,
    updated_at = now()
WHERE id = $1 AND attempts = $6 AND status = 'processing'`
	if _, err := r.pool.Exec(
		ctx,
		query,
		eventID,
		permanent,
		maxAttempts,
		availableAt,
		cause,
		attempts,
	); err != nil {
		return fmt.Errorf("mark outbox event failed: %w", err)
	}
	return nil
}
