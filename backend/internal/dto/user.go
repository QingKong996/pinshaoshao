package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20" example:"alice01"`
	Password string `json:"password" binding:"required,min=8,max=20" example:"Password123"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20" example:"alice01"`
	Password string `json:"password" binding:"required,min=8,max=20" example:"Password123"`
	Email    string `json:"email" binding:"omitempty,email" format:"email" example:"alice@example.com"`
}

type TokenResponse struct {
	Token string `json:"token" example:"9f1c2c3b8a7d4e5f6a7b8c9d0e1f2a3b"`
}

type PingResponse struct {
	Message string `json:"message" example:"pong"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
