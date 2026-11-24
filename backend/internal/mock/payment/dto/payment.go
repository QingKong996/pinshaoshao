package dto

import "time"

type CreatePaymentRequest struct {
	MerchantOrderNo  string `json:"merchant_order_no" binding:"required,max=64" example:"order-20261124-001"`
	AmountCents      int64  `json:"amount_cents" binding:"required,min=1" example:"3998"`
	Currency         string `json:"currency" binding:"omitempty,len=3" example:"CNY"`
	Subject          string `json:"subject" binding:"required,max=200" example:"T-shirt order"`
	NotifyURL        string `json:"notify_url" binding:"required,url,max=2048" example:"http://localhost:8080/api/v1/payment/webhook"`
	ReturnURL        string `json:"return_url" binding:"omitempty,url,max=2048" example:"http://localhost:5173/orders/1"`
	ExpiresInSeconds int    `json:"expires_in_seconds" binding:"omitempty,min=60,max=86400" example:"1800"`
}

type PaymentResponse struct {
	PaymentNo       string     `json:"payment_no" example:"4f2614df-f674-4b5e-a9e4-3fc5e7f62013"`
	MerchantOrderNo string     `json:"merchant_order_no" example:"order-20261124-001"`
	Status          string     `json:"status" example:"pending"`
	AmountCents     int64      `json:"amount_cents" example:"3998"`
	Currency        string     `json:"currency" example:"CNY"`
	Subject         string     `json:"subject" example:"T-shirt order"`
	CheckoutURL     string     `json:"checkout_url" example:"http://localhost:8080/mockpay/checkout/4f2614df-f674-4b5e-a9e4-3fc5e7f62013"`
	ReturnURL       string     `json:"return_url,omitempty" example:"http://localhost:5173/orders/1"`
	ExpiresAt       time.Time  `json:"expires_at"`
	PaidAt          *time.Time `json:"paid_at"`
	ClosedAt        *time.Time `json:"closed_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

type NotificationResponse struct {
	ID            int64      `json:"id"`
	EventID       string     `json:"event_id"`
	EventType     string     `json:"event_type"`
	TargetURL     string     `json:"target_url"`
	Status        string     `json:"status"`
	AttemptCount  int        `json:"attempt_count"`
	MaxAttempts   int        `json:"max_attempts"`
	NextAttemptAt time.Time  `json:"next_attempt_at"`
	LastAttemptAt *time.Time `json:"last_attempt_at"`
	DeliveredAt   *time.Time `json:"delivered_at"`
	LastError     *string    `json:"last_error"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type NotificationListResponse struct {
	Items []NotificationResponse `json:"items"`
}
