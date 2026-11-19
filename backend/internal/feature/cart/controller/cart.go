package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/feature/cart/apperr"
	"pinshaoshao/internal/feature/cart/dto"
	"pinshaoshao/internal/feature/cart/model"
	"pinshaoshao/internal/feature/shared/authctx"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CartService interface {
	List(ctx context.Context, userID int) ([]model.CartItem, error)
	Add(ctx context.Context, userID int, skuID int, quantity int) (*model.CartItem, error)
	UpdateQuantity(ctx context.Context, userID int, itemID int, quantity int) (*model.CartItem, error)
	Delete(ctx context.Context, userID int, itemID int) error
}

type CartController struct {
	cartService CartService
}

func NewCartController(cartService CartService) *CartController {
	return &CartController{cartService: cartService}
}

// List returns the authenticated buyer's cart.
//
// @Summary     List cart items
// @Description List the authenticated buyer's cart with snapshot data and current availability.
// @Tags        Cart
// @Produce     json
// @Success     200 {array} dto.CartItemResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /cart [get]
func (c *CartController) List(ctx *gin.Context) {
	items, err := c.cartService.List(ctx.Request.Context(), authctx.MustUserID(ctx))
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "list cart", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
		return
	}

	response := make([]dto.CartItemResponse, 0, len(items))
	for i := range items {
		response = append(response, newCartItemResponse(&items[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

// Add puts a SKU into the authenticated buyer's cart.
//
// @Summary     Add a cart item
// @Description Add an on-shelf, enabled SKU with available stock. Repeated adds accumulate quantity.
// @Tags        Cart
// @Accept      json
// @Produce     json
// @Param       request body dto.AddCartItemRequest true "SKU and quantity"
// @Success     201 {object} dto.CartItemResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /cart/items [post]
func (c *CartController) Add(ctx *gin.Context) {
	var req dto.AddCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidCartItem.Error()})
		return
	}

	item, err := c.cartService.Add(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		req.SKUID,
		req.Quantity,
	)
	if err != nil {
		c.writeError(ctx, "add cart item", err)
		return
	}
	ctx.JSON(http.StatusCreated, newCartItemResponse(item))
}

// UpdateQuantity sets the quantity of a cart item.
//
// @Summary     Update a cart item quantity
// @Description Set an absolute quantity for a cart item owned by the authenticated buyer.
// @Tags        Cart
// @Accept      json
// @Produce     json
// @Param       itemID path int true "Cart item ID"
// @Param       request body dto.UpdateCartItemRequest true "Quantity"
// @Success     200 {object} dto.CartItemResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /cart/items/{itemID} [patch]
func (c *CartController) UpdateQuantity(ctx *gin.Context) {
	itemID, err := strconv.Atoi(ctx.Param("itemID"))
	if err != nil || itemID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidCartItemID.Error()})
		return
	}

	var req dto.UpdateCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidCartQuantity.Error()})
		return
	}

	item, err := c.cartService.UpdateQuantity(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		itemID,
		req.Quantity,
	)
	if err != nil {
		c.writeError(ctx, "update cart item", err)
		return
	}
	ctx.JSON(http.StatusOK, newCartItemResponse(item))
}

// Delete removes a cart item.
//
// @Summary     Delete a cart item
// @Description Remove a cart item owned by the authenticated buyer.
// @Tags        Cart
// @Param       itemID path int true "Cart item ID"
// @Success     204
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /cart/items/{itemID} [delete]
func (c *CartController) Delete(ctx *gin.Context) {
	itemID, err := strconv.Atoi(ctx.Param("itemID"))
	if err != nil || itemID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidCartItemID.Error()})
		return
	}

	if err := c.cartService.Delete(ctx.Request.Context(), authctx.MustUserID(ctx), itemID); err != nil {
		c.writeError(ctx, "delete cart item", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *CartController) writeError(ctx *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidSKUID):
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidSKUID.Error()})
	case errors.Is(err, apperr.ErrInvalidCartItemID):
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidCartItemID.Error()})
	case errors.Is(err, apperr.ErrInvalidCartQuantity):
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidCartQuantity.Error()})
	case errors.Is(err, apperr.ErrInvalidCartItem):
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidCartItem.Error()})
	case errors.Is(err, apperr.ErrProductNotFound):
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: apperr.ErrProductNotFound.Error()})
	case errors.Is(err, apperr.ErrCartItemNotFound):
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: apperr.ErrCartItemNotFound.Error()})
	case errors.Is(err, apperr.ErrProductUnavailable):
		ctx.JSON(http.StatusConflict, dto.ErrorResponse{Error: apperr.ErrProductUnavailable.Error()})
	default:
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
	}
}

func newCartItemResponse(item *model.CartItem) dto.CartItemResponse {
	return dto.CartItemResponse{
		ID:             item.ID,
		ProductID:      item.ProductID,
		SKUID:          item.SKUID,
		ProductName:    item.ProductName,
		SKUSpecs:       item.Specs,
		UnitPriceCents: item.UnitPriceCents,
		Quantity:       item.Quantity,
		LineTotalCents: item.LineTotalCents(),
		AvailableStock: item.AvailableStock,
		Availability:   string(item.Availability()),
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}
