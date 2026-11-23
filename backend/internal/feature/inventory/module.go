package inventory

import (
	inventorycontroller "pinshaoshao/internal/feature/inventory/controller"
	inventoryrepository "pinshaoshao/internal/feature/inventory/repository"
	inventoryservice "pinshaoshao/internal/feature/inventory/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module wires the inventory feature and owns its routes.
type Module struct {
	controller  *inventorycontroller.InventoryController
	lifecycle   Lifecycle
	reservation Reservation
}

// NewModule builds the inventory feature from its dependencies.
func NewModule(pool *pgxpool.Pool) *Module {
	inventoryRepository := inventoryrepository.NewInventoryRepository(pool)
	inventoryService := inventoryservice.NewInventoryService(inventoryRepository)
	return &Module{
		controller:  inventorycontroller.NewInventoryController(inventoryService),
		lifecycle:   inventoryRepository,
		reservation: inventoryRepository,
	}
}

// Lifecycle returns the inventory operations used by catalog writes.
func (m *Module) Lifecycle() Lifecycle {
	return m.lifecycle
}

// Reservation returns the stock operations used by order placement and cancellation.
func (m *Module) Reservation() Reservation {
	return m.reservation
}

// Register mounts the inventory routes on the engine.
func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/api/v1/merchant")
	protected := merchant.Group("")
	protected.Use(authMiddleware)
	protected.GET("/inventory", m.controller.List)
	protected.PATCH("/inventory/:skuID", m.controller.Update)
}
