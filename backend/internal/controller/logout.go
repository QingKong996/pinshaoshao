package controller

import (
	"context"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"

	"github.com/gin-gonic/gin"
)

// logout revokes the session token and responds with 204 or 500.
func logout(
	c *gin.Context,
	token string,
	revoke func(ctx context.Context, token string) error,
) {
	if err := revoke(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
