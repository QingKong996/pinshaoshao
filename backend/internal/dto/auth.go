package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=5,max=20" example:"alice01"`
	Password string `json:"password" binding:"required,min=8,max=20" example:"Password123"`
}

type TokenResponse struct {
	Token string `json:"token" example:"9f1c2c3b8a7d4e5f6a7b8c9d0e1f2a3b"`
}
