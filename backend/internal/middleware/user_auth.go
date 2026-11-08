package middleware

import (
	"context"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
)

const (
	contextKeyUser      = "middleware.userAuth.user"
	contextKeyUserToken = "middleware.userAuth.token"
)

// UserSessionAuthenticator resolves a session token to its user.
type UserSessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*model.User, error)
}

// UserAuth validates the Bearer session token and exposes the authenticated user
// and token on the context.
func UserAuth(authenticator UserSessionAuthenticator) gin.HandlerFunc {
	return authenticate[model.User](authenticator, sessionAuthConfig{
		name:            "UserAuth",
		contextKey:      contextKeyUser,
		contextTokenKey: contextKeyUserToken,
		invalidSession:  apperr.ErrInvalidSession,
		banned:          apperr.ErrUserBanned,
	})
}

// CurrentUser returns the user set by UserAuth.
func CurrentUser(c *gin.Context) (*model.User, bool) {
	return currentSubject[model.User](c, contextKeyUser)
}

// UserSessionToken returns the raw session token set by UserAuth.
func UserSessionToken(c *gin.Context) (string, bool) {
	return sessionToken(c, contextKeyUserToken)
}

// MustUserSessionToken returns the raw session token set by UserAuth. It panics
// when the route is not protected by UserAuth, which is a routing bug.
func MustUserSessionToken(c *gin.Context) string {
	token, ok := UserSessionToken(c)
	if !ok {
		panic("middleware.MustUserSessionToken: route is not protected by UserAuth")
	}
	return token
}
