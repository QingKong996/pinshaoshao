package router

import (
	"github.com/gin-gonic/gin"
)

// Module registers a module's routes.
type Module interface {
	Register(r *gin.Engine, authMiddleware gin.HandlerFunc)
}

// Modules contains the registered modules and their authentication middleware.
type Modules struct {
	User        Module
	Shop        Module
	Product     Module
	Inventory   Module
	Cart        Module
	Feed        Module
	Order       Module
	Payment     Module
	MockPayment Module
	UserAuth    gin.HandlerFunc
	ShopAuth    gin.HandlerFunc
}

// Register mounts every module.
func Register(r *gin.Engine, modules Modules) {
	modules.User.Register(r, modules.UserAuth)
	modules.Shop.Register(r, modules.ShopAuth)
	modules.Product.Register(r, modules.ShopAuth)
	modules.Inventory.Register(r, modules.ShopAuth)
	modules.Cart.Register(r, modules.UserAuth)
	modules.Feed.Register(r, modules.UserAuth)
	modules.Order.Register(r, modules.UserAuth)
	modules.Payment.Register(r, modules.UserAuth)
	modules.MockPayment.Register(r, nil)
}
