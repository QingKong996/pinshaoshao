package httpx

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Logout revokes the session token and writes the HTTP response.
func Logout(c *gin.Context, token string, revoke func(ctx context.Context, token string) error) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	if err := revoke(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: InternalServerMessage,
		})
		return
	}

	c.Status(http.StatusNoContent)
}
