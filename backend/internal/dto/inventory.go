package dto

import "time"

type InventoryItemResponse struct {
	SKUID        int               `json:"sku_id" example:"5"`
	ProductID    int               `json:"product_id" example:"1"`
	ProductName  string            `json:"product_name" example:"纯棉T恤"`
	CategoryID   *int              `json:"category_id" example:"2"`
	CategoryName *string           `json:"category_name" example:"服饰"`
	Specs        map[string]string `json:"specs"`
	PriceCents   int64             `json:"price_cents" example:"1999"`
	Stock        int               `json:"stock" example:"50"`
	IsEnabled    bool              `json:"is_enabled" example:"true"`
	UpdatedAt    time.Time         `json:"updated_at" example:"2026-11-09T12:00:00Z"`
}
