package model

import (
	"errors"
	"testing"
	"time"

	"pinshaoshao/internal/feature/payment/apperr"
)

func TestPaymentMarkSucceededAcceptsLateCallbackAfterClose(t *testing.T) {
	existingProviderNo := "provider-1"
	paidAt := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	payment := &Payment{
		Status:                PaymentStatusClosed,
		ProviderTransactionID: &existingProviderNo,
	}

	if err := payment.MarkSucceeded(existingProviderNo, paidAt); err != nil {
		t.Fatalf("mark closed payment succeeded: %v", err)
	}
	if payment.Status != PaymentStatusSucceeded ||
		payment.PaidAt == nil ||
		!payment.PaidAt.Equal(paidAt) {
		t.Fatalf("payment = %+v, want succeeded at %s", payment, paidAt)
	}
}

func TestPaymentMarkSucceededRejectsDifferentProviderAfterClose(t *testing.T) {
	existingProviderNo := "provider-1"
	payment := &Payment{
		Status:                PaymentStatusClosed,
		ProviderTransactionID: &existingProviderNo,
	}

	err := payment.MarkSucceeded("provider-2", time.Now())
	if !errors.Is(err, apperr.ErrPaymentConflict) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrPaymentConflict)
	}
}
