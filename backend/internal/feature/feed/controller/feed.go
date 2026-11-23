package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"pinshaoshao/internal/feature/feed/apperr"
	"pinshaoshao/internal/feature/feed/dto"
	"pinshaoshao/internal/feature/feed/model"
	"pinshaoshao/internal/http/authctx"
	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
)

type FeedService interface {
	List(ctx context.Context, userID int, page int, pageSize int) (*model.FeedResult, error)
	Get(ctx context.Context, productID int) (*model.ProductDetail, error)
}

type FeedController struct {
	feedService FeedService
}

func NewFeedController(feedService FeedService) *FeedController {
	return &FeedController{feedService: feedService}
}

// List returns a seeded page of available products for the authenticated buyer.
//
// @Summary     List the buyer feed
// @Description List on-shelf products with available stock, personalized by the authenticated buyer ID.
// @Tags        Feed
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       page_size query int false "Page size" default(20)
// @Success     200 {object} dto.FeedListResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /feed [get]
func (c *FeedController) List(ctx *gin.Context) {
	var req dto.ListFeedQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{
			Error: apperr.ErrInvalidFeedQuery.Error(),
		})
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

	result, err := c.feedService.List(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		page,
		pageSize,
	)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidFeedQuery) {
			ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{
				Error: apperr.ErrInvalidFeedQuery.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "list feed", "error", err)
		ctx.JSON(http.StatusInternalServerError, httpx.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	items := make([]dto.FeedItemResponse, 0, len(result.Items))
	for i := range result.Items {
		items = append(items, newFeedItemResponse(&result.Items[i]))
	}
	totalPages := 0
	if result.PageSize > 0 {
		totalPages = int((result.Total + int64(result.PageSize) - 1) / int64(result.PageSize))
	}

	ctx.JSON(http.StatusOK, dto.FeedListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: totalPages,
	})
}

// Get returns an on-shelf product with every SKU for the authenticated buyer.
//
// @Summary     Get a buyer product
// @Description Return an on-shelf product with all of its SKUs, specifications, prices and available stock.
// @Tags        Feed
// @Produce     json
// @Param       productID path int true "Product ID"
// @Success     200 {object} dto.ProductDetailResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /products/{productID} [get]
func (c *FeedController) Get(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("productID"))
	if err != nil || productID <= 0 {
		ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	product, err := c.feedService.Get(ctx.Request.Context(), productID)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidProductID):
			ctx.JSON(http.StatusBadRequest, httpx.ErrorResponse{
				Error: apperr.ErrInvalidProductID.Error(),
			})
		case errors.Is(err, apperr.ErrProductNotFound):
			ctx.JSON(http.StatusNotFound, httpx.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
		default:
			slog.ErrorContext(ctx.Request.Context(), "get product detail", "product_id", productID, "error", err)
			ctx.JSON(http.StatusInternalServerError, httpx.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, newProductDetailResponse(product))
}

func newFeedItemResponse(item *model.FeedItem) dto.FeedItemResponse {
	return dto.FeedItemResponse{
		ID:             item.ID,
		ShopID:         item.ShopID,
		ShopName:       item.ShopName,
		Name:           item.Name,
		Description:    item.Description,
		CategoryName:   item.CategoryName,
		MinPriceCents:  item.MinPriceCents,
		MaxPriceCents:  item.MaxPriceCents,
		AvailableStock: item.AvailableStock,
	}
}

func newProductDetailResponse(product *model.ProductDetail) dto.ProductDetailResponse {
	skus := make([]dto.ProductDetailSKUResponse, 0, len(product.SKUs))
	for i := range product.SKUs {
		sku := &product.SKUs[i]
		skus = append(skus, dto.ProductDetailSKUResponse{
			ID:             sku.ID,
			Specs:          sku.Specs,
			PriceCents:     sku.PriceCents,
			IsEnabled:      sku.IsEnabled,
			AvailableStock: sku.AvailableStock,
		})
	}

	return dto.ProductDetailResponse{
		ID:           product.ID,
		ShopID:       product.ShopID,
		ShopName:     product.ShopName,
		Name:         product.Name,
		Description:  product.Description,
		CategoryName: product.CategoryName,
		SKUs:         skus,
	}
}
