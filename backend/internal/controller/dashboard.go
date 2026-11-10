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

// DashboardService is the dashboard service surface the controller depends on.
type DashboardService interface {
	Overview(ctx context.Context, shopID int) (*model.DashboardStats, error)
}

type DashboardController struct {
	dashboardService DashboardService
}

func NewDashboardController(dashboardService DashboardService) *DashboardController {
	return &DashboardController{dashboardService: dashboardService}
}

// Overview returns the dashboard of the authenticated shop.
//
// @Summary     Get merchant dashboard overview
// @Description Return the product, SKU, category and inventory counters of the authenticated shop.
// @Tags        Merchant Dashboard
// @Produce     json
// @Success     200 {object} dto.DashboardResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/dashboard [get]
func (c *DashboardController) Overview(ctx *gin.Context) {
	stats, err := c.dashboardService.Overview(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "get dashboard overview", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.DashboardResponse{
		ProductCount:        stats.ProductCount,
		OnShelfProductCount: stats.OnShelfProductCount,
		SKUCount:            stats.SKUCount,
		CategoryCount:       stats.CategoryCount,
		TotalStock:          stats.TotalStock,
		InventoryValueCents: stats.InventoryValueCents,
		OutOfStockSKUCount:  stats.OutOfStockSKUCount,
	})
}
