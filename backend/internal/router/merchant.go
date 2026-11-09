package router

import (
	"pinshaoshao/internal/controller"

	"github.com/gin-gonic/gin"
)

// MerchantRegister wires the merchant console endpoints under /api/v1/merchant.
func MerchantRegister(
	r *gin.Engine,
	shopController *controller.ShopController,
	categoryController *controller.ProductCategoryController,
	productController *controller.ProductController,
	skuController *controller.ProductSKUController,
	authMiddleware gin.HandlerFunc,
) {
	merchant := r.Group("/api/v1/merchant")

	protected := merchant.Group("")
	protected.Use(authMiddleware)

	protected.GET("/shop", shopController.GetProfile)
	protected.PATCH("/shop", shopController.UpdateProfile)

	protected.GET("/categories", categoryController.List)
	protected.POST("/categories", categoryController.Create)
	protected.PATCH("/categories/:categoryID", categoryController.Update)
	protected.DELETE("/categories/:categoryID", categoryController.Delete)

	protected.GET("/products", productController.List)
	protected.POST("/products", productController.Create)
	protected.GET("/products/:productID", productController.Get)
	protected.PATCH("/products/:productID", productController.Update)
	protected.DELETE("/products/:productID", productController.Delete)

	protected.GET("/products/:productID/skus", skuController.List)
	protected.POST("/products/:productID/skus", skuController.Create)
	protected.PATCH("/products/:productID/skus/:skuID", skuController.Update)
	protected.DELETE("/products/:productID/skus/:skuID", skuController.Delete)
}
