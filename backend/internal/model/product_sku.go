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
