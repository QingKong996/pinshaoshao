package shippingport

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrSubOrderNotFound     = errors.New("sub-order not found")
	ErrSubOrderNotShippable = errors.New("sub-order is not shippable")
)

type Target struct {
	SubOrderID      int
	OrderID         int
	UserID          int
	ShopID          int
	ShopName        string
	Status          string
	ItemSummary     string
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

// Lifecycle exposes the order transitions needed by the logistics feature.
type Lifecycle interface {
	GetShipmentTarget(ctx context.Context, shopID int, subOrderID int) (*Target, error)
	LockOrder(ctx context.Context, tx pgx.Tx, orderID int) error
	ShipSubOrder(
		ctx context.Context,
		tx pgx.Tx,
		shopID int,
		subOrderID int,
		shippedAt time.Time,
	) (*Target, bool, error)
	CompleteSubOrder(ctx context.Context, tx pgx.Tx, subOrderID int, deliveredAt time.Time) error
}
