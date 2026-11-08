package controller_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

const skuBody = `{"specs":{"color":"black","size":"M"},"price_cents":1999,"stock":100}`

func newTestProductSKUController(t *testing.T) (*controller.ProductSKUController, *mocks.MockProductSKUService) {
	t.Helper()

	skuService := mocks.NewMockProductSKUService(gomock.NewController(t))
	return controller.NewProductSKUController(skuService), skuService
}

func newSKUContext(t *testing.T, method, productID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	ctx, rec := newControllerContext(t, method, "/products/"+productID+"/skus", body)
	ctx.Params = gin.Params{{Key: "productID", Value: productID}}
	return ctx, rec
}

func TestProductSKUCreate(t *testing.T) {
	t.Run("returns the created sku", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		sku := &model.ProductSKU{
			ID:         1,
			ProductID:  1,
			Specs:      map[string]string{"color": "black", "size": "M"},
			PriceCents: 1999,
			Stock:      100,
			IsEnabled:  true,
		}
		skuService.EXPECT().
			Create(gomock.Any(), 1, gomock.Any()).
			Return(sku, nil)

		ctx, rec := newSKUContext(t, http.MethodPost, "1", skuBody)
		skuController.Create(ctx)

		requireStatus(t, rec, http.StatusCreated)

		var body dto.ProductSKUResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode sku response: %v", err)
		}
		if body.ID != 1 {
			t.Fatalf("id = %d, want 1", body.ID)
		}
		if body.ProductID != 1 {
			t.Fatalf("product id = %d, want 1", body.ProductID)
		}
		if body.PriceCents != 1999 {
			t.Fatalf("price cents = %d, want 1999", body.PriceCents)
		}
		if body.Stock != 100 {
			t.Fatalf("stock = %d, want 100", body.Stock)
		}
		if !body.IsEnabled {
			t.Fatal("sku is disabled, want enabled by default")
		}
		if body.Specs["color"] != "black" || body.Specs["size"] != "M" {
			t.Fatalf("specs = %v", body.Specs)
		}
	})

	t.Run("rejects a malformed product id", func(t *testing.T) {
		for _, productID := range []string{"abc", "0", "-1"} {
			t.Run(productID, func(t *testing.T) {
				skuController, _ := newTestProductSKUController(t)

				ctx, rec := newSKUContext(t, http.MethodPost, productID, skuBody)
				skuController.Create(ctx)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidProductID)
			})
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		skuController, _ := newTestProductSKUController(t)

		ctx, rec := newSKUContext(t, http.MethodPost, "1", `{bad`)
		skuController.Create(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		payloads := map[string]string{
			"negative price": `{"price_cents":-1}`,
			"negative stock": `{"stock":-1}`,
			"stock overflow": `{"stock":3000000000}`,
		}
		for name, payload := range payloads {
			t.Run(name, func(t *testing.T) {
				skuController, _ := newTestProductSKUController(t)

				ctx, rec := newSKUContext(t, http.MethodPost, "1", payload)
				skuController.Create(ctx)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidJsonFormat)
			})
		}
	})

	t.Run("maps service errors", func(t *testing.T) {
		tests := []struct {
			name         string
			serviceError error
			wantStatus   int
			wantError    error
		}{
			{"unknown product", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"specs mismatch", apperr.ErrProductSKUSpecsMismatch, http.StatusBadRequest, apperr.ErrProductSKUSpecsMismatch},
			{"duplicate specs", apperr.ErrProductSKUAlreadyExists, http.StatusConflict, apperr.ErrProductSKUAlreadyExists},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skuController, skuService := newTestProductSKUController(t)
				skuService.EXPECT().
					Create(gomock.Any(), 1, gomock.Any()).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				ctx, rec := newSKUContext(t, http.MethodPost, "1", skuBody)
				skuController.Create(ctx)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductSKUList(t *testing.T) {
	t.Run("returns the product skus", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		skuService.EXPECT().
			List(gomock.Any(), 1).
			Return([]model.ProductSKU{{
				ID:         1,
				ProductID:  1,
				Specs:      map[string]string{"color": "black"},
				PriceCents: 1999,
				Stock:      100,
				IsEnabled:  true,
			}}, nil)

		ctx, rec := newSKUContext(t, http.MethodGet, "1", "")
		skuController.List(ctx)

		requireStatus(t, rec, http.StatusOK)

		var skus []dto.ProductSKUResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &skus); err != nil {
			t.Fatalf("decode sku list: %v", err)
		}
		if len(skus) != 1 {
			t.Fatalf("sku count = %d, want 1", len(skus))
		}
		if skus[0].ID != 1 || skus[0].ProductID != 1 {
			t.Fatalf("sku = %+v, want id 1 of product 1", skus[0])
		}
		if skus[0].Specs["color"] != "black" {
			t.Fatalf("specs = %v", skus[0].Specs)
		}
	})

	t.Run("returns an empty array", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		skuService.EXPECT().
			List(gomock.Any(), 1).
			Return([]model.ProductSKU{}, nil)

		ctx, rec := newSKUContext(t, http.MethodGet, "1", "")
		skuController.List(ctx)

		requireStatus(t, rec, http.StatusOK)
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Fatalf("body = %q, want []", got)
		}
	})

	t.Run("rejects a malformed product id", func(t *testing.T) {
		for _, productID := range []string{"abc", "0", "-1"} {
			t.Run(productID, func(t *testing.T) {
				skuController, _ := newTestProductSKUController(t)

				ctx, rec := newSKUContext(t, http.MethodGet, productID, "")
				skuController.List(ctx)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidProductID)
			})
		}
	})

	t.Run("maps service errors", func(t *testing.T) {
		tests := []struct {
			name         string
			serviceError error
			wantStatus   int
			wantError    error
		}{
			{"unknown product", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skuController, skuService := newTestProductSKUController(t)
				skuService.EXPECT().
					List(gomock.Any(), 1).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				ctx, rec := newSKUContext(t, http.MethodGet, "1", "")
				skuController.List(ctx)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}
