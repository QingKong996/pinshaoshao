package inventory

import (
	inventorycontroller "pinshaoshao/internal/feature/inventory/controller"
	inventoryrepository "pinshaoshao/internal/feature/inventory/repository"
	inventoryservice "pinshaoshao/internal/feature/inventory/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	controller *inventorycontroller.InventoryController
}

func NewModule(pool *pgxpool.Pool) *Module {
	inventoryRepository := inventoryrepository.NewInventoryRepository(pool)
	inventoryService := inventoryservice.NewInventoryService(inventoryRepository)
	return &Module{
		controller: inventorycontroller.NewInventoryController(inventoryService),
	}
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/api/v1/merchant")
	protected := merchant.Group("")
	protected.Use(authMiddleware)
	protected.GET("/inventory", m.controller.List)
}
