package model

import (
	"time"

	"pinshaoshao/internal/feature/payment/apperr"
)

const (
	ProviderMockPay = "mockpay"
	CurrencyCNY     = "CNY"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusClosed    PaymentStatus = "closed"
)

type Payment struct {
	ID                    int64
	PaymentNo             string
	OrderID               int
	Provider              string
	ProviderTransactionID *string
	Status                PaymentStatus
	AmountCents           int64
	OrderExpiresAt        time.Time
	PaidAt                *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (p *Payment) MarkSucceeded(providerTransactionID string, paidAt time.Time) error {
	switch p.Status {
	case PaymentStatusSucceeded:
		if p.ProviderTransactionID != nil && *p.ProviderTransactionID == providerTransactionID {
			return nil
		}
		return apperr.ErrPaymentConflict
	case PaymentStatusClosed:
		return apperr.ErrPaymentClosed
	case PaymentStatusPending:
		if p.ProviderTransactionID != nil && *p.ProviderTransactionID != providerTransactionID {
			return apperr.ErrPaymentConflict
		}
		paidAt = paidAt.UTC()
		p.Status = PaymentStatusSucceeded
		p.ProviderTransactionID = &providerTransactionID
		p.PaidAt = &paidAt
		p.UpdatedAt = paidAt
		return nil
	default:
		return apperr.ErrPaymentConflict
	}
}

func (p *Payment) MarkClosed(closedAt time.Time) error {
	switch p.Status {
	case PaymentStatusClosed:
		return nil
	case PaymentStatusSucceeded:
		return apperr.ErrPaymentConflict
	case PaymentStatusPending:
		p.Status = PaymentStatusClosed
		p.UpdatedAt = closedAt.UTC()
		return nil
	default:
		return apperr.ErrPaymentConflict
	}
}

type PaidCallback struct {
	EventID               string
	EventType             string
	ProviderTransactionID string
	PaymentNo             string
	AmountCents           int64
	Currency              string
	PaidAt                time.Time
}

type ProviderPaymentRequest struct {
	MerchantOrderNo string
	AmountCents     int64
	Currency        string
	Subject         string
	NotifyURL       string
	ReturnURL       string
	ExpiresIn       time.Duration
}

type ProviderPaymentResult struct {
	PaymentNo   string
	MerchantNo  string
	Status      string
	AmountCents int64
	Currency    string
	CheckoutURL string
	ExpiresAt   time.Time
}
