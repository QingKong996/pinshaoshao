package order

import (
	inventoryfeature "pinshaoshao/internal/feature/inventory"
	ordercontroller "pinshaoshao/internal/feature/order/controller"
	orderrepository "pinshaoshao/internal/feature/order/repository"
	orderservice "pinshaoshao/internal/feature/order/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	controller *ordercontroller.OrderController
}

// NewModule wires the order feature and its stock reservation dependency.
func NewModule(pool *pgxpool.Pool, reservation inventoryfeature.Reservation) *Module {
	orderRepository := orderrepository.NewOrderRepository(pool, reservation)
	orderService := orderservice.NewOrderService(orderRepository)
	return &Module{
		controller: ordercontroller.NewOrderController(orderService),
	}
}

// Register mounts the buyer order routes on the engine.
func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware)
	protected.POST("/orders", m.controller.Create)
	protected.POST("/orders/direct", m.controller.CreateDirect)
	protected.GET("/orders", m.controller.List)
	protected.GET("/orders/:orderID", m.controller.Get)
	protected.POST("/orders/:orderID/cancel", m.controller.Cancel)
}
