package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/outbox"
)

type providerCloser interface {
	ClosePayment(ctx context.Context, providerPaymentNo string) error
}

type providerCloseHandler struct {
	gateway providerCloser
}

func (h *providerCloseHandler) Handle(ctx context.Context, event outbox.Event) error {
	var payload struct {
		ProviderTransactionID string `json:"provider_transaction_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return outbox.Permanent(fmt.Errorf("decode payment close event: %w", err))
	}
	providerTransactionID := strings.TrimSpace(payload.ProviderTransactionID)
	if providerTransactionID == "" {
		return outbox.Permanent(fmt.Errorf("payment close event has no provider transaction"))
	}

	err := h.gateway.ClosePayment(ctx, providerTransactionID)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, apperr.ErrPaymentConflict),
		errors.Is(err, apperr.ErrPaymentProviderInvalid):
		return outbox.Permanent(err)
	default:
		return err
	}
}
