package controller

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/dto"
	"pinshaoshao/internal/feature/order/model"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

// ListMerchant returns the authenticated shop's sub-orders.
//
// @Summary     List merchant orders
// @Description List the authenticated shop's sub-orders newest first.
// @Tags        Merchant Orders
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       page_size query int false "Page size" default(10)
// @Param       status query string false "Sub-order status" Enums(pending_payment,awaiting_shipment,shipped,completed,cancelled)
// @Success     200 {object} dto.MerchantOrderListResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/orders [get]
func (c *OrderController) ListMerchant(ctx *gin.Context) {
	var req dto.ListMerchantOrdersQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderList.Error()})
		return
	}

	page := 0
	if req.Page != nil {
		page = *req.Page
	}
	pageSize := 0
	if req.PageSize != nil {
		pageSize = *req.PageSize
	}

	result, err := c.orderService.ListMerchant(
		ctx.Request.Context(),
		authctx.MustShopID(ctx),
		model.MerchantOrderListFilter{
			Status:   model.SubOrderStatus(req.Status),
			Page:     page,
			PageSize: pageSize,
		},
	)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidOrderList) {
			ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderList.Error()})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "list merchant orders", "error", err)
		ctx.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
		return
	}

	items := make([]dto.MerchantOrderSummaryResponse, 0, len(result.Orders))
	for i := range result.Orders {
		items = append(items, newMerchantOrderSummaryResponse(&result.Orders[i]))
	}
	totalPages := 0
	if result.PageSize > 0 {
		totalPages = int((result.Total + int64(result.PageSize) - 1) / int64(result.PageSize))
	}
	ctx.JSON(http.StatusOK, dto.MerchantOrderListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: totalPages,
	})
}

// GetMerchant returns one sub-order owned by the authenticated shop.
//
// @Summary     Get a merchant order
// @Description Get one sub-order owned by the authenticated shop with immutable item snapshots.
// @Tags        Merchant Orders
// @Produce     json
// @Param       subOrderID path int true "Sub-order ID"
// @Success     200 {object} dto.MerchantOrderDetailResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/orders/{subOrderID} [get]
func (c *OrderController) GetMerchant(ctx *gin.Context) {
	subOrderID, err := strconv.Atoi(ctx.Param("subOrderID"))
	if err != nil || subOrderID <= 0 {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderID.Error()})
		return
	}

	order, err := c.orderService.GetMerchant(
		ctx.Request.Context(),
		authctx.MustShopID(ctx),
		subOrderID,
	)
	if err != nil {
		c.writeError(ctx, "get merchant order", err)
		return
	}
	ctx.JSON(http.StatusOK, newMerchantOrderDetailResponse(order))
}

func newMerchantOrderSummaryResponse(order *model.MerchantOrderSummary) dto.MerchantOrderSummaryResponse {
	return dto.MerchantOrderSummaryResponse{
		ID:               order.ID,
		OrderID:          order.OrderID,
		Status:           string(order.Status),
		TotalAmountCents: order.TotalAmountCents,
		ItemCount:        order.ItemCount,
		ReceiverName:     order.ReceiverName,
		ReceiverPhone:    order.ReceiverPhone,
		ShippingAddress:  order.ShippingAddress,
		PaidAt:           order.PaidAt,
		CreatedAt:        order.CreatedAt,
	}
}

func newMerchantOrderDetailResponse(order *model.MerchantOrderDetail) dto.MerchantOrderDetailResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.Items))
	for i := range order.Items {
		item := &order.Items[i]
		items = append(items, dto.OrderItemResponse{
			ID:              item.ID,
			ProductID:       item.ProductID,
			SKUID:           item.SKUID,
			ProductName:     item.ProductName,
			SKUSpecs:        item.Specs,
			UnitPriceCents:  item.UnitPriceCents,
			Quantity:        item.Quantity,
			LineAmountCents: item.LineAmountCents,
		})
	}
	return dto.MerchantOrderDetailResponse{
		ID:               order.ID,
		OrderID:          order.OrderID,
		Status:           string(order.Status),
		TotalAmountCents: order.TotalAmountCents,
		ItemCount:        order.ItemCount,
		ReceiverName:     order.ReceiverName,
		ReceiverPhone:    order.ReceiverPhone,
		ShippingAddress:  order.ShippingAddress,
		PaidAt:           order.PaidAt,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
		Items:            items,
	}
}
