package paymentport

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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
}

// Lifecycle exposes the order state changes needed by the payment feature.
type Lifecycle interface {
	LockForPayment(ctx context.Context, tx pgx.Tx, userID int, orderID int) (*Order, error)
	SettlePayment(ctx context.Context, tx pgx.Tx, orderID int, paidAt time.Time) (OrderStatus, error)
	CancelForPayment(ctx context.Context, tx pgx.Tx, userID int, orderID int) (OrderStatus, error)
}
