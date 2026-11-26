package outbox

import (
	"context"
	"encoding/json"
)

const TopicPaymentCloseProvider = "payment.close_provider"
const TopicPaymentRefundProvider = "payment.refund_provider"

type Event struct {
	ID            int64
	Key           string
	Topic         string
	AggregateType string
	AggregateID   int64
	Payload       json.RawMessage
	Attempts      int
}

type Handler interface {
	Handle(ctx context.Context, event Event) error
}

type HandlerFunc func(ctx context.Context, event Event) error

func (f HandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}
