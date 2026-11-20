package user

import (
	"pinshaoshao/internal/feature/user/apperr"
	"pinshaoshao/internal/feature/user/model"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates a user session token and stores the user on the context.
func AuthMiddleware(authenticator SessionAuthenticator) gin.HandlerFunc {
	return middleware.SessionAuth(authenticator, middleware.SessionAuthConfig[model.User]{
		Name:            "UserAuth",
		ContextKey:      authctx.UserSubjectKey,
		ContextTokenKey: authctx.UserTokenKey,
		InvalidSession:  apperr.ErrInvalidSession,
		Banned:          apperr.ErrUserBanned,
		Unauthorized:    apperr.ErrUnauthorized,
		InternalServer:  apperr.ErrInternalServer,
		SubjectIDKey:    authctx.UserIDKey,
		SubjectID:       func(user *model.User) int { return user.ID },
	})
}

// CurrentUser returns the user set by AuthMiddleware.
func CurrentUser(c *gin.Context) (*model.User, bool) {
	return authctx.Subject[model.User](c, authctx.UserSubjectKey)
}

// UserSessionToken returns the raw session token set by AuthMiddleware.
func UserSessionToken(c *gin.Context) (string, bool) {
	return authctx.Token(c, authctx.UserTokenKey)
}

// MustUserSessionToken returns the raw session token set by AuthMiddleware. It
// panics when the route is not protected by AuthMiddleware, which is a routing bug.
func MustUserSessionToken(c *gin.Context) string {
	return authctx.MustToken(c, authctx.UserTokenKey, "user.MustUserSessionToken: route is not protected by AuthMiddleware")
}
