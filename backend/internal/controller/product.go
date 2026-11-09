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

// ProductService is the product service surface the controller depends on.
type ProductService interface {
	List(ctx context.Context, shopID int) ([]model.ProductSummary, error)
	Get(ctx context.Context, shopID int, productID int) (*model.ProductSummary, error)
	Create(ctx context.Context, shopID int, req *dto.CreateProductRequest) (*model.ProductSummary, error)
	Update(ctx context.Context, shopID int, productID int, req *dto.UpdateProductRequest) (*model.ProductSummary, error)
	Delete(ctx context.Context, shopID int, productID int) error
}

type ProductController struct {
	productService ProductService
}

func NewProductController(productService ProductService) *ProductController {
	return &ProductController{productService: productService}
}

// List returns the products of the authenticated shop.
//
// @Summary     List merchant products
// @Description List the products of the authenticated shop with their SKU aggregates, newest first.
// @Tags        Merchant Products
// @Produce     json
// @Success     200 {array} dto.ProductResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products [get]
func (c *ProductController) List(ctx *gin.Context) {
	products, err := c.productService.List(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID)
	if err != nil {
		slog.ErrorContext(ctx.Request.Context(), "list products", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	response := make([]dto.ProductResponse, 0, len(products))
	for i := range products {
		response = append(response, newProductResponse(&products[i]))
	}
	ctx.JSON(http.StatusOK, response)
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

	product, err := c.productService.Create(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, &req)
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

	product, err := c.productService.Update(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, productID, &req)
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
