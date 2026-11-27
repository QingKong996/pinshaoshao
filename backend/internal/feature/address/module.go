package address

import (
	addresscontroller "pinshaoshao/internal/feature/address/controller"
	addressrepository "pinshaoshao/internal/feature/address/repository"
	addressservice "pinshaoshao/internal/feature/address/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Module struct {
	controller *addresscontroller.AddressController
}

func NewModule(pool *pgxpool.Pool) *Module {
	addressRepository := addressrepository.NewAddressRepository(pool)
	addressService := addressservice.NewAddressService(addressRepository)
	return &Module{
		controller: addresscontroller.NewAddressController(addressService),
	}
}

func (m *Module) Register(r *gin.Engine, authMiddleware gin.HandlerFunc) {
	protected := r.Group("/api/v1")
	protected.Use(authMiddleware)
	protected.GET("/addresses", m.controller.List)
	protected.POST("/addresses", m.controller.Create)
	protected.PUT("/addresses/:addressID", m.controller.Update)
	protected.DELETE("/addresses/:addressID", m.controller.Delete)
	protected.POST("/addresses/:addressID/default", m.controller.SetDefault)
}
