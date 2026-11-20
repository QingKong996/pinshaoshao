package inventory

import (
	"context"

	"pinshaoshao/internal/database"
)

// Lifecycle owns inventory rows that are created and removed with catalog records.
type Lifecycle interface {
	CreateEmptyForSKU(ctx context.Context, db database.DBTX, skuID int) error
	DeleteEmptyForSKU(ctx context.Context, db database.DBTX, productID int, skuID int) error
	DeleteEmptyForProduct(ctx context.Context, db database.DBTX, productID int) error
}
