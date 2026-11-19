package middleware

import (
	"context"
	"pinshaoshao/internal/feature/shared/authctx"
	"pinshaoshao/internal/feature/shop/apperr"
	"pinshaoshao/internal/feature/shop/model"

	"github.com/gin-gonic/gin"
)

// ShopSessionAuthenticator resolves a session token to its shop.
type ShopSessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*model.Shop, error)
}

// ShopAuth validates the Bearer session token and exposes the authenticated shop
// and token on the context.
func ShopAuth(authenticator ShopSessionAuthenticator) gin.HandlerFunc {
	return authenticate[model.Shop](authenticator, sessionAuthConfig[model.Shop]{
		name:            "ShopAuth",
		contextKey:      authctx.ShopSubjectKey,
		contextTokenKey: authctx.ShopTokenKey,
		invalidSession:  apperr.ErrInvalidShopSession,
		banned:          apperr.ErrShopBanned,
		unauthorized:    apperr.ErrUnauthorized,
		internalServer:  apperr.ErrInternalServer,
		subjectIDKey:    authctx.ShopIDKey,
		subjectID:       func(shop *model.Shop) int { return shop.ID },
	})
}

// CurrentShop returns the shop set by ShopAuth.
func CurrentShop(c *gin.Context) (*model.Shop, bool) {
	return authctx.Subject[model.Shop](c, authctx.ShopSubjectKey)
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
	return authctx.Token(c, authctx.ShopTokenKey)
}

// MustShopSessionToken returns the raw session token set by ShopAuth. It panics
// when the route is not protected by ShopAuth, which is a routing bug.
func MustShopSessionToken(c *gin.Context) string {
	return authctx.MustToken(c, authctx.ShopTokenKey, "middleware.MustShopSessionToken: route is not protected by ShopAuth")
}
