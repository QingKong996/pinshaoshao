package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/dto"
	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/http/httpx"
	"pinshaoshao/internal/webhooksig"

	"github.com/gin-gonic/gin"
)

const (
	headerEventID    = "X-MockPay-Event-Id"
	headerEventType  = "X-MockPay-Event-Type"
	headerTimestamp  = "X-MockPay-Timestamp"
	headerSignature  = "X-MockPay-Signature"
	paymentPaidEvent = "payment.paid"
	maxWebhookBytes  = 1 << 20
)

type PaymentCallbackService interface {
	HandlePaidCallback(ctx context.Context, request model.PaidCallback) (*model.Payment, error)
}

type WebhookHandler struct {
	service   PaymentCallbackService
	secret    string
	tolerance time.Duration
	now       func() time.Time
}

func NewWebhookHandler(
	service PaymentCallbackService,
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

// HandlePaymentCallback verifies and applies a signed MockPay payment callback.
//
// @Summary     Handle a payment callback
// @Description Verify the MockPay HMAC signature and settle an order exactly once.
// @Tags        Payments
// @Accept      json
// @Produce     json
// @Param       X-MockPay-Event-Id header string true "Event ID"
// @Param       X-MockPay-Event-Type header string true "Event type"
// @Param       X-MockPay-Timestamp header string true "Unix timestamp"
// @Param       X-MockPay-Signature header string true "v1 HMAC signature"
// @Param       request body dto.PaidCallbackRequest true "Payment callback"
// @Success     204
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Router      /payment/webhook [post]
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

	var request dto.PaidCallbackRequest
	if err := json.Unmarshal(body, &request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCallback.Error()})
		return
	}
	if request.Type != paymentPaidEvent ||
		ctx.GetHeader(headerEventType) != paymentPaidEvent ||
		request.ID != strings.TrimSpace(ctx.GetHeader(headerEventID)) ||
		request.Data.Status != "paid" {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCallback.Error()})
		return
	}

	_, err = h.service.HandlePaidCallback(ctx.Request.Context(), model.PaidCallback{
		EventID:               request.ID,
		EventType:             request.Type,
		ProviderTransactionID: request.Data.PaymentNo,
		PaymentNo:             request.Data.MerchantOrderNo,
		AmountCents:           request.Data.AmountCents,
		Currency:              request.Data.Currency,
		PaidAt:                request.Data.PaidAt,
	})
	if err != nil {
		writePaymentError(ctx, "handle payment callback", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
