package dto

import "time"

type InventoryItemResponse struct {
	SKUID          int               `json:"sku_id" example:"5"`
	ProductID      int               `json:"product_id" example:"1"`
	ProductName    string            `json:"product_name" example:"纯棉T恤"`
	CategoryID     *int              `json:"category_id" example:"2"`
	CategoryName   *string           `json:"category_name" example:"服饰"`
	Specs          map[string]string `json:"specs"`
	PriceCents     int64             `json:"price_cents" example:"1999"`
	TotalStock     int               `json:"total_stock" example:"50"`
	FrozenStock    int               `json:"frozen_stock" example:"5"`
	AvailableStock int               `json:"available_stock" example:"45"`
	Version        int64             `json:"version" example:"3"`
	IsEnabled      bool              `json:"is_enabled" example:"true"`
	UpdatedAt      time.Time         `json:"updated_at" example:"2026-11-09T12:00:00Z"`
}

type UpdateInventoryRequest struct {
	TotalStock *int   `json:"total_stock" binding:"required,min=0,max=2147483647" example:"50"`
	Version    *int64 `json:"version" binding:"required,min=0" example:"3"`
}
