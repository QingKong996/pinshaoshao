package order

import (
	"context"
	"time"

	"pinshaoshao/internal/database"
	inventoryfeature "pinshaoshao/internal/feature/inventory"
	ordercontroller "pinshaoshao/internal/feature/order/controller"
	"pinshaoshao/internal/feature/order/paymentport"
	orderrepository "pinshaoshao/internal/feature/order/repository"
	orderservice "pinshaoshao/internal/feature/order/service"
	"pinshaoshao/internal/feature/order/shippingport"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	controller        *ordercontroller.OrderController
	orderService      *orderservice.OrderService
	paymentLifecycle  paymentport.Lifecycle
	shippingLifecycle shippingport.Lifecycle
	expiryWorker      *orderservice.ExpiryWorker
}

// NewModule wires the order feature and its stock reservation dependency.
func NewModule(
	pool *pgxpool.Pool,
	reservation inventoryfeature.Reservation,
	outboxWriter orderservice.OutboxWriter,
) *Module {
	orderRepository := orderrepository.NewOrderRepository(pool, reservation)
	orderService := orderservice.NewOrderService(
		orderRepository,
		reservation,
		nil,
		outboxWriter,
		database.NewTransactor(pool),
	)
	return &Module{
		controller:        ordercontroller.NewOrderController(orderService),
		orderService:      orderService,
		paymentLifecycle:  orderrepository.NewPaymentLifecycle(),
		shippingLifecycle: orderrepository.NewShippingLifecycle(pool, reservation),
		expiryWorker:      orderservice.NewExpiryWorker(orderService, time.Minute, 100),
	}
}

func (m *Module) OrderService() *orderservice.OrderService {
	return m.orderService
}

func (m *Module) SetPaymentCloser(closer orderservice.PaymentCloser) {
	m.orderService.SetPaymentCloser(closer)
}

func (m *Module) StartBackgroundWorkers(ctx context.Context) {
	m.expiryWorker.Start(ctx)
}

func (m *Module) StopBackgroundWorkers() {
	m.expiryWorker.Stop()
}

// PaymentLifecycle returns the order state operations used by the payment feature.
func (m *Module) PaymentLifecycle() paymentport.Lifecycle {
	return m.paymentLifecycle
}

// ShippingLifecycle returns the order state operations used by the logistics feature.
func (m *Module) ShippingLifecycle() shippingport.Lifecycle {
	return m.shippingLifecycle
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

// RegisterMerchant mounts the shop-scoped order routes on the engine.
func (m *Module) RegisterMerchant(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1/merchant")
	protected.Use(authMiddleware)
	protected.GET("/orders", m.controller.ListMerchant)
	protected.GET("/orders/:subOrderID", m.controller.GetMerchant)
}
