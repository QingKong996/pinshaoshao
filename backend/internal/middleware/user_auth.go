package middleware

import (
	"context"
	"pinshaoshao/internal/feature/shared/authctx"
	"pinshaoshao/internal/feature/user/apperr"
	"pinshaoshao/internal/feature/user/model"

	"github.com/gin-gonic/gin"
)

// UserSessionAuthenticator resolves a session token to its user.
type UserSessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*model.User, error)
}

// UserAuth validates the Bearer session token and exposes the authenticated user
// and token on the context.
func UserAuth(authenticator UserSessionAuthenticator) gin.HandlerFunc {
	return authenticate[model.User](authenticator, sessionAuthConfig[model.User]{
		name:            "UserAuth",
		contextKey:      authctx.UserSubjectKey,
		contextTokenKey: authctx.UserTokenKey,
		invalidSession:  apperr.ErrInvalidSession,
		banned:          apperr.ErrUserBanned,
		unauthorized:    apperr.ErrUnauthorized,
		internalServer:  apperr.ErrInternalServer,
		subjectIDKey:    authctx.UserIDKey,
		subjectID:       func(user *model.User) int { return user.ID },
	})
}

// CurrentUser returns the user set by UserAuth.
func CurrentUser(c *gin.Context) (*model.User, bool) {
	return authctx.Subject[model.User](c, authctx.UserSubjectKey)
}

// UserSessionToken returns the raw session token set by UserAuth.
func UserSessionToken(c *gin.Context) (string, bool) {
	return authctx.Token(c, authctx.UserTokenKey)
}

// MustUserSessionToken returns the raw session token set by UserAuth. It panics
// when the route is not protected by UserAuth, which is a routing bug.
func MustUserSessionToken(c *gin.Context) string {
	return authctx.MustToken(c, authctx.UserTokenKey, "middleware.MustUserSessionToken: route is not protected by UserAuth")
}
