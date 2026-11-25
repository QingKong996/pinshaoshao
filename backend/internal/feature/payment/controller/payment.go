package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/dto"
	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/feature/payment/service"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

type PaymentService interface {
	Create(ctx context.Context, userID int, orderID int) (*model.Payment, bool, error)
	Get(ctx context.Context, userID int, paymentNo string) (*model.Payment, error)
	Pay(ctx context.Context, userID int, paymentNo string) (*service.PaymentSession, error)
	Close(ctx context.Context, userID int, paymentNo string) (*model.Payment, error)
}

type PaymentController struct {
	paymentService PaymentService
}

func NewPaymentController(paymentService PaymentService) *PaymentController {
	return &PaymentController{paymentService: paymentService}
}

// Create creates the single business payment for an order.
//
// @Summary     Create a payment
// @Description Create or return the single pending payment for an owned order. Repeated calls are idempotent.
// @Tags        Payments
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateOrderPaymentRequest true "Payment request"
// @Success     201 {object} dto.OrderPaymentResponse
// @Success     200 {object} dto.OrderPaymentResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /payments [post]
func (c *PaymentController) Create(ctx *gin.Context) {
	var request dto.CreateOrderPaymentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderID.Error()})
		return
	}

	payment, created, err := c.paymentService.Create(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		request.OrderID,
	)
	if err != nil {
		writePaymentError(ctx, "create payment", err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	ctx.JSON(status, newOrderPaymentResponse(payment, ""))
}

// Get returns an owned payment.
//
// @Summary     Get a payment
// @Description Get the status of an owned payment.
// @Tags        Payments
// @Produce     json
// @Param       paymentNo path string true "Payment number"
// @Success     200 {object} dto.OrderPaymentResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /payments/{paymentNo} [get]
func (c *PaymentController) Get(ctx *gin.Context) {
	payment, err := c.paymentService.Get(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		ctx.Param("paymentNo"),
	)
	if err != nil {
		writePaymentError(ctx, "get payment", err)
		return
	}
	ctx.JSON(http.StatusOK, newOrderPaymentResponse(payment, ""))
}

// Pay creates or reuses the provider checkout session.
//
// @Summary     Pay a payment
// @Description Open or reuse the provider checkout session for an owned pending payment.
// @Tags        Payments
// @Produce     json
// @Param       paymentNo path string true "Payment number"
// @Success     200 {object} dto.OrderPaymentResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     502 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /payments/{paymentNo}/pay [post]
func (c *PaymentController) Pay(ctx *gin.Context) {
	session, err := c.paymentService.Pay(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		ctx.Param("paymentNo"),
	)
	if err != nil {
		writePaymentError(ctx, "pay payment", err)
		return
	}
	ctx.JSON(http.StatusOK, newOrderPaymentResponse(session.Payment, session.CheckoutURL))
}

// Close closes an owned pending payment and cancels its order.
//
// @Summary     Close a payment
// @Description Close an owned pending payment, cancel its pending order, and release reserved inventory.
// @Tags        Payments
// @Produce     json
// @Param       paymentNo path string true "Payment number"
// @Success     200 {object} dto.OrderPaymentResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     502 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /payments/{paymentNo}/close [post]
func (c *PaymentController) Close(ctx *gin.Context) {
	payment, err := c.paymentService.Close(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		ctx.Param("paymentNo"),
	)
	if err != nil {
		writePaymentError(ctx, "close payment", err)
		return
	}
	ctx.JSON(http.StatusOK, newOrderPaymentResponse(payment, ""))
}

func newOrderPaymentResponse(payment *model.Payment, checkoutURL string) dto.OrderPaymentResponse {
	return dto.OrderPaymentResponse{
		PaymentNo:             payment.PaymentNo,
		OrderID:               payment.OrderID,
		Provider:              payment.Provider,
		ProviderTransactionID: payment.ProviderTransactionID,
		Status:                string(payment.Status),
		AmountCents:           payment.AmountCents,
		CheckoutURL:           checkoutURL,
		PaidAt:                payment.PaidAt,
		CreatedAt:             payment.CreatedAt,
		UpdatedAt:             payment.UpdatedAt,
	}
}

func writePaymentError(ctx *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidOrderID):
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderID.Error()})
	case errors.Is(err, apperr.ErrInvalidPaymentNo):
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidPaymentNo.Error()})
	case errors.Is(err, apperr.ErrInvalidCallback):
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCallback.Error()})
	case errors.Is(err, apperr.ErrInvalidWebhookSignature),
		errors.Is(err, apperr.ErrStaleWebhookSignature):
		ctx.JSON(http.StatusUnauthorized, httpx.ErrorResponse{Error: apperr.ErrInvalidWebhookSignature.Error()})
	case errors.Is(err, apperr.ErrPaymentNotFound):
		ctx.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: apperr.ErrPaymentNotFound.Error()})
	case errors.Is(err, apperr.ErrOrderNotFound):
		ctx.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: apperr.ErrOrderNotFound.Error()})
	case errors.Is(err, apperr.ErrOrderNotPayable):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrOrderNotPayable.Error()})
	case errors.Is(err, apperr.ErrPaymentClosed):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrPaymentClosed.Error()})
	case errors.Is(err, apperr.ErrPaymentConflict):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrPaymentConflict.Error()})
	case errors.Is(err, apperr.ErrPaymentProviderUnavailable),
		errors.Is(err, apperr.ErrPaymentProviderInvalid):
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
		ctx.JSON(http.StatusBadGateway, httpx.ErrorResponse{Error: apperr.ErrPaymentProviderUnavailable.Error()})
	default:
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
		ctx.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
	}
}
