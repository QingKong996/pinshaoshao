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

func newTestProductCategoryController(t *testing.T) (*controller.ProductCategoryController, *mocks.MockProductCategoryService) {
	t.Helper()

	categoryService := mocks.NewMockProductCategoryService(gomock.NewController(t))
	return controller.NewProductCategoryController(categoryService), categoryService
}

func newCategoryRouter(t *testing.T, categoryController *controller.ProductCategoryController, shop *model.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.ShopAuth(stubShopAuthenticator{shop: shop}))
	protected.GET("/categories", categoryController.List)
	protected.POST("/categories", categoryController.Create)
	protected.PATCH("/categories/:categoryID", categoryController.Update)
	protected.DELETE("/categories/:categoryID", categoryController.Delete)
	return r
}

func TestProductCategoryList(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the categories", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().
			List(gomock.Any(), activeShop.ID).
			Return([]model.ProductCategory{{ID: 1, ShopID: activeShop.ID, Name: "clothes", ProductCount: 3}}, nil)

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodGet, "/categories", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)

		var body []dto.ProductCategoryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode category list: %v", err)
		}
		if len(body) != 1 {
			t.Fatalf("category count = %d, want 1", len(body))
		}
		if body[0].Name != "clothes" || body[0].ProductCount != 3 {
			t.Fatalf("category = %+v, want name clothes with 3 products", body[0])
		}
	})

	t.Run("returns an empty array", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().List(gomock.Any(), activeShop.ID).Return([]model.ProductCategory{}, nil)

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodGet, "/categories", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Fatalf("body = %q, want []", got)
		}
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().List(gomock.Any(), activeShop.ID).Return(nil, errors.New("database is down"))

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodGet, "/categories", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestProductCategoryCreate(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}
	body := `{"name":"clothes"}`

	t.Run("returns the created category", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().
			Create(gomock.Any(), activeShop.ID, gomock.Any()).
			Return(&model.ProductCategory{ID: 1, ShopID: activeShop.ID, Name: "clothes"}, nil)

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPost, "/categories", "Bearer session-token", body)

		requireStatus(t, rec, http.StatusCreated)

		var response dto.ProductCategoryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode category response: %v", err)
		}
		if response.ID != 1 || response.Name != "clothes" || response.ProductCount != 0 {
			t.Fatalf("category = %+v, want created category", response)
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		categoryController, _ := newTestProductCategoryController(t)

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPost, "/categories", "Bearer session-token", `{bad`)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		payloads := map[string]string{
			"missing name":  `{}`,
			"empty name":    `{"name":""}`,
			"name too long": fmt.Sprintf(`{"name":%q}`, strings.Repeat("n", 51)),
		}
		for name, payload := range payloads {
			t.Run(name, func(t *testing.T) {
				categoryController, _ := newTestProductCategoryController(t)

				rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPost, "/categories", "Bearer session-token", payload)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidJsonFormat)
			})
		}
	})

	t.Run("rejects a duplicate name", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().
			Create(gomock.Any(), activeShop.ID, gomock.Any()).
			Return(nil, fmt.Errorf("service: %w", apperr.ErrCategoryNameTaken))

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPost, "/categories", "Bearer session-token", body)

		requireStatus(t, rec, http.StatusConflict)
		requireError(t, rec, apperr.ErrCategoryNameTaken)
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().
			Create(gomock.Any(), activeShop.ID, gomock.Any()).
			Return(nil, errors.New("database is down"))

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPost, "/categories", "Bearer session-token", body)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestProductCategoryUpdate(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}
	body := `{"name":"shoes"}`

	t.Run("returns the updated category", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().
			Update(gomock.Any(), activeShop.ID, 3, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int, _ int, req *dto.UpdateProductCategoryRequest) (*model.ProductCategory, error) {
				return &model.ProductCategory{ID: 3, ShopID: activeShop.ID, Name: req.Name, ProductCount: 2}, nil
			})

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPatch, "/categories/3", "Bearer session-token", body)

		requireStatus(t, rec, http.StatusOK)

		var response dto.ProductCategoryResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode category response: %v", err)
		}
		if response.ID != 3 || response.Name != "shoes" || response.ProductCount != 2 {
			t.Fatalf("category = %+v, want updated category", response)
		}
	})

	t.Run("rejects a malformed category id", func(t *testing.T) {
		for _, categoryID := range []string{"abc", "0", "-1"} {
			t.Run(categoryID, func(t *testing.T) {
				categoryController, _ := newTestProductCategoryController(t)

				rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPatch, "/categories/"+categoryID, "Bearer session-token", body)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidCategoryID)
			})
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		categoryController, _ := newTestProductCategoryController(t)

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPatch, "/categories/3", "Bearer session-token", `{bad`)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("maps service errors", func(t *testing.T) {
		tests := []struct {
			name         string
			serviceError error
			wantStatus   int
			wantError    error
		}{
			{"unknown category", apperr.ErrCategoryNotFound, http.StatusNotFound, apperr.ErrCategoryNotFound},
			{"duplicate name", apperr.ErrCategoryNameTaken, http.StatusConflict, apperr.ErrCategoryNameTaken},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				categoryController, categoryService := newTestProductCategoryController(t)
				categoryService.EXPECT().
					Update(gomock.Any(), activeShop.ID, 3, gomock.Any()).
					Return(nil, fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodPatch, "/categories/3", "Bearer session-token", body)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestProductCategoryDelete(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("deletes the category", func(t *testing.T) {
		categoryController, categoryService := newTestProductCategoryController(t)
		categoryService.EXPECT().Delete(gomock.Any(), activeShop.ID, 3).Return(nil)

		rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodDelete, "/categories/3", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Fatalf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("rejects a malformed category id", func(t *testing.T) {
		for _, categoryID := range []string{"abc", "0", "-1"} {
			t.Run(categoryID, func(t *testing.T) {
				categoryController, _ := newTestProductCategoryController(t)

				rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodDelete, "/categories/"+categoryID, "Bearer session-token", "")

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidCategoryID)
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
			{"unknown category", apperr.ErrCategoryNotFound, http.StatusNotFound, apperr.ErrCategoryNotFound},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				categoryController, categoryService := newTestProductCategoryController(t)
				categoryService.EXPECT().
					Delete(gomock.Any(), activeShop.ID, 3).
					Return(fmt.Errorf("service: %w", tt.serviceError))

				rec := doRequest(newCategoryRouter(t, categoryController, activeShop), http.MethodDelete, "/categories/3", "Bearer session-token", "")

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}
