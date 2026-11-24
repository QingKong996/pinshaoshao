package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"pinshaoshao/internal/http/httpx"
	"pinshaoshao/internal/mock/logistics/model"

	"github.com/gin-gonic/gin"
)

func writeMockLogisticsError(ctx *gin.Context, operation string, err error) {
	status, message := httpStatusForError(err)
	if status == http.StatusInternalServerError {
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
	}
	ctx.JSON(status, httpx.ErrorResponse{Error: message})
}

func httpStatusForError(err error) (int, string) {
	switch {
	case errors.Is(err, model.ErrInvalidShipmentInput):
		return http.StatusBadRequest, model.ErrInvalidShipmentInput.Error()
	case errors.Is(err, model.ErrInvalidShipmentNumber):
		return http.StatusBadRequest, model.ErrInvalidShipmentNumber.Error()
	case errors.Is(err, model.ErrInvalidNotificationID):
		return http.StatusBadRequest, model.ErrInvalidNotificationID.Error()
	case errors.Is(err, model.ErrShipmentNotFound):
		return http.StatusNotFound, model.ErrShipmentNotFound.Error()
	case errors.Is(err, model.ErrShipmentConflict):
		return http.StatusConflict, model.ErrShipmentConflict.Error()
	case errors.Is(err, model.ErrInvalidShipmentTransition):
		return http.StatusConflict, model.ErrInvalidShipmentTransition.Error()
	case errors.Is(err, model.ErrNotificationNotFound):
		return http.StatusNotFound, model.ErrNotificationNotFound.Error()
	case errors.Is(err, model.ErrNotificationNotRetryable):
		return http.StatusConflict, model.ErrNotificationNotRetryable.Error()
	default:
		return http.StatusInternalServerError, httpx.InternalServerMessage
	}
}
