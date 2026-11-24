package inventory

import (
	"context"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/inventory/model"
)

// Lifecycle owns inventory rows that are created and removed with catalog records.
type Lifecycle interface {
	CreateEmptyForSKU(ctx context.Context, db database.DBTX, skuID int) error
	DeleteEmptyForSKU(ctx context.Context, db database.DBTX, productID int, skuID int) error
	DeleteEmptyForProduct(ctx context.Context, db database.DBTX, productID int) error
}

// Reservation owns stock freezing for order placement and cancellation.
type Reservation interface {
	Freeze(ctx context.Context, db database.DBTX, items []model.ReservationItem) error
	Release(ctx context.Context, db database.DBTX, items []model.ReservationItem) error
	Commit(ctx context.Context, db database.DBTX, items []model.ReservationItem) error
}
