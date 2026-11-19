package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/feature/shared/httpx"
	"strings"

	"github.com/gin-gonic/gin"
)

// sessionAuthenticator resolves a session token to its subject.
type sessionAuthenticator[T any] interface {
	Authenticate(ctx context.Context, token string) (*T, error)
}

type sessionAuthConfig[T any] struct {
	name            string
	contextKey      string
	contextTokenKey string
	invalidSession  error
	banned          error
	unauthorized    error
	internalServer  error
	subjectIDKey    string
	subjectID       func(*T) int
}

// authenticate validates the Bearer session token and exposes the authenticated
// subject and token on the context.
func authenticate[T any](authenticator sessionAuthenticator[T], config sessionAuthConfig[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortWithError(c, http.StatusUnauthorized, config.unauthorized)
			return
		}

		subject, err := authenticator.Authenticate(c.Request.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, config.unauthorized):
				abortWithError(c, http.StatusUnauthorized, config.unauthorized)
			case errors.Is(err, config.invalidSession):
				abortWithError(c, http.StatusUnauthorized, config.invalidSession)
			case errors.Is(err, config.banned):
				abortWithError(c, http.StatusForbidden, config.banned)
			default:
				slog.ErrorContext(c.Request.Context(), "authenticate session", "middleware", config.name, "error", err)
				abortWithError(c, http.StatusInternalServerError, config.internalServer)
			}
			return
		}

		c.Set(config.contextKey, subject)
		c.Set(config.contextTokenKey, token)
		if config.subjectIDKey != "" {
			c.Set(config.subjectIDKey, config.subjectID(subject))
		}
		c.Next()
	}
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
	c.AbortWithStatusJSON(status, httpx.ErrorResponse{
		Error: err.Error(),
	})
}
