package dto

import "time"

type CreateOrderPaymentRequest struct {
	OrderID int `json:"order_id" binding:"required,min=1" example:"1"`
}

type OrderPaymentResponse struct {
	PaymentNo             string                 `json:"payment_no" example:"4f2614df-f674-4b5e-a9e4-3fc5e7f62013"`
	OrderID               int                    `json:"order_id" example:"1"`
	Provider              string                 `json:"provider" example:"mockpay"`
	ProviderTransactionID *string                `json:"provider_transaction_id"`
	Status                string                 `json:"status" example:"pending"`
	AmountCents           int64                  `json:"amount_cents" example:"3998"`
	CheckoutURL           string                 `json:"checkout_url,omitempty"`
	PaidAt                *time.Time             `json:"paid_at"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	Refund                *PaymentRefundResponse `json:"refund,omitempty"`
}

type PaymentRefundResponse struct {
	RefundNo         string     `json:"refund_no"`
	AmountCents      int64      `json:"amount_cents"`
	Reason           string     `json:"reason"`
	Status           string     `json:"status"`
	ProviderRefundNo *string    `json:"provider_refund_no"`
	RefundedAt       *time.Time `json:"refunded_at"`
	FailedAt         *time.Time `json:"failed_at"`
	LastError        *string    `json:"last_error"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type PaidCallbackRequest struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	OccurredAt time.Time        `json:"occurred_at"`
	Data       PaidCallbackData `json:"data"`
}

type PaidCallbackData struct {
	PaymentNo       string    `json:"payment_no"`
	MerchantOrderNo string    `json:"merchant_order_no"`
	Status          string    `json:"status"`
	AmountCents     int64     `json:"amount_cents"`
	Currency        string    `json:"currency"`
	Subject         string    `json:"subject"`
	PaidAt          time.Time `json:"paid_at"`
}
