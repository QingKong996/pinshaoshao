package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/dto"
	"pinshaoshao/internal/feature/order/model"
	"pinshaoshao/internal/feature/order/service"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/http/httpx"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderService interface {
	Create(ctx context.Context, userID int, input service.CreateOrderInput) (*model.Order, error)
	List(ctx context.Context, userID int, filter model.OrderListFilter) (*model.OrderListResult, error)
	Get(ctx context.Context, userID int, orderID int) (*model.Order, error)
	Cancel(ctx context.Context, userID int, orderID int) (*model.Order, error)
}

type OrderController struct {
	orderService OrderService
}

// NewOrderController builds the order HTTP controller.
func NewOrderController(orderService OrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

// Create places an order from selected cart items.
//
// @Summary     Create an order
// @Description Place an order from selected cart items. The checkout key makes retries idempotent.
// @Tags        Orders
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateOrderRequest true "Checkout request"
// @Success     201 {object} dto.OrderResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /orders [post]
func (c *OrderController) Create(ctx *gin.Context) {
	var req dto.CreateOrderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCheckout.Error()})
		return
	}

	items := make([]model.CreateOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, model.CreateOrderItem{
			CartItemID: item.CartItemID,
			Quantity:   item.Quantity,
		})
	}

	order, err := c.orderService.Create(ctx.Request.Context(), authctx.MustUserID(ctx), service.CreateOrderInput{
		CheckoutKey:     req.CheckoutKey,
		Items:           items,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ShippingAddress: req.ShippingAddress,
	})
	if err != nil {
		c.writeError(ctx, "create order", err)
		return
	}
	ctx.JSON(http.StatusCreated, newOrderResponse(order))
}

// List returns the authenticated buyer's orders.
//
// @Summary     List orders
// @Description List the authenticated buyer's orders newest first.
// @Tags        Orders
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       page_size query int false "Page size" default(10)
// @Success     200 {object} dto.OrderListResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /orders [get]
func (c *OrderController) List(ctx *gin.Context) {
	var req dto.ListOrdersQuery
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

	result, err := c.orderService.List(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		model.OrderListFilter{Page: page, PageSize: pageSize},
	)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidOrderList) {
			ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderList.Error()})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "list orders", "error", err)
		ctx.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
		return
	}

	items := make([]dto.OrderSummaryResponse, 0, len(result.Orders))
	for i := range result.Orders {
		items = append(items, newOrderSummaryResponse(&result.Orders[i]))
	}
	totalPages := 0
	if result.PageSize > 0 {
		totalPages = int((result.Total + int64(result.PageSize) - 1) / int64(result.PageSize))
	}
	ctx.JSON(http.StatusOK, dto.OrderListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: totalPages,
	})
}

// Get returns one order owned by the authenticated buyer.
//
// @Summary     Get an order
// @Description Get an order and its sub-orders with immutable item snapshots.
// @Tags        Orders
// @Produce     json
// @Param       orderID path int true "Order ID"
// @Success     200 {object} dto.OrderResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /orders/{orderID} [get]
func (c *OrderController) Get(ctx *gin.Context) {
	orderID, err := strconv.Atoi(ctx.Param("orderID"))
	if err != nil || orderID <= 0 {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderID.Error()})
		return
	}

	order, err := c.orderService.Get(ctx.Request.Context(), authctx.MustUserID(ctx), orderID)
	if err != nil {
		c.writeError(ctx, "get order", err)
		return
	}
	ctx.JSON(http.StatusOK, newOrderResponse(order))
}

// Cancel cancels an unpaid order.
//
// @Summary     Cancel an order
// @Description Cancel a pending-payment order and release its frozen stock. Repeated cancellation succeeds.
// @Tags        Orders
// @Produce     json
// @Param       orderID path int true "Order ID"
// @Success     200 {object} dto.OrderResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /orders/{orderID}/cancel [post]
func (c *OrderController) Cancel(ctx *gin.Context) {
	orderID, err := strconv.Atoi(ctx.Param("orderID"))
	if err != nil || orderID <= 0 {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderID.Error()})
		return
	}

	order, err := c.orderService.Cancel(ctx.Request.Context(), authctx.MustUserID(ctx), orderID)
	if err != nil {
		c.writeError(ctx, "cancel order", err)
		return
	}
	ctx.JSON(http.StatusOK, newOrderResponse(order))
}

func (c *OrderController) writeError(ctx *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidCheckout):
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidCheckout.Error()})
	case errors.Is(err, apperr.ErrInvalidOrderID):
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{Error: apperr.ErrInvalidOrderID.Error()})
	case errors.Is(err, apperr.ErrCheckoutConflict):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrCheckoutConflict.Error()})
	case errors.Is(err, apperr.ErrCartItemsChanged):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrCartItemsChanged.Error()})
	case errors.Is(err, apperr.ErrProductUnavailable):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrProductUnavailable.Error()})
	case errors.Is(err, apperr.ErrInsufficientStock):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrInsufficientStock.Error()})
	case errors.Is(err, apperr.ErrOrderNotFound):
		ctx.JSON(http.StatusNotFound, httpx.ErrorResponse{Error: apperr.ErrOrderNotFound.Error()})
	case errors.Is(err, apperr.ErrOrderCannotCancel):
		ctx.JSON(http.StatusConflict, httpx.ErrorResponse{Error: apperr.ErrOrderCannotCancel.Error()})
	default:
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
		ctx.JSON(http.StatusInternalServerError, httpx.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
	}
}

func newOrderResponse(order *model.Order) dto.OrderResponse {
	subOrders := make([]dto.SubOrderResponse, 0, len(order.SubOrders))
	for i := range order.SubOrders {
		subOrder := &order.SubOrders[i]
		items := make([]dto.OrderItemResponse, 0, len(subOrder.Items))
		for j := range subOrder.Items {
			item := &subOrder.Items[j]
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
		subOrders = append(subOrders, dto.SubOrderResponse{
			ID:               subOrder.ID,
			ShopID:           subOrder.ShopID,
			ShopName:         subOrder.ShopName,
			Status:           string(subOrder.Status),
			TotalAmountCents: subOrder.TotalAmountCents,
			Items:            items,
		})
	}

	return dto.OrderResponse{
		ID:               order.ID,
		Status:           string(order.Status),
		TotalAmountCents: order.TotalAmountCents,
		ReceiverName:     order.ReceiverName,
		ReceiverPhone:    order.ReceiverPhone,
		ShippingAddress:  order.ShippingAddress,
		ExpiresAt:        order.ExpiresAt(),
		PaidAt:           order.PaidAt,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
		SubOrders:        subOrders,
	}
}

func newOrderSummaryResponse(order *model.OrderSummary) dto.OrderSummaryResponse {
	return dto.OrderSummaryResponse{
		ID:               order.ID,
		Status:           string(order.Status),
		TotalAmountCents: order.TotalAmountCents,
		ItemCount:        order.ItemCount,
		ShopCount:        order.ShopCount,
		ExpiresAt:        order.ExpiresAt(),
		CreatedAt:        order.CreatedAt,
	}
}
