package controller_test

import (
	"encoding/json"
	"errors"
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

func newTestInventoryController(t *testing.T) (*controller.InventoryController, *mocks.MockInventoryService) {
	t.Helper()

	inventoryService := mocks.NewMockInventoryService(gomock.NewController(t))
	return controller.NewInventoryController(inventoryService), inventoryService
}

func newInventoryRouter(t *testing.T, inventoryController *controller.InventoryController, shop *model.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.ShopAuth(stubShopAuthenticator{shop: shop}))
	protected.GET("/inventory", inventoryController.List)
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
		Stock:        2,
		IsEnabled:    true,
	}
}

func TestInventoryList(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

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
		if item.Stock != 2 || !item.IsEnabled || item.Specs["color"] != "black" {
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
