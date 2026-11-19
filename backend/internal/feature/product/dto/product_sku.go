package dto

import "time"

type CreateProductSKURequest struct {
	Specs      map[string]string `json:"specs"`
	PriceCents int64             `json:"price_cents" binding:"min=0" example:"1999"`
	IsEnabled  *bool             `json:"is_enabled" example:"true"`
}

type UpdateProductSKURequest struct {
	Specs      map[string]string `json:"specs"`
	PriceCents *int64            `json:"price_cents" binding:"omitempty,min=0" example:"1999"`
	IsEnabled  *bool             `json:"is_enabled" example:"true"`
}

type ProductSKUResponse struct {
	ID         int               `json:"id" example:"1"`
	ProductID  int               `json:"product_id" example:"1"`
	Specs      map[string]string `json:"specs"`
	PriceCents int64             `json:"price_cents" example:"1999"`
	IsEnabled  bool              `json:"is_enabled" example:"true"`
	CreatedAt  time.Time         `json:"created_at" example:"2026-11-07T10:00:00Z"`
	UpdatedAt  time.Time         `json:"updated_at" example:"2026-11-07T10:00:00Z"`
}
