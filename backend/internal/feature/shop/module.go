package shop

import (
	shopcontroller "pinshaoshao/internal/feature/shop/controller"
	shoprepository "pinshaoshao/internal/feature/shop/repository"
	shopservice "pinshaoshao/internal/feature/shop/service"
	"pinshaoshao/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	shopController      *shopcontroller.ShopController
	dashboardController *shopcontroller.DashboardController
	service             *shopservice.ShopService
}

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

func (m *Module) AuthMiddleware() gin.HandlerFunc {
	return middleware.ShopAuth(m.service)
}

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
