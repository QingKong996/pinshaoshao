package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contextKeyUser  = "middleware.auth.user"
	contextKeyToken = "middleware.auth.token"
)

// SessionAuthenticator resolves a session token to its user.
type SessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*model.User, error)
}

// Auth validates the Bearer session token and exposes the authenticated user
// and token on the context.
func Auth(authenticator SessionAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortWithError(c, http.StatusUnauthorized, apperr.ErrUnauthorized)
			return
		}

		user, err := authenticator.Authenticate(c.Request.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, apperr.ErrUnauthorized):
				abortWithError(c, http.StatusUnauthorized, apperr.ErrUnauthorized)
			case errors.Is(err, apperr.ErrInvalidSession):
				abortWithError(c, http.StatusUnauthorized, apperr.ErrInvalidSession)
			case errors.Is(err, apperr.ErrUserBanned):
				abortWithError(c, http.StatusForbidden, apperr.ErrUserBanned)
			default:
				log.Printf("middleware.Auth: authenticate: %v", err)
				abortWithError(c, http.StatusInternalServerError, apperr.ErrInternalServer)
			}
			return
		}

		c.Set(contextKeyUser, user)
		c.Set(contextKeyToken, token)
		c.Next()
	}
}

// CurrentUser returns the user set by Auth.
func CurrentUser(c *gin.Context) (*model.User, bool) {
	value, ok := c.Get(contextKeyUser)
	if !ok {
		return nil, false
	}
	user, ok := value.(*model.User)
	if !ok || user == nil {
		return nil, false
	}
	return user, true
}

// SessionToken returns the raw session token set by Auth.
func SessionToken(c *gin.Context) (string, bool) {
	value, ok := c.Get(contextKeyToken)
	if !ok {
		return "", false
	}
	token, ok := value.(string)
	if !ok || token == "" {
		return "", false
	}
	return token, true
}

// MustSessionToken returns the raw session token set by Auth. It panics when
// the route is not protected by Auth, which is a routing bug.
func MustSessionToken(c *gin.Context) string {
	token, ok := SessionToken(c)
	if !ok {
		panic("middleware.MustSessionToken: route is not protected by Auth")
	}
	return token
}

// bearerToken parses an "Authorization: Bearer <token>" header.
func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	token = strings.TrimSpace(token)
	if token == "" || strings.ContainsAny(token, " \t\r\n") {
		return "", false
	}
	return token, true
}

func abortWithError(c *gin.Context, status int, err error) {
	if status == http.StatusUnauthorized {
		c.Header("WWW-Authenticate", "Bearer")
	}
	c.AbortWithStatusJSON(status, gin.H{
		"error": err.Error(),
	})
}
