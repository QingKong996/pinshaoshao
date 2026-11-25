package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const sessionTokenContextKey = "mock.auth.session_token"

type authService interface {
	Login(ctx context.Context, username, password string) (string, error)
	Authenticate(ctx context.Context, token string) (*Account, error)
	Logout(ctx context.Context, token string) error
}

type Handler struct {
	service authService
}

func NewHandler(service authService) *Handler {
	return &Handler{service: service}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: ErrInvalidJSON.Error()})
		return
	}

	token, err := h.service.Login(
		c.Request.Context(),
		request.Username,
		request.Password,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, errorResponse{
				Error: ErrInvalidCredentials.Error(),
			})
			return
		}
		slog.ErrorContext(c.Request.Context(), "login mock service", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: ErrInternalServer.Error(),
		})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, tokenResponse{Token: token})
}

func (h *Handler) Logout(c *gin.Context) {
	token, ok := SessionToken(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
			Error: ErrUnauthorized.Error(),
		})
		return
	}
	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		slog.ErrorContext(c.Request.Context(), "logout mock service", "error", err)
		c.JSON(http.StatusInternalServerError, errorResponse{
			Error: ErrInternalServer.Error(),
		})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortUnauthorized(c, ErrUnauthorized)
			return
		}
		if _, err := h.service.Authenticate(c.Request.Context(), token); err != nil {
			if errors.Is(err, ErrUnauthorized) ||
				errors.Is(err, ErrInvalidSession) {
				abortUnauthorized(c, ErrInvalidSession)
				return
			}
			slog.ErrorContext(c.Request.Context(), "authenticate mock service", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{
				Error: ErrInternalServer.Error(),
			})
			return
		}
		c.Set(sessionTokenContextKey, token)
		c.Next()
	}
}

func SessionToken(c *gin.Context) (string, bool) {
	value, ok := c.Get(sessionTokenContextKey)
	if !ok {
		return "", false
	}
	token, ok := value.(string)
	return token, ok && token != ""
}

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

func abortUnauthorized(c *gin.Context, err error) {
	c.Header("WWW-Authenticate", "Bearer")
	c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{
		Error: err.Error(),
	})
}
