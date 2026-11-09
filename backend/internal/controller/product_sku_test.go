package controller_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
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

func newSKURouter(t *testing.T, skuController *controller.ProductSKUController, shop *model.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.ShopAuth(stubShopAuthenticator{shop: shop}))
	protected.GET("/products/:productID/skus", skuController.List)
	protected.POST("/products/:productID/skus", skuController.Create)
	protected.PATCH("/products/:productID/skus/:skuID", skuController.Update)
	protected.DELETE("/products/:productID/skus/:skuID", skuController.Delete)
	return r
}

func newTestSKU() *model.ProductSKU {
	return &model.ProductSKU{
		ID:         1,
		ProductID:  1,
		Specs:      map[string]string{"color": "black", "size": "M"},
		PriceCents: 1999,
		Stock:      100,
		IsEnabled:  true,
	}
}

func TestProductSKUCreate(t *testing.T) {
	activeSKUShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the created sku", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		skuService.EXPECT().
			Create(gomock.Any(), activeSKUShop.ID, 1, gomock.Any()).
			Return(newTestSKU(), nil)

		rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPost, "/products/1/skus", "Bearer session-token", skuBody)

		requireStatus(t, rec, http.StatusCreated)

		var body dto.ProductSKUResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode sku response: %v", err)
		}
		if body.ID != 1 || body.ProductID != 1 {
			t.Fatalf("sku = %+v, want id 1 of product 1", body)
		}
		if body.PriceCents != 1999 || body.Stock != 100 || !body.IsEnabled {
			t.Fatalf("sku = %+v, want created values", body)
		}
		if body.Specs["color"] != "black" || body.Specs["size"] != "M" {
			t.Fatalf("specs = %v", body.Specs)
		}
	})

	t.Run("rejects a malformed product id", func(t *testing.T) {
		for _, productID := range []string{"abc", "0", "-1"} {
			t.Run(productID, func(t *testing.T) {
				skuController, _ := newTestProductSKUController(t)

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPost, "/products/"+productID+"/skus", "Bearer session-token", skuBody)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidProductID)
			})
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		skuController, _ := newTestProductSKUController(t)

		rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPost, "/products/1/skus", "Bearer session-token", `{bad`)

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

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPost, "/products/1/skus", "Bearer session-token", payload)

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
			{"product owned by another shop", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"specs mismatch", apperr.ErrProductSKUSpecsMismatch, http.StatusBadRequest, apperr.ErrProductSKUSpecsMismatch},
			{"invalid specs", apperr.ErrInvalidProductSKUSpecs, http.StatusBadRequest, apperr.ErrInvalidProductSKUSpecs},
			{"duplicate specs", apperr.ErrProductSKUAlreadyExists, http.StatusConflict, apperr.ErrProductSKUAlreadyExists},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skuController, skuService := newTestProductSKUController(t)
				skuService.EXPECT().
					Create(gomock.Any(), activeSKUShop.ID, 1, gomock.Any()).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPost, "/products/1/skus", "Bearer session-token", skuBody)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductSKUList(t *testing.T) {
	activeSKUShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the product skus", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		skuService.EXPECT().
			List(gomock.Any(), activeSKUShop.ID, 1).
			Return([]model.ProductSKU{*newTestSKU()}, nil)

		rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodGet, "/products/1/skus", "Bearer session-token", "")

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
	})

	t.Run("returns an empty array", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		skuService.EXPECT().
			List(gomock.Any(), activeSKUShop.ID, 1).
			Return([]model.ProductSKU{}, nil)

		rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodGet, "/products/1/skus", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Fatalf("body = %q, want []", got)
		}
	})

	t.Run("rejects a malformed product id", func(t *testing.T) {
		for _, productID := range []string{"abc", "0", "-1"} {
			t.Run(productID, func(t *testing.T) {
				skuController, _ := newTestProductSKUController(t)

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodGet, "/products/"+productID+"/skus", "Bearer session-token", "")

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
			{"unknown or foreign product", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skuController, skuService := newTestProductSKUController(t)
				skuService.EXPECT().
					List(gomock.Any(), activeSKUShop.ID, 1).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodGet, "/products/1/skus", "Bearer session-token", "")

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductSKUUpdate(t *testing.T) {
	activeSKUShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the updated sku", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		sku := newTestSKU()
		sku.Stock = 42
		skuService.EXPECT().
			Update(gomock.Any(), activeSKUShop.ID, 1, 5, gomock.Any()).
			Return(sku, nil)

		rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPatch, "/products/1/skus/5", "Bearer session-token", `{"stock":42}`)

		requireStatus(t, rec, http.StatusOK)

		var body dto.ProductSKUResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode sku response: %v", err)
		}
		if body.Stock != 42 {
			t.Fatalf("stock = %d, want 42", body.Stock)
		}
	})

	t.Run("rejects malformed ids", func(t *testing.T) {
		tests := map[string]string{
			"/products/abc/skus/5": apperr.ErrInvalidProductID.Error(),
			"/products/0/skus/5":   apperr.ErrInvalidProductID.Error(),
			"/products/1/skus/abc": apperr.ErrInvalidProductSKUID.Error(),
			"/products/1/skus/-2":  apperr.ErrInvalidProductSKUID.Error(),
		}
		for path, wantError := range tests {
			t.Run(path, func(t *testing.T) {
				skuController, _ := newTestProductSKUController(t)

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPatch, path, "Bearer session-token", `{"stock":42}`)

				requireStatus(t, rec, http.StatusBadRequest)
				if got := decodeError(t, rec); got != wantError {
					t.Fatalf("error = %q, want %q", got, wantError)
				}
			})
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		skuController, _ := newTestProductSKUController(t)

		rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPatch, "/products/1/skus/5", "Bearer session-token", `{bad`)

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

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPatch, "/products/1/skus/5", "Bearer session-token", payload)

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
			{"unknown or foreign product", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"unknown sku", apperr.ErrProductSKUNotFound, http.StatusNotFound, apperr.ErrProductSKUNotFound},
			{"empty update", apperr.ErrInvalidProductSKUUpdate, http.StatusBadRequest, apperr.ErrInvalidProductSKUUpdate},
			{"invalid specs", apperr.ErrInvalidProductSKUSpecs, http.StatusBadRequest, apperr.ErrInvalidProductSKUSpecs},
			{"specs mismatch", apperr.ErrProductSKUSpecsMismatch, http.StatusBadRequest, apperr.ErrProductSKUSpecsMismatch},
			{"duplicate specs", apperr.ErrProductSKUAlreadyExists, http.StatusConflict, apperr.ErrProductSKUAlreadyExists},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skuController, skuService := newTestProductSKUController(t)
				skuService.EXPECT().
					Update(gomock.Any(), activeSKUShop.ID, 1, 5, gomock.Any()).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodPatch, "/products/1/skus/5", "Bearer session-token", `{"stock":42}`)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductSKUDelete(t *testing.T) {
	activeSKUShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("deletes the sku", func(t *testing.T) {
		skuController, skuService := newTestProductSKUController(t)
		skuService.EXPECT().
			Delete(gomock.Any(), activeSKUShop.ID, 1, 5).
			Return(nil)

		rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodDelete, "/products/1/skus/5", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Fatalf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("rejects malformed ids", func(t *testing.T) {
		for _, path := range []string{"/products/abc/skus/5", "/products/1/skus/abc"} {
			t.Run(path, func(t *testing.T) {
				skuController, _ := newTestProductSKUController(t)

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodDelete, path, "Bearer session-token", "")

				requireStatus(t, rec, http.StatusBadRequest)
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
			{"unknown or foreign product", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"unknown sku", apperr.ErrProductSKUNotFound, http.StatusNotFound, apperr.ErrProductSKUNotFound},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skuController, skuService := newTestProductSKUController(t)
				skuService.EXPECT().
					Delete(gomock.Any(), activeSKUShop.ID, 1, 5).
					Return(fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newSKURouter(t, skuController, activeSKUShop), http.MethodDelete, "/products/1/skus/5", "Bearer session-token", "")

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}
