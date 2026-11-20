package authctx

import "github.com/gin-gonic/gin"

const (
	UserSubjectKey = "auth.user"
	UserIDKey      = "auth.user.id"
	UserTokenKey   = "auth.user.token"
	ShopSubjectKey = "auth.shop"
	ShopIDKey      = "auth.shop.id"
	ShopTokenKey   = "auth.shop.token"
)

// Subject returns the authenticated subject stored under key.
func Subject[T any](c *gin.Context, key string) (*T, bool) {
	value, ok := c.Get(key)
	if !ok {
		return nil, false
	}
	subject, ok := value.(*T)
	if !ok || subject == nil {
		return nil, false
	}
	return subject, true
}

// MustSubject returns the authenticated subject or panics when middleware did not set it.
func MustSubject[T any](c *gin.Context, key, panicMessage string) *T {
	subject, ok := Subject[T](c, key)
	if !ok {
		panic(panicMessage)
	}
	return subject
}

// Token returns the session token stored under key.
func Token(c *gin.Context, key string) (string, bool) {
	value, ok := c.Get(key)
	if !ok {
		return "", false
	}
	token, ok := value.(string)
	if !ok || token == "" {
		return "", false
	}
	return token, true
}

// MustToken returns the session token or panics when middleware did not set it.
func MustToken(c *gin.Context, key, panicMessage string) string {
	token, ok := Token(c, key)
	if !ok {
		panic(panicMessage)
	}
	return token
}

// UserID returns the authenticated user ID.
func UserID(c *gin.Context) (int, bool) {
	value, ok := c.Get(UserIDKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(int)
	return userID, ok && userID > 0
}

// MustUserID returns the authenticated user ID or panics when the route is not protected by UserAuth.
func MustUserID(c *gin.Context) int {
	userID, ok := UserID(c)
	if !ok {
		panic("authctx.MustUserID: route is not protected by UserAuth")
	}
	return userID
}

// ShopID returns the authenticated shop ID.
func ShopID(c *gin.Context) (int, bool) {
	value, ok := c.Get(ShopIDKey)
	if !ok {
		return 0, false
	}
	shopID, ok := value.(int)
	return shopID, ok && shopID > 0
}

// MustShopID returns the authenticated shop ID or panics when middleware did not set it.
func MustShopID(c *gin.Context) int {
	shopID, ok := ShopID(c)
	if !ok {
		panic("authctx.MustShopID: route is not protected by ShopAuth")
	}
	return shopID
}
