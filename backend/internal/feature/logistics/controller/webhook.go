package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"pinshaoshao/internal/feature/logistics/apperr"
	logisticsdto "pinshaoshao/internal/feature/logistics/dto"
	"pinshaoshao/internal/feature/logistics/model"
	"pinshaoshao/internal/http/httpx"
	"pinshaoshao/internal/webhooksig"

	"github.com/gin-gonic/gin"
)

const (
	headerEventID   = "X-MockLogistics-Event-Id"
	headerEventType = "X-MockLogistics-Event-Type"
	headerTimestamp = "X-MockLogistics-Timestamp"
	headerSignature = "X-MockLogistics-Signature"
	maxWebhookBytes = 1 << 20
)

type CallbackService interface {
	HandleCallback(ctx context.Context, callback model.Callback) (*model.Shipment, error)
}

type WebhookHandler struct {
	service   CallbackService
	secret    string
	tolerance time.Duration
	now       func() time.Time
}

func NewWebhookHandler(
	service CallbackService,
	secret string,
	tolerance time.Duration,
) *WebhookHandler {
	return &WebhookHandler{
		service:   service,
		secret:    secret,
		tolerance: tolerance,
		now:       time.Now,
	}
}

// Handle verifies and applies a signed MockLogistics shipment callback.
//
// @Summary     Handle a logistics callback
// @Description Verify the MockLogistics HMAC signature and record a tracking event.
// @Tags        Logistics
// @Accept      json
// @Produce     json
// @Param       X-MockLogistics-Event-Id header string true "Event ID"
// @Param       X-MockLogistics-Event-Type header string true "Event type"
// @Param       X-MockLogistics-Timestamp header string true "Unix timestamp"
// @Param       X-MockLogistics-Signature header string true "v1 HMAC signature"
// @Param       request body logisticsdto.ShipmentCallbackRequest true "Shipment callback"
// @Success     204
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Router      /logistics/webhook [post]
func (h *WebhookHandler) Handle(ctx *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxWebhookBytes))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCallback.Error()})
		return
	}
	if err := webhooksig.Verify(
		h.secret,
		ctx.GetHeader(headerTimestamp),
		ctx.GetHeader(headerSignature),
		body,
		h.now(),
		h.tolerance,
	); err != nil {
		ctx.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: apperr.ErrInvalidWebhookSignature.Error()})
		return
	}

	var request logisticsdto.ShipmentCallbackRequest
	if err := json.Unmarshal(body, &request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCallback.Error()})
		return
	}
	if request.ID != strings.TrimSpace(ctx.GetHeader(headerEventID)) ||
		request.Type != strings.TrimSpace(ctx.GetHeader(headerEventType)) ||
		request.Type == "" {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCallback.Error()})
		return
	}

	_, err = h.service.HandleCallback(ctx.Request.Context(), model.Callback{
		EventID:            request.ID,
		EventType:          request.Type,
		ShipmentNo:         request.Data.MerchantOrderNo,
		ProviderShipmentNo: request.Data.ShipmentNo,
		ProviderTrackingNo: request.Data.TrackingNo,
		Status:             model.Status(request.Data.Status),
		Sequence:           request.Data.Sequence,
		Description:        request.Data.Description,
		Location:           request.Data.Location,
		OccurredAt:         request.Data.OccurredAt,
	})
	if err != nil {
		writeLogisticsError(ctx, "handle logistics callback", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
