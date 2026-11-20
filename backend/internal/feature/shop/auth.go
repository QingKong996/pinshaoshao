package shop

import (
	"pinshaoshao/internal/feature/shop/apperr"
	"pinshaoshao/internal/feature/shop/model"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates a shop session token and stores the shop on the context.
func AuthMiddleware(authenticator SessionAuthenticator) gin.HandlerFunc {
	return middleware.SessionAuth(authenticator, middleware.SessionAuthConfig[model.Shop]{
		Name:            "ShopAuth",
		ContextKey:      authctx.ShopSubjectKey,
		ContextTokenKey: authctx.ShopTokenKey,
		InvalidSession:  apperr.ErrInvalidShopSession,
		Banned:          apperr.ErrShopBanned,
		Unauthorized:    apperr.ErrUnauthorized,
		InternalServer:  apperr.ErrInternalServer,
		SubjectIDKey:    authctx.ShopIDKey,
		SubjectID:       func(shop *model.Shop) int { return shop.ID },
	})
}

// CurrentShop returns the shop set by AuthMiddleware.
func CurrentShop(c *gin.Context) (*model.Shop, bool) {
	return authctx.Subject[model.Shop](c, authctx.ShopSubjectKey)
}

// MustCurrentShop returns the shop set by AuthMiddleware. It panics when the route
// is not protected by AuthMiddleware, which is a routing bug.
func MustCurrentShop(c *gin.Context) *model.Shop {
	shop, ok := CurrentShop(c)
	if !ok {
		panic("shop.MustCurrentShop: route is not protected by AuthMiddleware")
	}
	return shop
}

// ShopSessionToken returns the raw session token set by AuthMiddleware.
func ShopSessionToken(c *gin.Context) (string, bool) {
	return authctx.Token(c, authctx.ShopTokenKey)
}

// MustShopSessionToken returns the raw session token set by AuthMiddleware. It
// panics when the route is not protected by AuthMiddleware, which is a routing bug.
func MustShopSessionToken(c *gin.Context) string {
	return authctx.MustToken(c, authctx.ShopTokenKey, "shop.MustShopSessionToken: route is not protected by AuthMiddleware")
}
