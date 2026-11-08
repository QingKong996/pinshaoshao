package controller

import (
	"context"
	"errors"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"

	"github.com/gin-gonic/gin"
)

// login handles a login request and maps a banned account to 403.
func login(
	c *gin.Context,
	authenticate func(ctx context.Context, req *dto.LoginRequest) (string, error),
	banned error,
) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	token, err := authenticate(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: apperr.ErrInvalidCredentials.Error(),
			})
		case errors.Is(err, banned):
			c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: banned.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		Token: token,
	})
}
