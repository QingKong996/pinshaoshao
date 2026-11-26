package outbox

import (
	"context"
	"errors"
	"fmt"

	"pinshaoshao/internal/database"
)

var ErrInvalidEvent = errors.New("invalid outbox event")

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Enqueue(ctx context.Context, db database.DBTX, event Event) error {
	if event.Key == "" || event.Topic == "" || event.AggregateType == "" ||
		event.AggregateID <= 0 || len(event.Payload) == 0 {
		return ErrInvalidEvent
	}

	const query = `INSERT INTO outbox_events
    (event_key, topic, aggregate_type, aggregate_id, payload)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (event_key) DO NOTHING`
	if _, err := db.Exec(
		ctx,
		query,
		event.Key,
		event.Topic,
		event.AggregateType,
		event.AggregateID,
		event.Payload,
	); err != nil {
		return fmt.Errorf("enqueue outbox event: %w", err)
	}
	return nil
}
