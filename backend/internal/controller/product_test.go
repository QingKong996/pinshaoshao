package controller_test

import (
	"context"
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

func newTestProductController(t *testing.T) (*controller.ProductController, *mocks.MockProductService) {
	t.Helper()

	productService := mocks.NewMockProductService(gomock.NewController(t))
	return controller.NewProductController(productService), productService
}

func newProductRouter(t *testing.T, productController *controller.ProductController, shop *model.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.ShopAuth(stubShopAuthenticator{shop: shop}))
	protected.GET("/products", productController.List)
	protected.POST("/products", productController.Create)
	protected.GET("/products/:productID", productController.Get)
	protected.PATCH("/products/:productID", productController.Update)
	protected.DELETE("/products/:productID", productController.Delete)
	return r
}

func newTestProductSummary() *model.ProductSummary {
	categoryID := 2
	categoryName := "clothes"
	minPrice := int64(1999)
	maxPrice := int64(2999)
	return &model.ProductSummary{
		Product: model.Product{
			ID:          1,
			ShopID:      7,
			Name:        "T-shirt",
			Description: "Cotton T-shirt",
			Status:      model.ProductStatusOnShelf,
			CategoryID:  &categoryID,
		},
		CategoryName:  &categoryName,
		SKUCount:      3,
		TotalStock:    120,
		MinPriceCents: &minPrice,
		MaxPriceCents: &maxPrice,
	}
}

func TestProductList(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the products", func(t *testing.T) {
		productController, productService := newTestProductController(t)
		productService.EXPECT().
			List(gomock.Any(), activeShop.ID).
			Return([]model.ProductSummary{*newTestProductSummary()}, nil)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodGet, "/products", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)

		var body []dto.ProductResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode product list: %v", err)
		}
		if len(body) != 1 {
			t.Fatalf("product count = %d, want 1", len(body))
		}
		if body[0].Status != string(model.ProductStatusOnShelf) || body[0].SKUCount != 3 || body[0].TotalStock != 120 {
			t.Fatalf("product = %+v, want aggregates", body[0])
		}
		if body[0].CategoryID == nil || *body[0].CategoryID != 2 {
			t.Fatalf("category ID = %v, want 2", body[0].CategoryID)
		}
		if body[0].CategoryName == nil || *body[0].CategoryName != "clothes" {
			t.Fatalf("category name = %v, want clothes", body[0].CategoryName)
		}
		if body[0].MinPriceCents == nil || *body[0].MinPriceCents != 1999 {
			t.Fatalf("min price = %v, want 1999", body[0].MinPriceCents)
		}
	})

	t.Run("returns an empty array", func(t *testing.T) {
		productController, productService := newTestProductController(t)
		productService.EXPECT().List(gomock.Any(), activeShop.ID).Return([]model.ProductSummary{}, nil)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodGet, "/products", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Fatalf("body = %q, want []", got)
		}
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		productController, productService := newTestProductController(t)
		productService.EXPECT().List(gomock.Any(), activeShop.ID).Return(nil, errors.New("database is down"))

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodGet, "/products", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestProductGet(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the product", func(t *testing.T) {
		productController, productService := newTestProductController(t)
		productService.EXPECT().Get(gomock.Any(), activeShop.ID, 1).Return(newTestProductSummary(), nil)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodGet, "/products/1", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)

		var body dto.ProductResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode product response: %v", err)
		}
		if body.ID != 1 || body.Name != "T-shirt" {
			t.Fatalf("product = %+v, want T-shirt with id 1", body)
		}
	})

	t.Run("rejects a malformed product id", func(t *testing.T) {
		for _, productID := range []string{"abc", "0", "-1"} {
			t.Run(productID, func(t *testing.T) {
				productController, _ := newTestProductController(t)

				rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodGet, "/products/"+productID, "Bearer session-token", "")

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
				productController, productService := newTestProductController(t)
				productService.EXPECT().
					Get(gomock.Any(), activeShop.ID, 1).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodGet, "/products/1", "Bearer session-token", "")

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductCreate(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}
	body := `{"name":"T-shirt","description":"Cotton T-shirt","category_id":2,"status":"on_shelf"}`

	t.Run("returns the created product", func(t *testing.T) {
		productController, productService := newTestProductController(t)
		productService.EXPECT().
			Create(gomock.Any(), activeShop.ID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int, req *dto.CreateProductRequest) (*model.ProductSummary, error) {
				if req.Status != "on_shelf" || req.CategoryID == nil || *req.CategoryID != 2 {
					t.Fatalf("request = %+v, want on_shelf in category 2", req)
				}
				return newTestProductSummary(), nil
			})

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPost, "/products", "Bearer session-token", body)

		requireStatus(t, rec, http.StatusCreated)

		var response dto.ProductResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode product response: %v", err)
		}
		if response.ID != 1 {
			t.Fatalf("product ID = %d, want 1", response.ID)
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		productController, _ := newTestProductController(t)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPost, "/products", "Bearer session-token", `{bad`)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		payloads := map[string]string{
			"missing name":   `{"description":"Cotton T-shirt"}`,
			"empty name":     `{"name":""}`,
			"name too long":  fmt.Sprintf(`{"name":%q}`, strings.Repeat("n", 101)),
			"unknown status": `{"name":"T-shirt","status":"sold_out"}`,
		}
		for name, payload := range payloads {
			t.Run(name, func(t *testing.T) {
				productController, _ := newTestProductController(t)

				rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPost, "/products", "Bearer session-token", payload)

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
			{"invalid category", apperr.ErrInvalidCategoryID, http.StatusBadRequest, apperr.ErrInvalidCategoryID},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				productController, productService := newTestProductController(t)
				productService.EXPECT().
					Create(gomock.Any(), activeShop.ID, gomock.Any()).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPost, "/products", "Bearer session-token", body)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductUpdate(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}
	body := `{"name":"Hoodie","description":"Warm hoodie","category_id":null,"status":"off_shelf"}`

	t.Run("returns the updated product", func(t *testing.T) {
		productController, productService := newTestProductController(t)
		productService.EXPECT().
			Update(gomock.Any(), activeShop.ID, 1, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int, _ int, req *dto.UpdateProductRequest) (*model.ProductSummary, error) {
				if req.Status != "off_shelf" || req.CategoryID != nil {
					t.Fatalf("request = %+v, want off_shelf without a category", req)
				}
				updated := newTestProductSummary()
				updated.Name = req.Name
				updated.Status = model.ProductStatusOffShelf
				return updated, nil
			})

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPatch, "/products/1", "Bearer session-token", body)

		requireStatus(t, rec, http.StatusOK)

		var response dto.ProductResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode product response: %v", err)
		}
		if response.Name != "Hoodie" || response.Status != "off_shelf" {
			t.Fatalf("product = %+v, want updated product", response)
		}
	})

	t.Run("rejects a malformed product id", func(t *testing.T) {
		productController, _ := newTestProductController(t)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPatch, "/products/abc", "Bearer session-token", body)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidProductID)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		productController, _ := newTestProductController(t)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPatch, "/products/1", "Bearer session-token", `{bad`)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		payloads := map[string]string{
			"missing name":   `{"status":"off_shelf"}`,
			"missing status": `{"name":"Hoodie"}`,
			"unknown status": `{"name":"Hoodie","status":"sold_out"}`,
		}
		for name, payload := range payloads {
			t.Run(name, func(t *testing.T) {
				productController, _ := newTestProductController(t)

				rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPatch, "/products/1", "Bearer session-token", payload)

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
			{"invalid category", apperr.ErrInvalidCategoryID, http.StatusBadRequest, apperr.ErrInvalidCategoryID},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				productController, productService := newTestProductController(t)
				productService.EXPECT().
					Update(gomock.Any(), activeShop.ID, 1, gomock.Any()).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodPatch, "/products/1", "Bearer session-token", body)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductDelete(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("deletes the product", func(t *testing.T) {
		productController, productService := newTestProductController(t)
		productService.EXPECT().Delete(gomock.Any(), activeShop.ID, 1).Return(nil)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodDelete, "/products/1", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Fatalf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("rejects a malformed product id", func(t *testing.T) {
		productController, _ := newTestProductController(t)

		rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodDelete, "/products/abc", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidProductID)
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
				productController, productService := newTestProductController(t)
				productService.EXPECT().
					Delete(gomock.Any(), activeShop.ID, 1).
					Return(fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newProductRouter(t, productController, activeShop), http.MethodDelete, "/products/1", "Bearer session-token", "")

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}
