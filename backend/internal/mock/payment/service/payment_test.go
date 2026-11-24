package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"pinshaoshao/internal/mock/payment/model"
)

func TestNormalizeCreateInput(t *testing.T) {
	input, hash, err := normalizeCreateInput(CreatePaymentInput{
		MerchantOrderNo: " order-1 ",
		AmountCents:     3998,
		Subject:         " T-shirt ",
		NotifyURL:       "http://localhost:8080/webhook",
	})
	if err != nil {
		t.Fatalf("normalize input: %v", err)
	}
	if input.Currency != "CNY" || input.ExpiresIn != defaultPaymentTTL {
		t.Fatalf("normalized input = %+v, want CNY and default TTL", input)
	}
	if input.MerchantOrderNo != "order-1" || input.Subject != "T-shirt" {
		t.Fatalf("trimmed input = %+v", input)
	}
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hash))
	}
}

func TestNormalizeCreateInputRejectsInvalidCallbacks(t *testing.T) {
	for _, notifyURL := range []string{"", "ftp://example.com/hook", "http://user:pass@example.com/hook"} {
		_, _, err := normalizeCreateInput(CreatePaymentInput{
			MerchantOrderNo: "order-1",
			AmountCents:     100,
			Subject:         "Order",
			NotifyURL:       notifyURL,
		})
		if !errors.Is(err, model.ErrInvalidPaymentInput) {
			t.Fatalf("notify URL %q error = %v, want invalid input", notifyURL, err)
		}
	}
}

func TestNormalizeCreateInputRejectsInvalidFields(t *testing.T) {
	valid := CreatePaymentInput{
		MerchantOrderNo: "order-1",
		AmountCents:     100,
		Currency:        "CNY",
		Subject:         "Order",
		NotifyURL:       "http://localhost:8080/webhook",
		ExpiresIn:       time.Minute,
	}

	testCases := []struct {
		name   string
		mutate func(*CreatePaymentInput)
	}{
		{
			name: "empty merchant order number",
			mutate: func(input *CreatePaymentInput) {
				input.MerchantOrderNo = " "
			},
		},
		{
			name: "merchant order number too long",
			mutate: func(input *CreatePaymentInput) {
				input.MerchantOrderNo = strings.Repeat("a", maxMerchantNoLen+1)
			},
		},
		{
			name: "non-positive amount",
			mutate: func(input *CreatePaymentInput) {
				input.AmountCents = 0
			},
		},
		{
			name: "invalid currency",
			mutate: func(input *CreatePaymentInput) {
				input.Currency = "CN"
			},
		},
		{
			name: "empty subject",
			mutate: func(input *CreatePaymentInput) {
				input.Subject = " "
			},
		},
		{
			name: "subject too long",
			mutate: func(input *CreatePaymentInput) {
				input.Subject = strings.Repeat("a", maxSubjectLen+1)
			},
		},
		{
			name: "expiry below minimum",
			mutate: func(input *CreatePaymentInput) {
				input.ExpiresIn = minPaymentTTL - time.Second
			},
		},
		{
			name: "expiry above maximum",
			mutate: func(input *CreatePaymentInput) {
				input.ExpiresIn = maxPaymentTTL + time.Second
			},
		},
		{
			name: "invalid return URL",
			mutate: func(input *CreatePaymentInput) {
				input.ReturnURL = "ftp://example.com/return"
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			input := valid
			testCase.mutate(&input)
			if _, _, err := normalizeCreateInput(input); !errors.Is(err, model.ErrInvalidPaymentInput) {
				t.Fatalf("error = %v, want invalid input", err)
			}
		})
	}
}

func TestNormalizeCreateInputAcceptsBoundaries(t *testing.T) {
	input, _, err := normalizeCreateInput(CreatePaymentInput{
		MerchantOrderNo: "a",
		AmountCents:     1,
		Currency:        "usd",
		Subject:         "S",
		NotifyURL:       "https://example.com/webhook",
		ReturnURL:       "https://example.com/return?order=1",
		ExpiresIn:       maxPaymentTTL,
	})
	if err != nil {
		t.Fatalf("normalize boundary input: %v", err)
	}
	if input.Currency != "USD" || input.ExpiresIn != maxPaymentTTL {
		t.Fatalf("normalized input = %+v, want USD and maximum TTL", input)
	}
}

type expiredPaymentRepository struct {
	payment     *model.Payment
	closedCalls int
}

func (r *expiredPaymentRepository) Create(context.Context, *model.Payment) (*model.Payment, bool, error) {
	panic("unexpected Create call")
}

func (r *expiredPaymentRepository) Get(context.Context, string) (*model.Payment, error) {
	return r.payment, nil
}

func (r *expiredPaymentRepository) Pay(context.Context, string, time.Time) (*model.Payment, error) {
	panic("unexpected Pay call")
}

func (r *expiredPaymentRepository) Close(_ context.Context, _ string, now time.Time) (*model.Payment, error) {
	r.closedCalls++
	if err := r.payment.Apply(model.PaymentActionClose, now); err != nil {
		return nil, err
	}
	return r.payment, nil
}

func (r *expiredPaymentRepository) CloseExpired(context.Context, time.Time, int) ([]model.Payment, error) {
	panic("unexpected CloseExpired call")
}

func TestGetClosesExpiredPayment(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	repository := &expiredPaymentRepository{
		payment: &model.Payment{
			PaymentNo: "4f2614df-f674-4b5e-a9e4-3fc5e7f62013",
			Status:    model.PaymentStatusPending,
			ExpiresAt: now.Add(-time.Second),
		},
	}
	paymentService := NewPaymentService(repository)
	paymentService.now = func() time.Time { return now }

	payment, err := paymentService.Get(context.Background(), repository.payment.PaymentNo)
	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if payment.Status != model.PaymentStatusClosed || repository.closedCalls != 1 {
		t.Fatalf("status/close calls = %q/%d, want closed/1", payment.Status, repository.closedCalls)
	}
}
