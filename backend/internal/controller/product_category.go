package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
)

// ProductCategoryService is the category service surface the controller depends on.
type ProductCategoryService interface {
	List(ctx context.Context, shopID int) ([]model.ProductCategory, error)
	Create(ctx context.Context, shopID int, req *dto.CreateProductCategoryRequest) (*model.ProductCategory, error)
	Update(ctx context.Context, shopID int, categoryID int, req *dto.UpdateProductCategoryRequest) (*model.ProductCategory, error)
	Delete(ctx context.Context, shopID int, categoryID int) error
}

type ProductCategoryController struct {
	categoryService ProductCategoryService
}

func NewProductCategoryController(categoryService ProductCategoryService) *ProductCategoryController {
	return &ProductCategoryController{categoryService: categoryService}
}

// List returns the categories of the authenticated shop.
//
// @Summary     List merchant categories
// @Description List the categories of the authenticated shop with their product counts.
// @Tags        Merchant Categories
// @Produce     json
// @Success     200 {array} dto.ProductCategoryResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/categories [get]
func (c *ProductCategoryController) List(ctx *gin.Context) {
	shop := middleware.MustCurrentShop(ctx)

	categories, err := c.categoryService.List(ctx.Request.Context(), shop.ID)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "list product categories", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	response := make([]dto.ProductCategoryResponse, 0, len(categories))
	for i := range categories {
		response = append(response, newProductCategoryResponse(&categories[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

// Create adds a category to the authenticated shop.
//
// @Summary     Create a product category
// @Description Add a category to the authenticated shop.
// @Tags        Merchant Categories
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateProductCategoryRequest true "Category details"
// @Success     201 {object} dto.ProductCategoryResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse "Category name already exists"
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/categories [post]
func (c *ProductCategoryController) Create(ctx *gin.Context) {
	var req dto.CreateProductCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	category, err := c.categoryService.Create(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, &req)
	if err != nil {
		if errors.Is(err, apperr.ErrCategoryNameTaken) {
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrCategoryNameTaken.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "create product category", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, newProductCategoryResponse(category))
}

// Update renames a category of the authenticated shop.
//
// @Summary     Update a product category
// @Description Rename a category of the authenticated shop.
// @Tags        Merchant Categories
// @Accept      json
// @Produce     json
// @Param       categoryID path int true "Category ID"
// @Param       request body dto.UpdateProductCategoryRequest true "Category details"
// @Success     200 {object} dto.ProductCategoryResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse "Category name already exists"
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/categories/{categoryID} [patch]
func (c *ProductCategoryController) Update(ctx *gin.Context) {
	categoryID, ok := pathID(ctx, "categoryID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidCategoryID.Error(),
		})
		return
	}

	var req dto.UpdateProductCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	category, err := c.categoryService.Update(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, categoryID, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrCategoryNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrCategoryNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrCategoryNameTaken):
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrCategoryNameTaken.Error(),
			})
		default:
			slog.ErrorContext(ctx.Request.Context(), "update product category", "category_id", categoryID, "error", err)
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, newProductCategoryResponse(category))
}

// Delete removes a category of the authenticated shop.
//
// @Summary     Delete a product category
// @Description Delete a category of the authenticated shop; its products become uncategorized.
// @Tags        Merchant Categories
// @Param       categoryID path int true "Category ID"
// @Success     204
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/categories/{categoryID} [delete]
func (c *ProductCategoryController) Delete(ctx *gin.Context) {
	categoryID, ok := pathID(ctx, "categoryID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidCategoryID.Error(),
		})
		return
	}

	err := c.categoryService.Delete(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, categoryID)
	if err != nil {
		if errors.Is(err, apperr.ErrCategoryNotFound) {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrCategoryNotFound.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "delete product category", "category_id", categoryID, "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func newProductCategoryResponse(category *model.ProductCategory) dto.ProductCategoryResponse {
	return dto.ProductCategoryResponse{
		ID:           category.ID,
		Name:         category.Name,
		ProductCount: category.ProductCount,
		CreatedAt:    category.CreatedAt,
		UpdatedAt:    category.UpdatedAt,
	}
}
