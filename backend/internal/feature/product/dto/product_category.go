package dto

import "time"

type CreateProductCategoryRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50" example:"服饰"`
}

type UpdateProductCategoryRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50" example:"服饰"`
}

type ProductCategoryResponse struct {
	ID           int       `json:"id" example:"1"`
	Name         string    `json:"name" example:"服饰"`
	ProductCount int64     `json:"product_count" example:"3"`
	CreatedAt    time.Time `json:"created_at" example:"2026-11-09T12:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2026-11-09T12:00:00Z"`
}
