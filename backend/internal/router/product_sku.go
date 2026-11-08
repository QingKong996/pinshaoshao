package router

import (
	"pinshaoshao/internal/controller"

	"github.com/gin-gonic/gin"
)

// ProductSKURegister wires the product SKU endpoints.
func ProductSKURegister(r *gin.Engine, c *controller.ProductSKUController, authMiddleware gin.HandlerFunc) {
	api := r.Group("/api")
	v1 := api.Group("/v1")

	v1.GET("/products/:productID/skus", c.List)

	protected := v1.Group("")
	protected.Use(authMiddleware)
	protected.POST("/products/:productID/skus", c.Create)
}
