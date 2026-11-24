package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pinshaoshao/internal/http/httpx"
	"pinshaoshao/internal/mockpayment/dto"
	"pinshaoshao/internal/mockpayment/model"
	"pinshaoshao/internal/mockpayment/service"

	"github.com/gin-gonic/gin"
)

const checkoutPathPrefix = "/mock-payment/checkout/"

type PaymentService interface {
	Create(ctx context.Context, input service.CreatePaymentInput) (*model.Payment, bool, error)
	Get(ctx context.Context, paymentNo string) (*model.Payment, error)
	Close(ctx context.Context, paymentNo string) (*model.Payment, error)
}

type PaymentHandler struct {
	service       PaymentService
	publicBaseURL string
}

func NewPaymentHandler(paymentService PaymentService, publicBaseURL string) *PaymentHandler {
	return &PaymentHandler{
		service:       paymentService,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

func (h *PaymentHandler) Create(ctx *gin.Context) {
	var request dto.CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: httpx.InvalidJSONMessage})
		return
	}

	payment, created, err := h.service.Create(ctx.Request.Context(), service.CreatePaymentInput{
		MerchantOrderNo: request.MerchantOrderNo,
		AmountCents:     request.AmountCents,
		Currency:        request.Currency,
		Subject:         request.Subject,
		NotifyURL:       request.NotifyURL,
		ReturnURL:       request.ReturnURL,
		ExpiresIn:       time.Duration(request.ExpiresInSeconds) * time.Second,
	})
	if err != nil {
		writeMockPaymentError(ctx, "create mock payment", err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	ctx.JSON(status, h.toPaymentResponse(payment))
}

func (h *PaymentHandler) Get(ctx *gin.Context) {
	payment, err := h.service.Get(ctx.Request.Context(), ctx.Param("paymentNo"))
	if err != nil {
		writeMockPaymentError(ctx, "get mock payment", err)
		return
	}
	ctx.JSON(http.StatusOK, h.toPaymentResponse(payment))
}

func (h *PaymentHandler) Close(ctx *gin.Context) {
	payment, err := h.service.Close(ctx.Request.Context(), ctx.Param("paymentNo"))
	if err != nil {
		writeMockPaymentError(ctx, "close mock payment", err)
		return
	}
	ctx.JSON(http.StatusOK, h.toPaymentResponse(payment))
}

func (h *PaymentHandler) toPaymentResponse(payment *model.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		PaymentNo:       payment.PaymentNo,
		MerchantOrderNo: payment.MerchantOrderNo,
		Status:          string(payment.Status),
		AmountCents:     payment.AmountCents,
		Currency:        payment.Currency,
		Subject:         payment.Subject,
		CheckoutURL:     h.checkoutURL(payment.PaymentNo),
		ReturnURL:       payment.ReturnURL,
		ExpiresAt:       payment.ExpiresAt,
		PaidAt:          payment.PaidAt,
		ClosedAt:        payment.ClosedAt,
		CreatedAt:       payment.CreatedAt,
	}
}

func (h *PaymentHandler) checkoutURL(paymentNo string) string {
	path := checkoutPathPrefix + url.PathEscape(paymentNo)
	if h.publicBaseURL == "" {
		return path
	}
	return h.publicBaseURL + path
}

func writeMockPaymentError(ctx *gin.Context, operation string, err error) {
	status, message := httpStatusForError(err)
	if status == http.StatusInternalServerError {
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
	}
	ctx.JSON(status, httpx.ErrorResponse{Error: message})
}

func httpStatusForError(err error) (int, string) {
	switch {
	case errors.Is(err, model.ErrInvalidPaymentInput):
		return http.StatusBadRequest, model.ErrInvalidPaymentInput.Error()
	case errors.Is(err, model.ErrInvalidPaymentNumber):
		return http.StatusBadRequest, model.ErrInvalidPaymentNumber.Error()
	case errors.Is(err, model.ErrPaymentNotFound):
		return http.StatusNotFound, model.ErrPaymentNotFound.Error()
	case errors.Is(err, model.ErrPaymentConflict):
		return http.StatusConflict, model.ErrPaymentConflict.Error()
	case errors.Is(err, model.ErrPaymentExpired):
		return http.StatusGone, model.ErrPaymentExpired.Error()
	case errors.Is(err, model.ErrPaymentClosed):
		return http.StatusConflict, model.ErrPaymentClosed.Error()
	case errors.Is(err, model.ErrPaymentNotClosable):
		return http.StatusConflict, model.ErrPaymentNotClosable.Error()
	case errors.Is(err, model.ErrNotificationNotFound):
		return http.StatusNotFound, model.ErrNotificationNotFound.Error()
	case errors.Is(err, model.ErrNotificationNotRetryable):
		return http.StatusConflict, model.ErrNotificationNotRetryable.Error()
	default:
		return http.StatusInternalServerError, httpx.InternalServerMessage
	}
}
