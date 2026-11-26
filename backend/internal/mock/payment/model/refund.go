package model

import (
	"errors"
	"time"
)

var (
	ErrInvalidRefundInput   = errors.New("invalid refund request")
	ErrRefundNotFound       = errors.New("refund not found")
	ErrRefundConflict       = errors.New("refund conflicts with the existing refund")
	ErrPaymentNotRefundable = errors.New("payment is not refundable")
)

type RefundStatus string

const RefundStatusSucceeded RefundStatus = "succeeded"

type Refund struct {
	ID               int64
	RefundNo         string
	ProviderRefundNo string
	PaymentID        int64
	PaymentNo        string
	AmountCents      int64
	Reason           string
	Status           RefundStatus
	RefundedAt       time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type RefundInput struct {
	RefundNo    string
	AmountCents int64
	Reason      string
}
