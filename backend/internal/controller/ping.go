package controller

import (
	"net/http"
	"pinshaoshao/internal/dto"

	"github.com/gin-gonic/gin"
)

// Ping reports that the server is running.
//
// @Summary     Health check
// @Description Check whether the server is running.
// @Tags        System
// @Produce     json
// @Success     200 {object} dto.PingResponse
// @Router      /ping [get]
func Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, dto.PingResponse{
		Message: "pong",
	})
}
