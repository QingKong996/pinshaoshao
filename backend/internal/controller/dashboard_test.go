package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
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

func newTestDashboardController(t *testing.T) (*controller.DashboardController, *mocks.MockDashboardService) {
	t.Helper()

	dashboardService := mocks.NewMockDashboardService(gomock.NewController(t))
	return controller.NewDashboardController(dashboardService), dashboardService
}

func newDashboardRouter(t *testing.T, dashboardController *controller.DashboardController, shop *model.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.ShopAuth(stubShopAuthenticator{shop: shop}))
	protected.GET("/dashboard", dashboardController.Overview)
	return r
}

func TestDashboardOverview(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the dashboard", func(t *testing.T) {
		dashboardController, dashboardService := newTestDashboardController(t)
		dashboardService.EXPECT().
			Overview(gomock.Any(), activeShop.ID).
			Return(&model.DashboardStats{
				ProductCount:        12,
				OnShelfProductCount: 8,
				SKUCount:            34,
				CategoryCount:       4,
				TotalStock:          500,
				InventoryValueCents: 123456,
				OutOfStockSKUCount:  1,
			}, nil)

		rec := doRequest(newDashboardRouter(t, dashboardController, activeShop), http.MethodGet, "/dashboard", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)

		var body dto.DashboardResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode dashboard response: %v", err)
		}
		if body.ProductCount != 12 || body.OnShelfProductCount != 8 || body.SKUCount != 34 || body.CategoryCount != 4 {
			t.Fatalf("dashboard = %+v, want counters", body)
		}
		if body.TotalStock != 500 || body.InventoryValueCents != 123456 {
			t.Fatalf("dashboard = %+v, want stock counters", body)
		}
		if body.OutOfStockSKUCount != 1 {
			t.Fatalf("out of stock count = %d, want 1", body.OutOfStockSKUCount)
		}
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		dashboardController, dashboardService := newTestDashboardController(t)
		dashboardService.EXPECT().Overview(gomock.Any(), activeShop.ID).Return(nil, errors.New("database is down"))

		rec := doRequest(newDashboardRouter(t, dashboardController, activeShop), http.MethodGet, "/dashboard", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}
