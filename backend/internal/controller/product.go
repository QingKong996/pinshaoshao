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
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
)

// ProductService is the product service surface the controller depends on.
type ProductService interface {
	List(ctx context.Context, shopID int, filter model.ProductListFilter) (*model.ProductListResult, error)
	Get(ctx context.Context, shopID int, productID int) (*model.ProductSummary, error)
	Create(ctx context.Context, shopID int, input service.CreateProductInput) (*model.ProductSummary, error)
	Update(ctx context.Context, shopID int, productID int, input service.UpdateProductInput) (*model.ProductSummary, error)
	Delete(ctx context.Context, shopID int, productID int) error
}

type ProductController struct {
	productService ProductService
}

func NewProductController(productService ProductService) *ProductController {
	return &ProductController{productService: productService}
}

// List returns a filtered and sorted page of products of the authenticated shop.
//
// @Summary     List merchant products
// @Description List a filtered and sorted page of the authenticated shop's products with their SKU aggregates.
// @Tags        Merchant Products
// @Produce     json
// @Param       page query int false "Page number" default(1)
// @Param       page_size query int false "Page size" default(10)
// @Param       sort query string false "Sort field" Enums(created_at,updated_at,name,category,status,sku_count,total_stock,min_price) default(created_at)
// @Param       order query string false "Sort direction" Enums(asc,desc) default(desc)
// @Param       search query string false "Search product name or description"
// @Param       status query string false "Product status" Enums(on_shelf,off_shelf)
// @Param       category_id query int false "Category ID"
// @Param       uncategorized query bool false "Only uncategorized products"
// @Success     200 {object} dto.ProductListResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products [get]
func (c *ProductController) List(ctx *gin.Context) {
	var req dto.ListProductsQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductListQuery.Error(),
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

	result, err := c.productService.List(
		ctx.Request.Context(),
		middleware.MustCurrentShop(ctx).ID,
		model.ProductListFilter{
			Search:        req.Search,
			Status:        model.ProductStatus(req.Status),
			CategoryID:    req.CategoryID,
			Uncategorized: req.Uncategorized,
			Sort:          model.ProductSort(req.Sort),
			Order:         model.ProductSortOrder(req.Order),
			Page:          page,
			PageSize:      pageSize,
		},
	)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidProductListQuery) {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidProductListQuery.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "list products", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	items := make([]dto.ProductResponse, 0, len(result.Products))
	for i := range result.Products {
		items = append(items, newProductResponse(&result.Products[i]))
	}
	totalPages := 0
	if result.PageSize > 0 {
		totalPages = int((result.Total + int64(result.PageSize) - 1) / int64(result.PageSize))
	}
	ctx.JSON(http.StatusOK, dto.ProductListResponse{
		Items:      items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: totalPages,
	})
}

// Get returns a product of the authenticated shop.
//
// @Summary     Get a merchant product
// @Description Return a product of the authenticated shop with its SKU aggregates.
// @Tags        Merchant Products
// @Produce     json
// @Param       productID path int true "Product ID"
// @Success     200 {object} dto.ProductResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products/{productID} [get]
func (c *ProductController) Get(ctx *gin.Context) {
	productID, ok := pathID(ctx, "productID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	product, err := c.productService.Get(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, productID)
	if err != nil {
		if errors.Is(err, apperr.ErrProductNotFound) {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "get product", "product_id", productID, "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, newProductResponse(product))
}

// Create adds a product to the authenticated shop.
//
// @Summary     Create a merchant product
// @Description Add a product to the authenticated shop.
// @Tags        Merchant Products
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateProductRequest true "Product details"
// @Success     201 {object} dto.ProductResponse
// @Failure     400 {object} dto.ErrorResponse "Invalid payload or a category of another shop"
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products [post]
func (c *ProductController) Create(ctx *gin.Context) {
	var req dto.CreateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	product, err := c.productService.Create(
		ctx.Request.Context(),
		middleware.MustCurrentShop(ctx).ID,
		service.CreateProductInput{
			Name:        req.Name,
			Description: req.Description,
			CategoryID:  req.CategoryID,
			Status:      model.ProductStatus(req.Status),
		},
	)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidCategoryID) {
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidCategoryID.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "create product", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, newProductResponse(product))
}

// Update replaces the editable fields of a product of the authenticated shop.
//
// @Summary     Update a merchant product
// @Description Update the name, description, category and status of a product of the authenticated shop.
// @Tags        Merchant Products
// @Accept      json
// @Produce     json
// @Param       productID path int true "Product ID"
// @Param       request body dto.UpdateProductRequest true "Product details"
// @Success     200 {object} dto.ProductResponse
// @Failure     400 {object} dto.ErrorResponse "Invalid payload or a category of another shop"
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products/{productID} [patch]
func (c *ProductController) Update(ctx *gin.Context) {
	productID, ok := pathID(ctx, "productID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	var req dto.UpdateProductRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	product, err := c.productService.Update(
		ctx.Request.Context(),
		middleware.MustCurrentShop(ctx).ID,
		productID,
		service.UpdateProductInput{
			Name:        req.Name,
			Description: req.Description,
			CategoryID:  req.CategoryID,
			Status:      model.ProductStatus(req.Status),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrProductNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrInvalidCategoryID):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidCategoryID.Error(),
			})
		default:
			slog.ErrorContext(ctx.Request.Context(), "update product", "product_id", productID, "error", err)
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, newProductResponse(product))
}

// Delete removes a product of the authenticated shop.
//
// @Summary     Delete a merchant product
// @Description Delete a product of the authenticated shop; its SKUs are removed with it.
// @Tags        Merchant Products
// @Param       productID path int true "Product ID"
// @Success     204
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products/{productID} [delete]
func (c *ProductController) Delete(ctx *gin.Context) {
	productID, ok := pathID(ctx, "productID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	err := c.productService.Delete(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, productID)
	if err != nil {
		if errors.Is(err, apperr.ErrProductNotFound) {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "delete product", "product_id", productID, "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func newProductResponse(product *model.ProductSummary) dto.ProductResponse {
	return dto.ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Status:        string(product.Status),
		CategoryID:    product.CategoryID,
		CategoryName:  product.CategoryName,
		SKUCount:      product.SKUCount,
		TotalStock:    product.TotalStock,
		MinPriceCents: product.MinPriceCents,
		MaxPriceCents: product.MaxPriceCents,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}
}
