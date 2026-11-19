package dto

import "time"

type ShopRegisterRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20" example:"shop01"`
	Name     string `json:"name" binding:"required,min=2,max=50" example:"拼少少旗舰店"`
	Password string `json:"password" binding:"required,min=8,max=20" example:"Password123"`
	Email    string `json:"email" binding:"required,email" format:"email" example:"shop@example.com"`
}

type UpdateShopProfileRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=50" example:"拼少少旗舰店"`
	Description *string `json:"description" binding:"required,max=500" example:"店铺简介"`
}

type ShopProfileResponse struct {
	ID          int       `json:"id" example:"1"`
	Username    string    `json:"username" example:"shop01"`
	Name        string    `json:"name" example:"拼少少旗舰店"`
	Description string    `json:"description" example:"店铺简介"`
	Email       string    `json:"email" example:"shop01@example.com"`
	Status      string    `json:"status" example:"active"`
	CreatedAt   time.Time `json:"created_at" example:"2026-11-09T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2026-11-09T12:00:00Z"`
}
