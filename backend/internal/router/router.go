package router

import (
	"github.com/gin-gonic/gin"
)

// Module registers a module's routes.
type Module interface {
	Register(r *gin.Engine, authMiddleware gin.HandlerFunc)
}

// OrderModule registers both buyer and merchant order routes.
type OrderModule interface {
	Module
	RegisterMerchant(r *gin.Engine, authMiddleware gin.HandlerFunc)
}

// LogisticsModule registers buyer and merchant shipment routes.
type LogisticsModule interface {
	Module
	RegisterMerchant(r *gin.Engine, authMiddleware gin.HandlerFunc)
}

// Modules contains the registered modules and their authentication middleware.
type Modules struct {
	User              Module
	Shop              Module
	Product           Module
	Inventory         Module
	Cart              Module
	Address           Module
	Feed              Module
	Order             OrderModule
	Payment           Module
	Logistics         LogisticsModule
	MockPayment       Module
	MockLogistics     Module
	UserAuth          gin.HandlerFunc
	ShopAuth          gin.HandlerFunc
	MockPaymentAuth   gin.HandlerFunc
	MockLogisticsAuth gin.HandlerFunc
}

var publicAPIRoutes = map[string]struct{}{
	"POST /api/v1/login":             {},
	"POST /api/v1/register":          {},
	"POST /api/v1/merchant/login":    {},
	"POST /api/v1/merchant/register": {},
	"POST /api/v1/payment/webhook":   {},
	"POST /api/v1/logistics/webhook": {},
}

// IsPublicAPI reports whether an API route handles unauthenticated traffic.
func IsPublicAPI(method, route string) bool {
	_, ok := publicAPIRoutes[method+" "+route]
	return ok
}

// Register mounts every module.
func Register(r *gin.Engine, modules Modules) {
	modules.User.Register(r, modules.UserAuth)
	modules.Shop.Register(r, modules.ShopAuth)
	modules.Product.Register(r, modules.ShopAuth)
	modules.Inventory.Register(r, modules.ShopAuth)
	modules.Cart.Register(r, modules.UserAuth)
	modules.Address.Register(r, modules.UserAuth)
	modules.Feed.Register(r, modules.UserAuth)
	modules.Order.Register(r, modules.UserAuth)
	modules.Order.RegisterMerchant(r, modules.ShopAuth)
	modules.Payment.Register(r, modules.UserAuth)
	modules.Logistics.Register(r, modules.UserAuth)
	modules.Logistics.RegisterMerchant(r, modules.ShopAuth)
	modules.MockPayment.Register(r, modules.MockPaymentAuth)
	modules.MockLogistics.Register(r, modules.MockLogisticsAuth)
}
