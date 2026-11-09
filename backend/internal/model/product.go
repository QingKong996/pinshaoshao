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
	CategoryID  *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductSummary is a product joined with its category and SKU aggregates.
type ProductSummary struct {
	Product
	CategoryName  *string
	SKUCount      int64
	TotalStock    int64
	MinPriceCents *int64
	MaxPriceCents *int64
}
