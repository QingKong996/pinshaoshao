package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/outbox"
)

type refundStore interface {
	GetForProcessing(ctx context.Context, refundNo string) (*model.Refund, error)
	MarkSucceeded(
		ctx context.Context,
		refundNo string,
		providerRefundNo string,
		refundedAt time.Time,
	) error
	MarkFailed(ctx context.Context, refundNo string, cause string) error
}

type refundGateway interface {
	RefundPayment(
		ctx context.Context,
		providerPaymentNo string,
		input model.ProviderRefundRequest,
	) (*model.ProviderRefundResult, error)
}

type providerRefundHandler struct {
	store   refundStore
	gateway refundGateway
}

func (h *providerRefundHandler) Handle(ctx context.Context, event outbox.Event) error {
	var payload struct {
		RefundNo string `json:"refund_no"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return outbox.Permanent(fmt.Errorf("decode payment refund event: %w", err))
	}
	refundNo := strings.TrimSpace(payload.RefundNo)
	if refundNo == "" {
		return outbox.Permanent(errors.New("payment refund event has no refund number"))
	}

	refund, err := h.store.GetForProcessing(ctx, refundNo)
	if err != nil {
		if errors.Is(err, apperr.ErrRefundNotFound) {
			return outbox.Permanent(err)
		}
		return err
	}
	switch refund.Status {
	case model.RefundStatusSucceeded, model.RefundStatusFailed:
		return nil
	}
	if refund.ProviderTransactionID == nil ||
		strings.TrimSpace(*refund.ProviderTransactionID) == "" {
		return h.failPermanently(ctx, refundNo, "payment has no provider transaction")
	}

	result, err := h.gateway.RefundPayment(
		ctx,
		*refund.ProviderTransactionID,
		model.ProviderRefundRequest{
			RefundNo:     refund.RefundNo,
			AmountCents:  refund.AmountCents,
			Reason:       refund.Reason,
			Currency:     model.CurrencyCNY,
			MerchantNo:   refund.PaymentNo,
			ProviderTxID: *refund.ProviderTransactionID,
		},
	)
	if err != nil {
		if isPermanentRefundError(err) {
			return h.failPermanently(ctx, refundNo, err.Error())
		}
		return err
	}
	if result == nil ||
		result.RefundNo != refund.RefundNo ||
		result.AmountCents != refund.AmountCents ||
		result.Status != "succeeded" ||
		strings.TrimSpace(result.ProviderRefundNo) == "" ||
		result.RefundedAt.IsZero() {
		return h.failPermanently(ctx, refundNo, "provider returned an invalid refund result")
	}

	if err := h.store.MarkSucceeded(
		ctx,
		refund.RefundNo,
		result.ProviderRefundNo,
		result.RefundedAt.UTC(),
	); err != nil {
		return err
	}
	return nil
}

func (h *providerRefundHandler) failPermanently(
	ctx context.Context,
	refundNo string,
	cause string,
) error {
	if err := h.store.MarkFailed(ctx, refundNo, cause); err != nil {
		return err
	}
	return outbox.Permanent(errors.New(cause))
}

func isPermanentRefundError(err error) bool {
	return errors.Is(err, apperr.ErrPaymentConflict) ||
		errors.Is(err, apperr.ErrPaymentProviderInvalid)
}
