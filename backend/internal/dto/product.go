package dto

import "time"

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100" example:"纯棉T恤"`
	Description string `json:"description" example:"柔软透气的纯棉T恤"`
	CategoryID  *int   `json:"category_id" example:"2"`
	Status      string `json:"status" binding:"omitempty,oneof=on_shelf off_shelf" example:"off_shelf"`
}

type UpdateProductRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100" example:"纯棉T恤"`
	Description string `json:"description" example:"柔软透气的纯棉T恤"`
	CategoryID  *int   `json:"category_id" example:"2"`
	Status      string `json:"status" binding:"required,oneof=on_shelf off_shelf" example:"on_shelf"`
}

type ProductResponse struct {
	ID            int       `json:"id" example:"1"`
	Name          string    `json:"name" example:"纯棉T恤"`
	Description   string    `json:"description" example:"柔软透气的纯棉T恤"`
	Status        string    `json:"status" example:"on_shelf"`
	CategoryID    *int      `json:"category_id" example:"2"`
	CategoryName  *string   `json:"category_name" example:"服饰"`
	SKUCount      int64     `json:"sku_count" example:"3"`
	TotalStock    int64     `json:"total_stock" example:"120"`
	MinPriceCents *int64    `json:"min_price_cents" example:"1999"`
	MaxPriceCents *int64    `json:"max_price_cents" example:"2999"`
	CreatedAt     time.Time `json:"created_at" example:"2026-11-09T12:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2026-11-09T12:00:00Z"`
}
