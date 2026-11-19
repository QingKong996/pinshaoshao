package model

import "time"

// InventoryItem is an independent stock record enriched with its SKU, product, and category.
type InventoryItem struct {
	SKUID        int
	ProductID    int
	ProductName  string
	CategoryID   *int
	CategoryName *string
	Specs        map[string]string
	PriceCents   int64
	TotalStock   int
	FrozenStock  int
	Version      int64
	IsEnabled    bool
	UpdatedAt    time.Time
}

func (i InventoryItem) AvailableStock() int {
	return i.TotalStock - i.FrozenStock
}
