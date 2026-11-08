package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"strings"

	"github.com/gin-gonic/gin"
)

// sessionAuthenticator resolves a session token to its subject.
type sessionAuthenticator[T any] interface {
	Authenticate(ctx context.Context, token string) (*T, error)
}

type sessionAuthConfig struct {
	name            string
	contextKey      string
	contextTokenKey string
	invalidSession  error
	banned          error
}

// authenticate validates the Bearer session token and exposes the authenticated
// subject and token on the context.
func authenticate[T any](authenticator sessionAuthenticator[T], config sessionAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortWithError(c, http.StatusUnauthorized, apperr.ErrUnauthorized)
			return
		}

		subject, err := authenticator.Authenticate(c.Request.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, apperr.ErrUnauthorized):
				abortWithError(c, http.StatusUnauthorized, apperr.ErrUnauthorized)
			case errors.Is(err, config.invalidSession):
				abortWithError(c, http.StatusUnauthorized, config.invalidSession)
			case errors.Is(err, config.banned):
				abortWithError(c, http.StatusForbidden, config.banned)
			default:
				slog.ErrorContext(c.Request.Context(), "authenticate session", "middleware", config.name, "error", err)
				abortWithError(c, http.StatusInternalServerError, apperr.ErrInternalServer)
			}
			return
		}

		c.Set(config.contextKey, subject)
		c.Set(config.contextTokenKey, token)
		c.Next()
	}
}

func currentSubject[T any](c *gin.Context, key string) (*T, bool) {
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

func sessionToken(c *gin.Context, key string) (string, bool) {
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
	c.AbortWithStatusJSON(status, dto.ErrorResponse{
		Error: err.Error(),
	})
}
