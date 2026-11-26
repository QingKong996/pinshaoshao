package model

import "time"

const RefundReasonOrderCancelled = "order_cancelled_after_payment"

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
)

type Refund struct {
	ID                    int64
	RefundNo              string
	PaymentID             int64
	PaymentNo             string
	ProviderTransactionID *string
	AmountCents           int64
	Reason                string
	Status                RefundStatus
	ProviderRefundNo      *string
	RefundedAt            *time.Time
	FailedAt              *time.Time
	LastError             *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ProviderRefundRequest struct {
	RefundNo     string
	AmountCents  int64
	Reason       string
	Currency     string
	MerchantNo   string
	ProviderTxID string
}

type ProviderRefundResult struct {
	RefundNo         string
	ProviderRefundNo string
	Status           string
	AmountCents      int64
	RefundedAt       time.Time
}
