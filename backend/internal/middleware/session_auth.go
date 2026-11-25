package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/http/httpx"
	"strings"

	"github.com/gin-gonic/gin"
)

// SessionAuthenticator resolves a session token to its subject.
type SessionAuthenticator[T any] interface {
	Authenticate(ctx context.Context, token string) (*T, error)
}

// SessionAuthConfig maps one session subject type onto the HTTP auth context.
type SessionAuthConfig[T any] struct {
	Name            string
	ContextKey      string
	ContextTokenKey string
	InvalidSession  error
	Banned          error
	Unauthorized    error
	InternalServer  error
	SubjectIDKey    string
	SubjectID       func(*T) int
	// Guard runs after authentication and must not advance the Gin handler chain.
	Guard gin.HandlerFunc
	// UnauthenticatedGuard runs before an unauthorized response.
	UnauthenticatedGuard gin.HandlerFunc
}

// SessionAuth validates the Bearer session token and exposes the authenticated
// subject and token on the context.
func SessionAuth[T any](authenticator SessionAuthenticator[T], config SessionAuthConfig[T]) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			if applyUnauthenticatedGuard(c, config.UnauthenticatedGuard) {
				return
			}
			abortWithError(c, http.StatusUnauthorized, config.Unauthorized)
			return
		}

		subject, err := authenticator.Authenticate(c.Request.Context(), token)
		if err != nil {
			switch {
			case errors.Is(err, config.Unauthorized):
				if applyUnauthenticatedGuard(c, config.UnauthenticatedGuard) {
					return
				}
				abortWithError(c, http.StatusUnauthorized, config.Unauthorized)
			case errors.Is(err, config.InvalidSession):
				if applyUnauthenticatedGuard(c, config.UnauthenticatedGuard) {
					return
				}
				abortWithError(c, http.StatusUnauthorized, config.InvalidSession)
			case errors.Is(err, config.Banned):
				abortWithError(c, http.StatusForbidden, config.Banned)
			default:
				slog.ErrorContext(c.Request.Context(), "authenticate session", "middleware", config.Name, "error", err)
				abortWithError(c, http.StatusInternalServerError, config.InternalServer)
			}
			return
		}

		c.Set(config.ContextKey, subject)
		c.Set(config.ContextTokenKey, token)
		if config.SubjectIDKey != "" {
			c.Set(config.SubjectIDKey, config.SubjectID(subject))
		}
		if config.Guard != nil {
			config.Guard(c)
			if c.IsAborted() {
				return
			}
		}
		c.Next()
	}
}

func applyUnauthenticatedGuard(c *gin.Context, guard gin.HandlerFunc) bool {
	if guard == nil {
		return false
	}
	guard(c)
	return c.IsAborted()
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
