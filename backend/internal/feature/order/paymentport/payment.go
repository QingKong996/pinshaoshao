package paymentport

import (
	"context"
	"time"

	"pinshaoshao/internal/database"
)

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusPaid           OrderStatus = "paid"
	OrderStatusCancelled      OrderStatus = "cancelled"
)

type Order struct {
	ID               int
	UserID           int
	Status           OrderStatus
	TotalAmountCents int64
	ExpiresAt        time.Time
	Expired          bool
}

type CloseResult struct {
	PaymentID             int64
	PaymentNo             string
	ProviderTransactionID *string
	Closed                bool
	AlreadySettled        bool
}

// Lifecycle exposes the order state changes needed by the payment feature.
type Lifecycle interface {
	LockForPayment(ctx context.Context, db database.DBTX, userID int, orderID int) (*Order, error)
	SettlePayment(ctx context.Context, db database.DBTX, orderID int, paidAt time.Time) (OrderStatus, error)
}
