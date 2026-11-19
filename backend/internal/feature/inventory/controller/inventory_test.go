package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"pinshaoshao/internal/feature/inventory/apperr"
	"pinshaoshao/internal/feature/inventory/controller"
	"pinshaoshao/internal/feature/inventory/dto"
	"pinshaoshao/internal/feature/inventory/mocks"
	"pinshaoshao/internal/feature/inventory/model"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	"pinshaoshao/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func newTestInventoryController(t *testing.T) (*controller.InventoryController, *mocks.MockInventoryService) {
	t.Helper()

	inventoryService := mocks.NewMockInventoryService(gomock.NewController(t))
	return controller.NewInventoryController(inventoryService), inventoryService
}

func newInventoryRouter(t *testing.T, inventoryController *controller.InventoryController, shop *shopmodel.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.ShopAuth(stubShopAuthenticator{shop: shop}))
	protected.GET("/inventory", inventoryController.List)
	protected.PATCH("/inventory/:skuID", inventoryController.Update)
	return r
}

func newTestInventoryItem() model.InventoryItem {
	categoryID := 2
	categoryName := "clothes"
	return model.InventoryItem{
		SKUID:        5,
		ProductID:    1,
		ProductName:  "T-shirt",
		CategoryID:   &categoryID,
		CategoryName: &categoryName,
		Specs:        map[string]string{"color": "black"},
		PriceCents:   1999,
		TotalStock:   2,
		FrozenStock:  1,
		Version:      3,
		IsEnabled:    true,
	}
}

func TestInventoryList(t *testing.T) {
	activeShop := &shopmodel.Shop{ID: 7, Username: testShopUsername, Status: shopmodel.ShopStatusActive}

	t.Run("returns the inventory", func(t *testing.T) {
		inventoryController, inventoryService := newTestInventoryController(t)
		inventoryService.EXPECT().
			List(gomock.Any(), activeShop.ID).
			Return([]model.InventoryItem{newTestInventoryItem()}, nil)

		rec := doRequest(newInventoryRouter(t, inventoryController, activeShop), http.MethodGet, "/inventory", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)

		var body []dto.InventoryItemResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode inventory response: %v", err)
		}
		if len(body) != 1 {
			t.Fatalf("item count = %d, want 1", len(body))
		}
		item := body[0]
		if item.SKUID != 5 || item.ProductID != 1 || item.ProductName != "T-shirt" {
			t.Fatalf("item = %+v, want sku 5 of product 1", item)
		}
		if item.CategoryName == nil || *item.CategoryName != "clothes" {
			t.Fatalf("category name = %v, want clothes", item.CategoryName)
		}
		if item.TotalStock != 2 || item.FrozenStock != 1 || item.AvailableStock != 1 || item.Version != 3 || !item.IsEnabled || item.Specs["color"] != "black" {
			t.Fatalf("item = %+v, want inventory values", item)
		}
	})

	t.Run("returns an empty item array", func(t *testing.T) {
		inventoryController, inventoryService := newTestInventoryController(t)
		inventoryService.EXPECT().List(gomock.Any(), activeShop.ID).Return([]model.InventoryItem{}, nil)

		rec := doRequest(newInventoryRouter(t, inventoryController, activeShop), http.MethodGet, "/inventory", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)

		var body []dto.InventoryItemResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode inventory response: %v", err)
		}
		if body == nil {
			t.Fatal("body is null, want empty array")
		}
		if len(body) != 0 {
			t.Fatalf("item count = %d, want 0", len(body))
		}
		if got := strings.TrimSpace(rec.Body.String()); got != "[]" {
			t.Fatalf("body = %q, want empty array", got)
		}
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		inventoryController, inventoryService := newTestInventoryController(t)
		inventoryService.EXPECT().List(gomock.Any(), activeShop.ID).Return(nil, errors.New("database is down"))

		rec := doRequest(newInventoryRouter(t, inventoryController, activeShop), http.MethodGet, "/inventory", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestInventoryUpdate(t *testing.T) {
	activeShop := &shopmodel.Shop{ID: 7, Username: testShopUsername, Status: shopmodel.ShopStatusActive}

	t.Run("returns the updated inventory", func(t *testing.T) {
		inventoryController, inventoryService := newTestInventoryController(t)
		item := newTestInventoryItem()
		item.TotalStock = 12
		item.Version = 4
		inventoryService.EXPECT().
			Update(gomock.Any(), activeShop.ID, 5, gomock.Any()).
			Return(&item, nil)

		rec := doRequest(
			newInventoryRouter(t, inventoryController, activeShop),
			http.MethodPatch,
			"/inventory/5",
			"Bearer session-token",
			`{"total_stock":12,"version":3}`,
		)

		requireStatus(t, rec, http.StatusOK)

		var body dto.InventoryItemResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode inventory response: %v", err)
		}
		if body.TotalStock != 12 || body.AvailableStock != 11 || body.Version != 4 {
			t.Fatalf("inventory = %+v, want updated total stock", body)
		}
	})

	t.Run("rejects an invalid request", func(t *testing.T) {
		tests := map[string]struct {
			path string
			body string
		}{
			"invalid sku id":  {path: "/inventory/abc", body: `{"total_stock":12,"version":3}`},
			"missing total":   {path: "/inventory/5", body: `{"version":3}`},
			"negative total":  {path: "/inventory/5", body: `{"total_stock":-1,"version":3}`},
			"missing version": {path: "/inventory/5", body: `{"total_stock":12}`},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				inventoryController, _ := newTestInventoryController(t)

				rec := doRequest(
					newInventoryRouter(t, inventoryController, activeShop),
					http.MethodPatch,
					tt.path,
					"Bearer session-token",
					tt.body,
				)

				requireStatus(t, rec, http.StatusBadRequest)
			})
		}
	})

	t.Run("maps service errors", func(t *testing.T) {
		tests := []struct {
			name       string
			err        error
			wantStatus int
			wantError  error
		}{
			{"unknown sku", apperr.ErrInventoryItemNotFound, http.StatusNotFound, apperr.ErrInventoryItemNotFound},
			{"version conflict", apperr.ErrInventoryVersion, http.StatusConflict, apperr.ErrInventoryVersion},
			{"below frozen", apperr.ErrInventoryBelowFrozen, http.StatusConflict, apperr.ErrInventoryBelowFrozen},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				inventoryController, inventoryService := newTestInventoryController(t)
				inventoryService.EXPECT().
					Update(gomock.Any(), activeShop.ID, 5, gomock.Any()).
					Return(nil, tt.err)

				rec := doRequest(
					newInventoryRouter(t, inventoryController, activeShop),
					http.MethodPatch,
					"/inventory/5",
					"Bearer session-token",
					`{"total_stock":12,"version":3}`,
				)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}
