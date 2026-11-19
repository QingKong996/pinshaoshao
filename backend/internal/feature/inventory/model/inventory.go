package model

import "time"

// InventoryItem is a SKU enriched with its product and category, as shown on the console.
type InventoryItem struct {
	SKUID        int
	ProductID    int
	ProductName  string
	CategoryID   *int
	CategoryName *string
	Specs        map[string]string
	PriceCents   int64
	Stock        int
	IsEnabled    bool
	UpdatedAt    time.Time
}
