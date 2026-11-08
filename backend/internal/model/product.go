package model

import "time"

type ProductStatus string

const (
	ProductStatusOnShelf  ProductStatus = "on_shelf"
	ProductStatusOffShelf ProductStatus = "off_shelf"
)

type Product struct {
	ID          int
	ShopID      int
	Name        string
	Description string
	Status      ProductStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
