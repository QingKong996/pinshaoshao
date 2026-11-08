package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20" example:"alice01"`
	Password string `json:"password" binding:"required,min=8,max=20" example:"Password123"`
	Email    string `json:"email" binding:"omitempty,email" format:"email" example:"alice@example.com"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
