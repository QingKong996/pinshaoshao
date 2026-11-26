package payment

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/outbox"
)

type closeGatewayStub struct {
	err      error
	closedNo string
}

func (s *closeGatewayStub) ClosePayment(_ context.Context, providerPaymentNo string) error {
	s.closedNo = providerPaymentNo
	return s.err
}

func TestProviderCloseHandlerClosesProviderTransaction(t *testing.T) {
	gateway := &closeGatewayStub{}
	handler := &providerCloseHandler{gateway: gateway}
	payload, err := json.Marshal(map[string]string{
		"provider_transaction_id": "provider-1",
	})
	if err != nil {
		t.Fatalf("encode payload: %v", err)
	}

	err = handler.Handle(context.Background(), outbox.Event{Payload: payload})
	if err != nil {
		t.Fatalf("handle event: %v", err)
	}
	if gateway.closedNo != "provider-1" {
		t.Fatalf("closed provider transaction = %q, want provider-1", gateway.closedNo)
	}
}

func TestProviderCloseHandlerMarksInvalidPayloadPermanent(t *testing.T) {
	handler := &providerCloseHandler{gateway: &closeGatewayStub{}}

	tests := []struct {
		name    string
		payload []byte
	}{
		{name: "invalid json", payload: []byte("{")},
		{name: "missing provider transaction", payload: []byte(`{}`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.Background(), outbox.Event{Payload: tt.payload})
			if !outbox.IsPermanent(err) {
				t.Fatalf("error = %v, want permanent", err)
			}
		})
	}
}

func TestProviderCloseHandlerClassifiesGatewayErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		permanent bool
	}{
		{
			name:      "provider completed payment",
			err:       apperr.ErrPaymentConflict,
			permanent: true,
		},
		{
			name:      "provider rejected transaction",
			err:       apperr.ErrPaymentProviderInvalid,
			permanent: true,
		},
		{
			name: "provider unavailable",
			err:  apperr.ErrPaymentProviderUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &providerCloseHandler{gateway: &closeGatewayStub{err: tt.err}}
			payload, err := json.Marshal(map[string]string{
				"provider_transaction_id": "provider-1",
			})
			if err != nil {
				t.Fatalf("encode payload: %v", err)
			}

			got := handler.Handle(context.Background(), outbox.Event{Payload: payload})
			if !errors.Is(got, tt.err) {
				t.Fatalf("error = %v, want %v", got, tt.err)
			}
			if outbox.IsPermanent(got) != tt.permanent {
				t.Fatalf("permanent = %t, want %t", outbox.IsPermanent(got), tt.permanent)
			}
		})
	}
}
