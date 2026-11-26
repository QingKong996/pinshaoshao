package shippingport

import (
	"context"
	"errors"
	"time"

	"pinshaoshao/internal/database"
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
	LockOrder(ctx context.Context, db database.DBTX, orderID int) error
	ShipSubOrder(
		ctx context.Context,
		db database.DBTX,
		shopID int,
		subOrderID int,
		shippedAt time.Time,
	) (*Target, bool, error)
	CompleteSubOrder(ctx context.Context, db database.DBTX, subOrderID int, deliveredAt time.Time) error
}
