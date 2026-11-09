package model

import "time"

type ProductCategory struct {
	ID           int
	ShopID       int
	Name         string
	ProductCount int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
