package dto

type LoginRequest struct {
	Username string `json:"username", binding:"required,min=5,max=20"`
	Password string `json:"password", binding:"required,min=8,max=20"`
}

type RegisterRequest struct {
	Username string `json:"username", binding:"required,min=5,max=20"`
	Password string `json:"password", binding:"required,min=8,max=20"`
	Email    string `json:"email", binding:"required,email"`
}
