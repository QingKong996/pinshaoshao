package cart

import (
	cartcontroller "pinshaoshao/internal/feature/cart/controller"
	cartrepository "pinshaoshao/internal/feature/cart/repository"
	cartservice "pinshaoshao/internal/feature/cart/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	controller *cartcontroller.CartController
}

func NewModule(pool *pgxpool.Pool) *Module {
	cartRepository := cartrepository.NewCartRepository(pool)
	cartService := cartservice.NewCartService(cartRepository)
	return &Module{
		controller: cartcontroller.NewCartController(cartService),
	}
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware)
	protected.GET("/cart", m.controller.List)
	protected.POST("/cart/items", m.controller.Add)
	protected.PATCH("/cart/items/:itemID", m.controller.UpdateQuantity)
	protected.DELETE("/cart/items/:itemID", m.controller.Delete)
}
