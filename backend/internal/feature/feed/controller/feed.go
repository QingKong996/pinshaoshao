package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"pinshaoshao/internal/feature/feed/apperr"
	"pinshaoshao/internal/feature/feed/dto"
	"pinshaoshao/internal/feature/feed/model"
	"pinshaoshao/internal/http/authctx"

	"github.com/gin-gonic/gin"
)

type FeedService interface {
	List(ctx context.Context, userID int, page int, pageSize int) (*model.FeedResult, error)
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
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /feed [get]
func (c *FeedController) List(ctx *gin.Context) {
	var req dto.ListFeedQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
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
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidFeedQuery.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "list feed", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
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
