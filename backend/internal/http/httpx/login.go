package httpx

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login maps a JSON login request to an authentication result.
func Login(
	c *gin.Context,
	authenticate func(ctx context.Context, username, password string) (string, error),
	invalidCredentials error,
	banned error,
) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: InvalidJSONMessage,
		})
		return
	}

	token, err := authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, invalidCredentials):
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error: invalidCredentials.Error(),
			})
		case errors.Is(err, banned):
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error: banned.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: InternalServerMessage,
			})
		}
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Token: token,
	})
}
