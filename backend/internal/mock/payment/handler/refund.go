package handler

import (
	"context"
	"net/http"

	"pinshaoshao/internal/http/httpx"
	"pinshaoshao/internal/mock/payment/dto"
	"pinshaoshao/internal/mock/payment/model"

	"github.com/gin-gonic/gin"
)

type RefundService interface {
	Refund(ctx context.Context, paymentNo string, input model.RefundInput) (*model.Refund, error)
}

type RefundHandler struct {
	service RefundService
}

func NewRefundHandler(refundService RefundService) *RefundHandler {
	return &RefundHandler{service: refundService}
}

// Refund creates an idempotent full refund for a paid mock payment.
//
// @Summary     Refund a mock payment
// @Description Refund the full amount of a paid mock payment. Repeated calls with the same refund number are idempotent.
// @Tags        MockPay
// @Accept      json
// @Produce     json
// @Param       paymentNo path string true "Provider payment number"
// @Param       request body dto.RefundRequest true "Refund request"
// @Success     200 {object} dto.RefundResponse
// @Success     201 {object} dto.RefundResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    MockPayBearerAuth
// @Router      /mockpay/api/v1/payments/{paymentNo}/refunds [post]
func (h *RefundHandler) Refund(ctx *gin.Context) {
	var request dto.RefundRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: model.ErrInvalidRefundInput.Error()})
		return
	}

	refund, err := h.service.Refund(
		ctx.Request.Context(),
		ctx.Param("paymentNo"),
		model.RefundInput{
			RefundNo:    request.RefundNo,
			AmountCents: request.AmountCents,
			Reason:      request.Reason,
		},
	)
	if err != nil {
		writeMockPaymentError(ctx, "refund mock payment", err)
		return
	}
	ctx.JSON(http.StatusOK, dto.RefundResponse{
		RefundNo:         refund.RefundNo,
		ProviderRefundNo: refund.ProviderRefundNo,
		PaymentNo:        refund.PaymentNo,
		AmountCents:      refund.AmountCents,
		Reason:           refund.Reason,
		Status:           string(refund.Status),
		RefundedAt:       refund.RefundedAt,
	})
}
