package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/feature/inventory/apperr"
	"pinshaoshao/internal/feature/inventory/dto"
	"pinshaoshao/internal/feature/inventory/model"
	"pinshaoshao/internal/feature/inventory/service"
	"pinshaoshao/internal/feature/shared/authctx"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InventoryService is the inventory service surface the controller depends on.
type InventoryService interface {
	List(ctx context.Context, shopID int) ([]model.InventoryItem, error)
	Update(ctx context.Context, shopID int, skuID int, input service.UpdateInventoryInput) (*model.InventoryItem, error)
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
// @Description List every inventory record of the authenticated shop ordered by available stock; disabled SKUs are flagged.
// @Tags        Merchant Inventory
// @Produce     json
// @Success     200 {array} dto.InventoryItemResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/inventory [get]
func (c *InventoryController) List(ctx *gin.Context) {
	items, err := c.inventoryService.List(ctx.Request.Context(), authctx.MustShopID(ctx))
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

// Update sets the total stock of an inventory record.
//
// @Summary     Update merchant inventory
// @Description Set the total stock of an inventory record. The supplied version must match the current inventory version.
// @Tags        Merchant Inventory
// @Accept      json
// @Produce     json
// @Param       skuID path int true "SKU ID"
// @Param       request body dto.UpdateInventoryRequest true "Inventory total and version"
// @Success     200 {object} dto.InventoryItemResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse "Version conflict or total stock below frozen stock"
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/inventory/{skuID} [patch]
func (c *InventoryController) Update(ctx *gin.Context) {
	skuID, err := strconv.Atoi(ctx.Param("skuID"))
	if err != nil || skuID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidSKUID.Error(),
		})
		return
	}

	var req dto.UpdateInventoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidInventoryUpdate.Error(),
		})
		return
	}

	item, err := c.inventoryService.Update(ctx.Request.Context(), authctx.MustShopID(ctx), skuID, service.UpdateInventoryInput{
		TotalStock: *req.TotalStock,
		Version:    *req.Version,
	})
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidInventoryUpdate):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidInventoryUpdate.Error(),
			})
		case errors.Is(err, apperr.ErrInventoryItemNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrInventoryItemNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrInventoryVersion):
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrInventoryVersion.Error(),
			})
		case errors.Is(err, apperr.ErrInventoryBelowFrozen):
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrInventoryBelowFrozen.Error(),
			})
		default:
			slog.ErrorContext(ctx.Request.Context(), "update inventory", "sku_id", skuID, "error", err)
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, newInventoryItemResponse(item))
}

func newInventoryItemResponse(item *model.InventoryItem) dto.InventoryItemResponse {
	return dto.InventoryItemResponse{
		SKUID:          item.SKUID,
		ProductID:      item.ProductID,
		ProductName:    item.ProductName,
		CategoryID:     item.CategoryID,
		CategoryName:   item.CategoryName,
		Specs:          item.Specs,
		PriceCents:     item.PriceCents,
		TotalStock:     item.TotalStock,
		FrozenStock:    item.FrozenStock,
		AvailableStock: item.AvailableStock(),
		Version:        item.Version,
		IsEnabled:      item.IsEnabled,
		UpdatedAt:      item.UpdatedAt,
	}
}
