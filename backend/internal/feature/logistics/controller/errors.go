package controller

import (
	"errors"
	"log/slog"
	"net/http"

	"pinshaoshao/internal/feature/logistics/apperr"
	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

func writeLogisticsError(ctx *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidSubOrderID),
		errors.Is(err, apperr.ErrInvalidShipmentNo),
		errors.Is(err, apperr.ErrInvalidShipmentList),
		errors.Is(err, apperr.ErrInvalidShipmentInput),
		errors.Is(err, apperr.ErrInvalidCallback):
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperr.ErrInvalidWebhookSignature):
		ctx.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: apperr.ErrInvalidWebhookSignature.Error()})
	case errors.Is(err, apperr.ErrShipmentNotFound):
		ctx.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: apperr.ErrShipmentNotFound.Error()})
	case errors.Is(err, apperr.ErrOrderNotShippable),
		errors.Is(err, apperr.ErrShipmentConflict):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperr.ErrProviderUnavailable),
		errors.Is(err, apperr.ErrProviderInvalid):
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
		ctx.JSON(http.StatusBadGateway, httpx.ErrorResponse{Error: apperr.ErrProviderUnavailable.Error()})
	default:
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
		ctx.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
	}
}
