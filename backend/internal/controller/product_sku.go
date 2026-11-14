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

// ProductSKUService is the SKU service surface the controller depends on.
type ProductSKUService interface {
	Create(ctx context.Context, shopID int, productID int, input service.CreateProductSKUInput) (*model.ProductSKU, error)
	List(ctx context.Context, shopID int, productID int) ([]model.ProductSKU, error)
	Update(ctx context.Context, shopID int, productID int, skuID int, update *model.ProductSKUUpdate) (*model.ProductSKU, error)
	Delete(ctx context.Context, shopID int, productID int, skuID int) error
}

type ProductSKUController struct {
	skuService ProductSKUService
}

func NewProductSKUController(skuService ProductSKUService) *ProductSKUController {
	return &ProductSKUController{skuService: skuService}
}

// Create adds a purchasable SKU to a product of the authenticated shop.
//
// @Summary     Create a product SKU
// @Description Add a purchasable SKU to a product of the authenticated shop. All SKUs of a product must share the same spec keys.
// @Tags        Product SKUs
// @Accept      json
// @Produce     json
// @Param       productID path int true "Product ID"
// @Param       request body dto.CreateProductSKURequest true "SKU details"
// @Success     201 {object} dto.ProductSKUResponse
// @Failure     400 {object} dto.ErrorResponse "Invalid payload or spec keys that differ from the product's existing SKUs"
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse "Product not found"
// @Failure     409 {object} dto.ErrorResponse "Product already has a SKU with the same specs"
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products/{productID}/skus [post]
func (c *ProductSKUController) Create(ctx *gin.Context) {
	productID, ok := pathID(ctx, "productID")
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

	sku, err := c.skuService.Create(
		ctx.Request.Context(),
		middleware.MustCurrentShop(ctx).ID,
		productID,
		service.CreateProductSKUInput{
			Specs:      req.Specs,
			PriceCents: req.PriceCents,
			Stock:      req.Stock,
			IsEnabled:  req.IsEnabled,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrProductNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrProductSKUSpecsMismatch):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrProductSKUSpecsMismatch.Error(),
			})
		case errors.Is(err, apperr.ErrInvalidProductSKUSpecs):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidProductSKUSpecs.Error(),
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

// List returns the SKUs of a product of the authenticated shop.
//
// @Summary     List product SKUs
// @Description List the purchasable SKUs of a product of the authenticated shop.
// @Tags        Product SKUs
// @Produce     json
// @Param       productID path int true "Product ID"
// @Success     200 {array} dto.ProductSKUResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products/{productID}/skus [get]
func (c *ProductSKUController) List(ctx *gin.Context) {
	productID, ok := pathID(ctx, "productID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	skus, err := c.skuService.List(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, productID)
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

// Update partially updates a SKU of a product of the authenticated shop.
//
// @Summary     Update a product SKU
// @Description Partially update a SKU of a product of the authenticated shop; at least one field is required.
// @Tags        Product SKUs
// @Accept      json
// @Produce     json
// @Param       productID path int true "Product ID"
// @Param       skuID path int true "SKU ID"
// @Param       request body dto.UpdateProductSKURequest true "SKU fields"
// @Success     200 {object} dto.ProductSKUResponse
// @Failure     400 {object} dto.ErrorResponse "Invalid payload, no fields, or spec keys that differ from the product's existing SKUs"
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse "Product already has a SKU with the same specs"
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products/{productID}/skus/{skuID} [patch]
func (c *ProductSKUController) Update(ctx *gin.Context) {
	productID, ok := pathID(ctx, "productID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	skuID, ok := pathID(ctx, "skuID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductSKUID.Error(),
		})
		return
	}

	var req dto.UpdateProductSKURequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	sku, err := c.skuService.Update(
		ctx.Request.Context(),
		middleware.MustCurrentShop(ctx).ID,
		productID,
		skuID,
		&model.ProductSKUUpdate{
			Specs:      req.Specs,
			PriceCents: req.PriceCents,
			Stock:      req.Stock,
			IsEnabled:  req.IsEnabled,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrProductNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrProductSKUNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductSKUNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrInvalidProductSKUUpdate):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidProductSKUUpdate.Error(),
			})
		case errors.Is(err, apperr.ErrInvalidProductSKUSpecs):
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: apperr.ErrInvalidProductSKUSpecs.Error(),
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
			slog.ErrorContext(ctx.Request.Context(), "update product SKU", "product_id", productID, "sku_id", skuID, "error", err)
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, newProductSKUResponse(sku))
}

// Delete removes a SKU of a product of the authenticated shop.
//
// @Summary     Delete a product SKU
// @Description Delete a SKU of a product of the authenticated shop.
// @Tags        Product SKUs
// @Param       productID path int true "Product ID"
// @Param       skuID path int true "SKU ID"
// @Success     204
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/products/{productID}/skus/{skuID} [delete]
func (c *ProductSKUController) Delete(ctx *gin.Context) {
	productID, ok := pathID(ctx, "productID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductID.Error(),
		})
		return
	}

	skuID, ok := pathID(ctx, "skuID")
	if !ok {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidProductSKUID.Error(),
		})
		return
	}

	err := c.skuService.Delete(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, productID, skuID)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrProductNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductNotFound.Error(),
			})
		case errors.Is(err, apperr.ErrProductSKUNotFound):
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrProductSKUNotFound.Error(),
			})
		default:
			slog.ErrorContext(ctx.Request.Context(), "delete product SKU", "product_id", productID, "sku_id", skuID, "error", err)
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
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
