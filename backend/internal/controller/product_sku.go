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
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProductSKUService is the SKU service surface the controller depends on.
type ProductSKUService interface {
	Create(ctx context.Context, shopID int, productID int, req *dto.CreateProductSKURequest) (*model.ProductSKU, error)
	List(ctx context.Context, productID int) ([]model.ProductSKU, error)
}

type ProductSKUController struct {
	skuService ProductSKUService
}

func NewProductSKUController(skuService ProductSKUService) *ProductSKUController {
	return &ProductSKUController{skuService: skuService}
}

// Create adds a purchasable SKU to a product.
//
// @Summary     Create a product SKU
// @Description Add a purchasable SKU to a product. Only the shop that owns the product may add SKUs, and all SKUs of a product must share the same spec keys.
// @Tags        Product SKUs
// @Accept      json
// @Produce     json
// @Param       productID path int true "Product ID"
// @Param       request body dto.CreateProductSKURequest true "SKU details"
// @Success     201 {object} dto.ProductSKUResponse
// @Failure     400 {object} dto.ErrorResponse "Invalid payload or spec keys that differ from the product's existing SKUs"
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse "Product belongs to another shop"
// @Failure     404 {object} dto.ErrorResponse "Product not found"
// @Failure     409 {object} dto.ErrorResponse "Product already has a SKU with the same specs"
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /products/{productID}/skus [post]
func (c *ProductSKUController) Create(ctx *gin.Context) {
	productID, ok := pathProductID(ctx)
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	var req dto.CreateProductSKURequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	sku, err := c.skuService.Create(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, productID, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrProductNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrProductAccessDenied):
			ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
				Error: apperr.ErrProductAccessDenied.Error(),
			})
		case errors.Is(err, apperr.ErrProductSKUSpecsMismatch):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrProductSKUSpecsMismatch.Error(),
			})
		case errors.Is(err, apperr.ErrProductSKUAlreadyExists):
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrProductSKUAlreadyExists.Error(),
			})
		default:
			slog.ErrorContext(ctx.Request.Context(), "create product SKU", "product_id", productID, "error", err)
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusCreated, newProductSKUResponse(sku))
}

// List returns the SKUs of a product.
//
// @Summary     List product SKUs
// @Description List the purchasable SKUs of a product.
// @Tags        Product SKUs
// @Produce     json
// @Param       productID path int true "Product ID"
// @Success     200 {array} dto.ProductSKUResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /products/{productID}/skus [get]
func (c *ProductSKUController) List(ctx *gin.Context) {
	productID, ok := pathProductID(ctx)
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	skus, err := c.skuService.List(ctx.Request.Context(), productID)
	if err != nil {
		if errors.Is(err, apperr.ErrProductNotFound) {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "list product SKUs", "product_id", productID, "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	response := make([]dto.ProductSKUResponse, 0, len(skus))
	for i := range skus {
		response = append(response, newProductSKUResponse(&skus[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

func pathProductID(ctx *gin.Context) (int, bool) {
	productID, err := strconv.Atoi(ctx.Param("productID"))
	if err != nil || productID <= 0 {
		return 0, false
	}
	return productID, true
}

func newProductSKUResponse(sku *model.ProductSKU) dto.ProductSKUResponse {
	return dto.ProductSKUResponse{
		ID:         sku.ID,
		ProductID:  sku.ProductID,
		Specs:      sku.Specs,
		PriceCents: sku.PriceCents,
		Stock:      sku.Stock,
		IsEnabled:  sku.IsEnabled,
		CreatedAt:  sku.CreatedAt,
		UpdatedAt:  sku.UpdatedAt,
	}
}
