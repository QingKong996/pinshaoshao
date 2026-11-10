package controller

import (
	"context"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
)

// InventoryService is the inventory service surface the controller depends on.
type InventoryService interface {
	List(ctx context.Context, shopID int) ([]model.InventoryItem, error)
}

type InventoryController struct {
	inventoryService InventoryService
}

func NewInventoryController(inventoryService InventoryService) *InventoryController {
	return &InventoryController{inventoryService: inventoryService}
}

// List returns the inventory of the authenticated shop.
//
// @Summary     List merchant inventory
// @Description List every SKU of the authenticated shop ordered by stock; disabled SKUs are flagged.
// @Tags        Merchant Inventory
// @Produce     json
// @Success     200 {array} dto.InventoryItemResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/inventory [get]
func (c *InventoryController) List(ctx *gin.Context) {
	items, err := c.inventoryService.List(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "list inventory", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	response := make([]dto.InventoryItemResponse, 0, len(items))
	for i := range items {
		response = append(response, newInventoryItemResponse(&items[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

func newInventoryItemResponse(item *model.InventoryItem) dto.InventoryItemResponse {
	return dto.InventoryItemResponse{
		SKUID:        item.SKUID,
		ProductID:    item.ProductID,
		ProductName:  item.ProductName,
		CategoryID:   item.CategoryID,
		CategoryName: item.CategoryName,
		Specs:        item.Specs,
		PriceCents:   item.PriceCents,
		Stock:        item.Stock,
		IsEnabled:    item.IsEnabled,
		UpdatedAt:    item.UpdatedAt,
	}
}
