package middleware

import (
	"context"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
)

const (
	contextKeyShop      = "middleware.shopAuth.shop"
	contextKeyShopToken = "middleware.shopAuth.token"
)

// ShopSessionAuthenticator resolves a session token to its shop.
type ShopSessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*model.Shop, error)
}

// ShopAuth validates the Bearer session token and exposes the authenticated shop
// and token on the context.
func ShopAuth(authenticator ShopSessionAuthenticator) gin.HandlerFunc {
	return authenticate[model.Shop](authenticator, sessionAuthConfig{
		name:            "ShopAuth",
		contextKey:      contextKeyShop,
		contextTokenKey: contextKeyShopToken,
		invalidSession:  apperr.ErrInvalidShopSession,
		banned:          apperr.ErrShopBanned,
	})
}

// CurrentShop returns the shop set by ShopAuth.
func CurrentShop(c *gin.Context) (*model.Shop, bool) {
	return currentSubject[model.Shop](c, contextKeyShop)
}

// MustCurrentShop returns the shop set by ShopAuth. It panics when the route is
// not protected by ShopAuth, which is a routing bug.
func MustCurrentShop(c *gin.Context) *model.Shop {
	shop, ok := CurrentShop(c)
	if !ok {
		panic("middleware.MustCurrentShop: route is not protected by ShopAuth")
	}
	return shop
}

// ShopSessionToken returns the raw session token set by ShopAuth.
func ShopSessionToken(c *gin.Context) (string, bool) {
	return sessionToken(c, contextKeyShopToken)
}

// MustShopSessionToken returns the raw session token set by ShopAuth. It panics
// when the route is not protected by ShopAuth, which is a routing bug.
func MustShopSessionToken(c *gin.Context) string {
	token, ok := ShopSessionToken(c)
	if !ok {
		panic("middleware.MustShopSessionToken: route is not protected by ShopAuth")
	}
	return token
}
