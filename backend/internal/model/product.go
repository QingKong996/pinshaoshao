package model

import "time"

type ProductStatus string

const (
	ProductStatusOnShelf  ProductStatus = "on_shelf"
	ProductStatusOffShelf ProductStatus = "off_shelf"
)

type ProductSort string

const (
	ProductSortCreatedAt  ProductSort = "created_at"
	ProductSortUpdatedAt  ProductSort = "updated_at"
	ProductSortName       ProductSort = "name"
	ProductSortCategory   ProductSort = "category"
	ProductSortStatus     ProductSort = "status"
	ProductSortSKUCount   ProductSort = "sku_count"
	ProductSortTotalStock ProductSort = "total_stock"
	ProductSortMinPrice   ProductSort = "min_price"
)

type ProductSortOrder string

const (
	ProductSortOrderAsc  ProductSortOrder = "asc"
	ProductSortOrderDesc ProductSortOrder = "desc"
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

type ProductListFilter struct {
	Search        string
	Status        ProductStatus
	CategoryID    *int
	Uncategorized bool
	Sort          ProductSort
	Order         ProductSortOrder
	Page          int
	PageSize      int
}

type ProductListResult struct {
	Products []ProductSummary
	Page     int
	PageSize int
	Total    int64
}
