package model_test

import (
	"errors"
	"testing"
	"time"

	"pinshaoshao/internal/mock/payment/model"
)

func TestPaymentStateTransitions(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(time.Hour)

	t.Run("pending payment can be paid", func(t *testing.T) {
		payment := model.Payment{Status: model.PaymentStatusPending, ExpiresAt: expiresAt}

		if err := payment.Apply(model.PaymentActionPay, now); err != nil {
			t.Fatalf("apply pay: %v", err)
		}
		if payment.Status != model.PaymentStatusPaid || payment.PaidAt == nil {
			t.Fatalf("payment = %+v, want paid with paid_at", payment)
		}
		if payment.CanTransition(model.PaymentActionPay) || payment.CanTransition(model.PaymentActionClose) {
			t.Fatal("paid payment should be terminal")
		}
	})

	t.Run("pending payment can be closed", func(t *testing.T) {
		payment := model.Payment{Status: model.PaymentStatusPending, ExpiresAt: expiresAt}

		if err := payment.Apply(model.PaymentActionClose, now); err != nil {
			t.Fatalf("apply close: %v", err)
		}
		if payment.Status != model.PaymentStatusClosed || payment.ClosedAt == nil {
			t.Fatalf("payment = %+v, want closed with closed_at", payment)
		}
		if err := payment.Apply(model.PaymentActionPay, now); !errors.Is(err, model.ErrInvalidPaymentTransition) {
			t.Fatalf("pay closed error = %v, want invalid transition", err)
		}
	})

	t.Run("expired payment cannot be paid", func(t *testing.T) {
		payment := model.Payment{Status: model.PaymentStatusPending, ExpiresAt: now}

		if err := payment.Apply(model.PaymentActionPay, now); !errors.Is(err, model.ErrPaymentExpired) {
			t.Fatalf("apply pay error = %v, want payment expired", err)
		}
		if payment.Status != model.PaymentStatusPending {
			t.Fatalf("status = %q, want pending", payment.Status)
		}
	})
}

func TestPaymentCloseIfExpired(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	payment := model.Payment{Status: model.PaymentStatusPending, ExpiresAt: now}

	closed, err := payment.CloseIfExpired(now)
	if err != nil {
		t.Fatalf("close expired: %v", err)
	}
	if !closed || payment.Status != model.PaymentStatusClosed {
		t.Fatalf("closed/status = %v/%q, want true/closed", closed, payment.Status)
	}

	notExpired := model.Payment{
		Status:    model.PaymentStatusPending,
		ExpiresAt: now.Add(time.Second),
	}
	closed, err = notExpired.CloseIfExpired(now)
	if err != nil || closed || notExpired.Status != model.PaymentStatusPending {
		t.Fatalf("not expired closed/status/error = %v/%q/%v", closed, notExpired.Status, err)
	}

	paid := model.Payment{
		Status:    model.PaymentStatusPaid,
		ExpiresAt: now.Add(-time.Second),
	}
	closed, err = paid.CloseIfExpired(now)
	if err != nil || closed || paid.Status != model.PaymentStatusPaid {
		t.Fatalf("paid closed/status/error = %v/%q/%v", closed, paid.Status, err)
	}
}
