package router

import (
	"pinshaoshao/internal/controller"

	"github.com/gin-gonic/gin"
)

// ShopRegister wires the merchant authentication endpoints under /api/v1/merchant.
func ShopRegister(r *gin.Engine, s *controller.ShopController, authMiddleware gin.HandlerFunc) {
	api := r.Group("/api")
	merchant := api.Group("/v1/merchant")

	merchant.POST("/login", s.Login)
	merchant.POST("/register", s.Register)

	protected := merchant.Group("")
	protected.Use(authMiddleware)
	protected.POST("/logout", s.Logout)
}
