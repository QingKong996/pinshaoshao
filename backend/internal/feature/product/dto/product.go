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

type ListProductsQuery struct {
	Page          *int   `form:"page" binding:"omitempty,min=1,max=1000000" example:"1"`
	PageSize      *int   `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	Sort          string `form:"sort" binding:"omitempty,oneof=created_at updated_at name category status sku_count total_stock min_price" example:"created_at"`
	Order         string `form:"order" binding:"omitempty,oneof=asc desc" example:"desc"`
	Search        string `form:"search" binding:"omitempty,max=100" example:"T-shirt"`
	Status        string `form:"status" binding:"omitempty,oneof=on_shelf off_shelf" example:"on_shelf"`
	CategoryID    *int   `form:"category_id" binding:"omitempty,min=1" example:"2"`
	Uncategorized bool   `form:"uncategorized" example:"false"`
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

type ProductListResponse struct {
	Items      []ProductResponse `json:"items"`
	Page       int               `json:"page" example:"1"`
	PageSize   int               `json:"page_size" example:"10"`
	Total      int64             `json:"total" example:"25"`
	TotalPages int               `json:"total_pages" example:"3"`
}
