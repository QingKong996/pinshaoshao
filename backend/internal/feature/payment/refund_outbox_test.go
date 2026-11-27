package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/outbox"
)

type refundStoreStub struct {
	refund           *model.Refund
	getErr           error
	succeededNo      string
	providerNo       string
	refundedAt       time.Time
	failedCause      string
	markSucceededErr error
}

func (s *refundStoreStub) GetForProcessing(
	context.Context,
	string,
) (*model.Refund, error) {
	return s.refund, s.getErr
}

func (s *refundStoreStub) MarkSucceeded(
	_ context.Context,
	refundNo string,
	providerRefundNo string,
	refundedAt time.Time,
) error {
	s.succeededNo = refundNo
	s.providerNo = providerRefundNo
	s.refundedAt = refundedAt
	return s.markSucceededErr
}

func (s *refundStoreStub) MarkFailed(_ context.Context, _ string, cause string) error {
	s.failedCause = cause
	return nil
}

type refundGatewayStub struct {
	result *model.ProviderRefundResult
	err    error
}

func (s *refundGatewayStub) RefundPayment(
	context.Context,
	string,
	model.ProviderRefundRequest,
) (*model.ProviderRefundResult, error) {
	return s.result, s.err
}

func TestProviderRefundHandlerRefundsAndMarksSucceeded(t *testing.T) {
	refundedAt := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	store := &refundStoreStub{
		refund: &model.Refund{
			RefundNo:              "refund-1",
			PaymentNo:             "payment-1",
			ProviderTransactionID: stringPointerForTest("provider-1"),
			AmountCents:           3998,
			Reason:                model.RefundReasonOrderCancelled,
			Status:                model.RefundStatusPending,
		},
	}
	gateway := &refundGatewayStub{
		result: &model.ProviderRefundResult{
			RefundNo:         "refund-1",
			ProviderRefundNo: "provider-refund-1",
			Status:           "succeeded",
			AmountCents:      3998,
			RefundedAt:       refundedAt,
		},
	}
	handler := &providerRefundHandler{store: store, gateway: gateway}

	if err := handler.Handle(context.Background(), refundEvent(t, "refund-1")); err != nil {
		t.Fatalf("handle refund: %v", err)
	}
	if store.succeededNo != "refund-1" ||
		store.providerNo != "provider-refund-1" ||
		!store.refundedAt.Equal(refundedAt) {
		t.Fatalf(
			"successful refund = %q/%q/%s",
			store.succeededNo,
			store.providerNo,
			store.refundedAt,
		)
	}
}

func TestProviderRefundHandlerRetriesUnavailableProvider(t *testing.T) {
	store := &refundStoreStub{
		refund: &model.Refund{
			RefundNo:              "refund-1",
			ProviderTransactionID: stringPointerForTest("provider-1"),
			AmountCents:           3998,
			Status:                model.RefundStatusPending,
		},
	}
	handler := &providerRefundHandler{
		store:   store,
		gateway: &refundGatewayStub{err: apperr.ErrPaymentProviderUnavailable},
	}

	err := handler.Handle(context.Background(), refundEvent(t, "refund-1"))
	if !errors.Is(err, apperr.ErrPaymentProviderUnavailable) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentProviderUnavailable)
	}
	if outbox.IsPermanent(err) || store.failedCause != "" {
		t.Fatalf("unavailable provider should be retried")
	}
}

func TestProviderRefundHandlerMarksPermanentFailure(t *testing.T) {
	store := &refundStoreStub{
		refund: &model.Refund{
			RefundNo:              "refund-1",
			ProviderTransactionID: stringPointerForTest("provider-1"),
			AmountCents:           3998,
			Status:                model.RefundStatusPending,
		},
	}
	handler := &providerRefundHandler{
		store: store,
		gateway: &refundGatewayStub{err: fmt.Errorf(
			"%w: provider response contained internal data",
			apperr.ErrPaymentProviderInvalid,
		)},
	}

	err := handler.Handle(context.Background(), refundEvent(t, "refund-1"))
	if !outbox.IsPermanent(err) {
		t.Fatalf("error = %v, want permanent", err)
	}
	if store.failedCause != "payment provider rejected the refund" {
		t.Fatalf("failed cause = %q, want a sanitized summary", store.failedCause)
	}
}

func TestProviderRefundHandlerSkipsTerminalRefunds(t *testing.T) {
	for _, status := range []model.RefundStatus{
		model.RefundStatusSucceeded,
		model.RefundStatusFailed,
	} {
		t.Run(string(status), func(t *testing.T) {
			store := &refundStoreStub{
				refund: &model.Refund{
					RefundNo:              "refund-1",
					ProviderTransactionID: stringPointerForTest("provider-1"),
					Status:                status,
				},
			}
			handler := &providerRefundHandler{
				store:   store,
				gateway: &refundGatewayStub{err: errors.New("gateway must not be called")},
			}

			if err := handler.Handle(context.Background(), refundEvent(t, "refund-1")); err != nil {
				t.Fatalf("handle refund: %v", err)
			}
			if store.failedCause != "" {
				t.Fatalf("terminal refund changed to failed: %q", store.failedCause)
			}
		})
	}
}

func TestProviderRefundHandlerFailsMissingProviderTransaction(t *testing.T) {
	store := &refundStoreStub{
		refund: &model.Refund{
			RefundNo:    "refund-1",
			AmountCents: 3998,
			Status:      model.RefundStatusPending,
		},
	}
	handler := &providerRefundHandler{
		store:   store,
		gateway: &refundGatewayStub{},
	}

	err := handler.Handle(context.Background(), refundEvent(t, "refund-1"))
	if !outbox.IsPermanent(err) {
		t.Fatalf("error = %v, want permanent", err)
	}
	if store.failedCause == "" {
		t.Fatal("missing provider transaction was not recorded")
	}
}

func TestProviderRefundHandlerRejectsInvalidResult(t *testing.T) {
	store := &refundStoreStub{
		refund: &model.Refund{
			RefundNo:              "refund-1",
			ProviderTransactionID: stringPointerForTest("provider-1"),
			AmountCents:           3998,
			Reason:                model.RefundReasonOrderCancelled,
			Status:                model.RefundStatusPending,
		},
	}
	handler := &providerRefundHandler{
		store: store,
		gateway: &refundGatewayStub{result: &model.ProviderRefundResult{
			RefundNo:         "refund-1",
			ProviderRefundNo: "provider-refund-1",
			Status:           "succeeded",
			AmountCents:      1,
			RefundedAt:       time.Now().UTC(),
		}},
	}

	err := handler.Handle(context.Background(), refundEvent(t, "refund-1"))
	if !outbox.IsPermanent(err) {
		t.Fatalf("error = %v, want permanent", err)
	}
	if store.failedCause == "" {
		t.Fatal("invalid provider result was not recorded")
	}
}

func TestProviderRefundHandlerRejectsInvalidPayload(t *testing.T) {
	handler := &providerRefundHandler{
		store:   &refundStoreStub{},
		gateway: &refundGatewayStub{},
	}

	for _, payload := range [][]byte{[]byte("{"), []byte(`{}`)} {
		err := handler.Handle(context.Background(), outbox.Event{Payload: payload})
		if !outbox.IsPermanent(err) {
			t.Fatalf("payload %s: error = %v, want permanent", payload, err)
		}
	}
}

func refundEvent(t *testing.T, refundNo string) outbox.Event {
	t.Helper()

	payload, err := json.Marshal(map[string]string{"refund_no": refundNo})
	if err != nil {
		t.Fatalf("encode refund event: %v", err)
	}
	return outbox.Event{Payload: payload}
}

func stringPointerForTest(value string) *string {
	return &value
}
