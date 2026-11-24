package model

import (
	"errors"
	"fmt"
	"time"
)

const DefaultPaymentTTL = 30 * time.Minute

var (
	ErrInvalidPaymentInput      = errors.New("invalid payment request")
	ErrInvalidPaymentNumber     = errors.New("invalid payment number")
	ErrPaymentNotFound          = errors.New("payment order not found")
	ErrPaymentConflict          = errors.New("merchant order number was already used with different data")
	ErrPaymentExpired           = errors.New("payment order has expired")
	ErrPaymentClosed            = errors.New("payment order is closed")
	ErrPaymentNotClosable       = errors.New("payment order cannot be closed")
	ErrInvalidPaymentTransition = errors.New("invalid payment state transition")
	ErrNotificationNotFound     = errors.New("payment notification not found")
	ErrNotificationNotRetryable = errors.New("payment notification cannot be retried")
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusClosed  PaymentStatus = "closed"
)

type PaymentAction string

const (
	PaymentActionPay   PaymentAction = "pay"
	PaymentActionClose PaymentAction = "close"
)

var paymentTransitions = map[PaymentStatus]map[PaymentAction]PaymentStatus{
	PaymentStatusPending: {
		PaymentActionPay:   PaymentStatusPaid,
		PaymentActionClose: PaymentStatusClosed,
	},
	PaymentStatusPaid:   {},
	PaymentStatusClosed: {},
}

type Payment struct {
	ID              int64
	PaymentNo       string
	MerchantOrderNo string
	RequestHash     string
	Status          PaymentStatus
	AmountCents     int64
	Currency        string
	Subject         string
	NotifyURL       string
	ReturnURL       string
	ExpiresAt       time.Time
	PaidAt          *time.Time
	ClosedAt        *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (p Payment) CanTransition(action PaymentAction) bool {
	_, ok := paymentTransitions[p.Status][action]
	return ok
}

func (p *Payment) Apply(action PaymentAction, now time.Time) error {
	next, ok := paymentTransitions[p.Status][action]
	if !ok {
		return fmt.Errorf("%w: cannot %s a %s payment", ErrInvalidPaymentTransition, action, p.Status)
	}

	now = now.UTC()
	switch action {
	case PaymentActionPay:
		if !now.Before(p.ExpiresAt) {
			return ErrPaymentExpired
		}
		paidAt := now
		p.PaidAt = &paidAt
	case PaymentActionClose:
		closedAt := now
		p.ClosedAt = &closedAt
	}
	p.Status = next
	p.UpdatedAt = now
	return nil
}

func (p *Payment) CloseIfExpired(now time.Time) (bool, error) {
	if p.Status != PaymentStatusPending || now.UTC().Before(p.ExpiresAt) {
		return false, nil
	}
	if err := p.Apply(PaymentActionClose, now); err != nil {
		return false, err
	}
	return true, nil
}
