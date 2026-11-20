package shop

import (
	"context"

	shopcontroller "pinshaoshao/internal/feature/shop/controller"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	shoprepository "pinshaoshao/internal/feature/shop/repository"
	shopservice "pinshaoshao/internal/feature/shop/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionAuthenticator resolves shop session tokens.
type SessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*shopmodel.Shop, error)
}

// Module wires the shop feature and owns its routes.
type Module struct {
	shopController      *shopcontroller.ShopController
	dashboardController *shopcontroller.DashboardController
	service             *shopservice.ShopService
}

// NewModule builds the shop feature from its dependencies.
func NewModule(pool *pgxpool.Pool) *Module {
	shopRepository := shoprepository.NewShopRepository(pool)
	shopService := shopservice.NewShopService(shopRepository)
	dashboardRepository := shoprepository.NewDashboardRepository(pool)
	dashboardService := shopservice.NewDashboardService(dashboardRepository)
	return &Module{
		shopController:      shopcontroller.NewShopController(shopService),
		dashboardController: shopcontroller.NewDashboardController(dashboardService),
		service:             shopService,
	}
}

// AuthMiddleware protects shop routes with the module's session authenticator.
func (m *Module) AuthMiddleware() gin.HandlerFunc {
	return AuthMiddleware(m.service)
}

// Register mounts the shop routes on the engine.
func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	merchant := r.Group("/api/v1/merchant")

	merchant.POST("/login", m.shopController.Login)
	merchant.POST("/register", m.shopController.Register)

	protected := merchant.Group("")
	protected.Use(authMiddleware)
	protected.POST("/logout", m.shopController.Logout)
	protected.GET("/shop", m.shopController.GetProfile)
	protected.PATCH("/shop", m.shopController.UpdateProfile)
	protected.GET("/dashboard", m.dashboardController.Overview)
}
