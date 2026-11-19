package product

import (
	productcontroller "pinshaoshao/internal/feature/product/controller"
	productrepository "pinshaoshao/internal/feature/product/repository"
	productservice "pinshaoshao/internal/feature/product/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	categoryController *productcontroller.ProductCategoryController
	productController  *productcontroller.ProductController
	skuController      *productcontroller.ProductSKUController
}

func NewModule(pool *pgxpool.Pool) *Module {
	productRepository := productrepository.NewProductRepository(pool)
	categoryRepository := productrepository.NewProductCategoryRepository(pool)
	skuRepository := productrepository.NewProductSKURepository(pool)

	return &Module{
		categoryController: productcontroller.NewProductCategoryController(
			productservice.NewProductCategoryService(categoryRepository),
		),
		productController: productcontroller.NewProductController(
			productservice.NewProductService(productRepository, categoryRepository),
		),
		skuController: productcontroller.NewProductSKUController(
			productservice.NewProductSKUService(productRepository, skuRepository),
		),
	}
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/api/v1/merchant")
	protected := merchant.Group("")
	protected.Use(authMiddleware)

	protected.GET("/categories", m.categoryController.List)
	protected.POST("/categories", m.categoryController.Create)
	protected.PATCH("/categories/:categoryID", m.categoryController.Update)
	protected.DELETE("/categories/:categoryID", m.categoryController.Delete)

	protected.GET("/products", m.productController.List)
	protected.POST("/products", m.productController.Create)
	protected.GET("/products/:productID", m.productController.Get)
	protected.PATCH("/products/:productID", m.productController.Update)
	protected.DELETE("/products/:productID", m.productController.Delete)

	protected.GET("/products/:productID/skus", m.skuController.List)
	protected.POST("/products/:productID/skus", m.skuController.Create)
	protected.PATCH("/products/:productID/skus/:skuID", m.skuController.Update)
	protected.DELETE("/products/:productID/skus/:skuID", m.skuController.Delete)
}
