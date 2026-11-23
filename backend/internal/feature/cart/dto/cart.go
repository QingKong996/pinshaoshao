package dto

import "time"

type AddCartItemRequest struct {
	SKUID    int `json:"sku_id" binding:"required,min=1" example:"5"`
	Quantity int `json:"quantity" binding:"required,min=1,max=2147483647" example:"2"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1,max=2147483647" example:"3"`
}

type CartItemResponse struct {
	ID             int               `json:"id" example:"1"`
	ProductID      int               `json:"product_id" example:"10"`
	SKUID          int               `json:"sku_id" example:"5"`
	ProductName    string            `json:"product_name" example:"纯棉T恤"`
	SKUSpecs       map[string]string `json:"sku_specs"`
	UnitPriceCents int64             `json:"unit_price_cents" example:"1999"`
	Quantity       int               `json:"quantity" example:"2"`
	LineTotalCents int64             `json:"line_total_cents" example:"3998"`
	Availability   string            `json:"availability" example:"available"`
	CreatedAt      time.Time         `json:"created_at" example:"2026-11-19T12:00:00Z"`
	UpdatedAt      time.Time         `json:"updated_at" example:"2026-11-19T12:00:00Z"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid cart item"`
}
