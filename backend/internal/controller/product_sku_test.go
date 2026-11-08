package controller_test

import (
	"context"
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
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

const skuBody = `{"specs":{"color":"black","size":"M"},"price_cents":1999,"stock":100}`

func newTestProductSKUController(t *testing.T) (*controller.ProductSKUController, *mocks.MockProductReader, *mocks.MockProductSKURepository) {
	t.Helper()

	productReader := mocks.NewMockProductReader(gomock.NewController(t))
	skuRepository := mocks.NewMockProductSKURepository(gomock.NewController(t))
	return controller.NewProductSKUController(service.NewProductSKUService(productReader, skuRepository)), productReader, skuRepository
}

func newSKUContext(t *testing.T, method, productID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	ctx, rec := newControllerContext(t, method, "/products/"+productID+"/skus", body)
	ctx.Params = gin.Params{{Key: "productID", Value: productID}}
	return ctx, rec
}

func TestProductSKUCreate(t *testing.T) {
	t.Run("returns the created sku", func(t *testing.T) {
		skuController, _, skuRepository := newTestProductSKUController(t)
		skuRepository.EXPECT().
			CreateProductSKU(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, sku *model.ProductSKU) error {
				sku.ID = 1
				return nil
			})

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
				skuController, _, _ := newTestProductSKUController(t)

				ctx, rec := newSKUContext(t, http.MethodPost, productID, skuBody)
				skuController.Create(ctx)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidProductID)
			})
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		skuController, _, _ := newTestProductSKUController(t)

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
				skuController, _, _ := newTestProductSKUController(t)

				ctx, rec := newSKUContext(t, http.MethodPost, "1", payload)
				skuController.Create(ctx)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidJsonFormat)
			})
		}
	})

	t.Run("maps repository errors", func(t *testing.T) {
		tests := []struct {
			name            string
			repositoryError error
			wantStatus      int
			wantError       error
		}{
			{"unknown product", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"specs mismatch", apperr.ErrProductSKUSpecsMismatch, http.StatusBadRequest, apperr.ErrProductSKUSpecsMismatch},
			{"duplicate specs", apperr.ErrProductSKUAlreadyExists, http.StatusConflict, apperr.ErrProductSKUAlreadyExists},
			{"repository failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skuController, _, skuRepository := newTestProductSKUController(t)
				skuRepository.EXPECT().
					CreateProductSKU(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("repository: %w", tt.repositoryError))

				ctx, rec := newSKUContext(t, http.MethodPost, "1", skuBody)
				skuController.Create(ctx)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductSKUList(t *testing.T) {
	product := &model.Product{ID: 1, Name: "T-shirt"}

	t.Run("returns the product skus", func(t *testing.T) {
		skuController, productReader, skuRepository := newTestProductSKUController(t)
		productReader.EXPECT().
			GetProductByID(gomock.Any(), 1).
			Return(product, nil)
		skuRepository.EXPECT().
			ListProductSKUsByProductID(gomock.Any(), 1).
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
		skuController, productReader, skuRepository := newTestProductSKUController(t)
		productReader.EXPECT().
			GetProductByID(gomock.Any(), 1).
			Return(product, nil)
		skuRepository.EXPECT().
			ListProductSKUsByProductID(gomock.Any(), 1).
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
				skuController, _, _ := newTestProductSKUController(t)

				ctx, rec := newSKUContext(t, http.MethodGet, productID, "")
				skuController.List(ctx)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidProductID)
			})
		}
	})

	t.Run("maps repository errors", func(t *testing.T) {
		t.Run("unknown product", func(t *testing.T) {
			skuController, productReader, _ := newTestProductSKUController(t)
			productReader.EXPECT().
				GetProductByID(gomock.Any(), 1).
				Return(nil, fmt.Errorf("repository: %w", apperr.ErrProductNotFound))

			ctx, rec := newSKUContext(t, http.MethodGet, "1", "")
			skuController.List(ctx)

			requireStatus(t, rec, http.StatusNotFound)
			requireError(t, rec, apperr.ErrProductNotFound)
		})

		t.Run("product lookup failure", func(t *testing.T) {
			skuController, productReader, _ := newTestProductSKUController(t)
			productReader.EXPECT().
				GetProductByID(gomock.Any(), 1).
				Return(nil, errors.New("database is down"))

			ctx, rec := newSKUContext(t, http.MethodGet, "1", "")
			skuController.List(ctx)

			requireStatus(t, rec, http.StatusInternalServerError)
			requireError(t, rec, apperr.ErrInternalServer)
		})

		t.Run("sku lookup failure", func(t *testing.T) {
			skuController, productReader, skuRepository := newTestProductSKUController(t)
			productReader.EXPECT().
				GetProductByID(gomock.Any(), 1).
				Return(product, nil)
			skuRepository.EXPECT().
				ListProductSKUsByProductID(gomock.Any(), 1).
				Return(nil, errors.New("database is down"))

			ctx, rec := newSKUContext(t, http.MethodGet, "1", "")
			skuController.List(ctx)

			requireStatus(t, rec, http.StatusInternalServerError)
			requireError(t, rec, apperr.ErrInternalServer)
		})
	})
}
