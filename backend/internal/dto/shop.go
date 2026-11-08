package dto

type ShopRegisterRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20" example:"shop01"`
	Password string `json:"password" binding:"required,min=8,max=20" example:"Password123"`
	Email    string `json:"email" binding:"required,email" format:"email" example:"shop@example.com"`
}
