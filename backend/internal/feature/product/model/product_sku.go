package model

import "time"

type ProductSKU struct {
	ID         int
	ProductID  int
	Specs      map[string]string
	PriceCents int64
	Stock      int
	IsEnabled  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ProductSKUUpdate carries the optional fields of a SKU partial update.
// A nil Specs keeps the current specs; a non-nil map replaces them.
type ProductSKUUpdate struct {
	Specs      map[string]string
	PriceCents *int64
	Stock      *int
	IsEnabled  *bool
}
