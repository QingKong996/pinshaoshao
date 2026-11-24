package model

import (
	"encoding/json"
	"time"
)

const (
	DefaultNotificationMaxAttempts = 7
	EventTypePaymentPaid           = "payment.paid"
)

type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "pending"
	NotificationStatusDelivering NotificationStatus = "delivering"
	NotificationStatusSucceeded  NotificationStatus = "succeeded"
	NotificationStatusExhausted  NotificationStatus = "exhausted"
)

type Notification struct {
	ID            int64
	EventID       string
	PaymentID     int64
	EventType     string
	TargetURL     string
	Payload       []byte
	Status        NotificationStatus
	AttemptCount  int
	MaxAttempts   int
	NextAttemptAt time.Time
	LockedUntil   *time.Time
	LastAttemptAt *time.Time
	DeliveredAt   *time.Time
	LastError     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentEvent struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	OccurredAt time.Time        `json:"occurred_at"`
	Data       PaymentEventData `json:"data"`
}

type PaymentEventData struct {
	PaymentNo       string        `json:"payment_no"`
	MerchantOrderNo string        `json:"merchant_order_no"`
	Status          PaymentStatus `json:"status"`
	AmountCents     int64         `json:"amount_cents"`
	Currency        string        `json:"currency"`
	Subject         string        `json:"subject"`
	PaidAt          time.Time     `json:"paid_at"`
}

func PaymentPaidPayload(eventID string, payment *Payment, now time.Time) ([]byte, error) {
	paidAt := now.UTC()
	if payment.PaidAt != nil {
		paidAt = payment.PaidAt.UTC()
	}
	return json.Marshal(PaymentEvent{
		ID:         eventID,
		Type:       EventTypePaymentPaid,
		OccurredAt: paidAt,
		Data: PaymentEventData{
			PaymentNo:       payment.PaymentNo,
			MerchantOrderNo: payment.MerchantOrderNo,
			Status:          payment.Status,
			AmountCents:     payment.AmountCents,
			Currency:        payment.Currency,
			Subject:         payment.Subject,
			PaidAt:          paidAt,
		},
	})
}

func NewPaidNotification(
	eventID string,
	payment *Payment,
	payload []byte,
	maxAttempts int,
	now time.Time,
) *Notification {
	now = now.UTC()
	return &Notification{
		EventID:       eventID,
		PaymentID:     payment.ID,
		EventType:     EventTypePaymentPaid,
		TargetURL:     payment.NotifyURL,
		Payload:       payload,
		Status:        NotificationStatusPending,
		MaxAttempts:   maxAttempts,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
