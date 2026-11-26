package dto

import "time"

type RefundRequest struct {
	RefundNo    string `json:"refund_no" binding:"required" example:"4f2614df-f674-4b5e-a9e4-3fc5e7f62013"`
	AmountCents int64  `json:"amount_cents" binding:"required,min=1" example:"3998"`
	Reason      string `json:"reason" binding:"required,max=200" example:"order_cancelled_after_payment"`
}

type RefundResponse struct {
	RefundNo         string    `json:"refund_no" example:"4f2614df-f674-4b5e-a9e4-3fc5e7f62013"`
	ProviderRefundNo string    `json:"provider_refund_no" example:"8ad23f19-24f5-48b8-bbc7-243e1f1e0a98"`
	PaymentNo        string    `json:"payment_no" example:"bdce2f67-3b2e-44c5-b3c4-2f3edc913775"`
	AmountCents      int64     `json:"amount_cents" example:"3998"`
	Reason           string    `json:"reason" example:"order_cancelled_after_payment"`
	Status           string    `json:"status" example:"succeeded"`
	RefundedAt       time.Time `json:"refunded_at"`
}
