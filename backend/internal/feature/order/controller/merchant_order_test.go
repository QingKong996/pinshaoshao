package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/controller"
	"pinshaoshao/internal/feature/order/dto"
	"pinshaoshao/internal/feature/order/model"
	shopmodule "pinshaoshao/internal/feature/shop"
	shopmodel "pinshaoshao/internal/feature/shop/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

type stubShopAuthenticator struct {
	shop *shopmodel.Shop
}

func (s stubShopAuthenticator) Authenticate(context.Context, string) (*shopmodel.Shop, error) {
	return s.shop, nil
}

func newMerchantOrderRouter(t *testing.T, orderController *controller.OrderController, shop *shopmodel.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("/merchant")
	protected.Use(shopmodule.AuthMiddleware(stubShopAuthenticator{shop: shop}))
	protected.GET("/orders", orderController.ListMerchant)
	protected.GET("/orders/:subOrderID", orderController.GetMerchant)
	return r
}

func TestMerchantOrderList(t *testing.T) {
	shop := &shopmodel.Shop{ID: 9, Username: "shop01", Status: shopmodel.ShopStatusActive}
	paidAt := time.Date(2026, time.November, 20, 12, 5, 0, 0, time.UTC)

	t.Run("lists merchant orders", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		orderService.EXPECT().
			ListMerchant(gomock.Any(), shop.ID, model.MerchantOrderListFilter{
				Status:   model.SubOrderStatusAwaitingShipment,
				Page:     2,
				PageSize: 5,
			}).
			Return(&model.MerchantOrderListResult{
				Orders: []model.MerchantOrderSummary{
					{
						ID:               4,
						OrderID:          12,
						Status:           model.SubOrderStatusAwaitingShipment,
						TotalAmountCents: 3998,
						ItemCount:        2,
						ReceiverName:     "Alice",
						ReceiverPhone:    "1234567890",
						ShippingAddress:  "Shanghai",
						PaidAt:           &paidAt,
						CreatedAt:        paidAt.Add(-time.Minute),
					},
				},
				Page:     2,
				PageSize: 5,
				Total:    6,
			}, nil)

		rec := doRequest(
			newMerchantOrderRouter(t, orderController, shop),
			http.MethodGet,
			"/merchant/orders?page=2&page_size=5&status=awaiting_shipment",
			"Bearer session-token",
			"",
		)
		requireOrderStatus(t, rec, http.StatusOK)

		var body dto.MerchantOrderListResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode merchant order list: %v", err)
		}
		if len(body.Items) != 1 ||
			body.Items[0].ID != 4 ||
			body.Items[0].OrderID != 12 ||
			body.Items[0].ItemCount != 2 ||
			body.TotalPages != 2 {
			t.Fatalf("merchant orders = %+v", body)
		}
	})

	t.Run("rejects invalid status", func(t *testing.T) {
		orderController, _ := newTestOrderController(t)
		rec := doRequest(
			newMerchantOrderRouter(t, orderController, shop),
			http.MethodGet,
			"/merchant/orders?status=unknown",
			"Bearer session-token",
			"",
		)
		requireOrderStatus(t, rec, http.StatusBadRequest)
		requireOrderError(t, rec, apperr.ErrInvalidOrderList)
	})
}

func TestMerchantOrderDetail(t *testing.T) {
	shop := &shopmodel.Shop{ID: 9, Username: "shop01", Status: shopmodel.ShopStatusActive}
	paidAt := time.Date(2026, time.November, 20, 12, 5, 0, 0, time.UTC)

	t.Run("returns an owned sub-order", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		orderService.EXPECT().
			GetMerchant(gomock.Any(), shop.ID, 4).
			Return(&model.MerchantOrderDetail{
				ID:               4,
				OrderID:          12,
				Status:           model.SubOrderStatusAwaitingShipment,
				TotalAmountCents: 3998,
				ItemCount:        2,
				ReceiverName:     "Alice",
				ReceiverPhone:    "1234567890",
				ShippingAddress:  "Shanghai",
				PaidAt:           &paidAt,
				CreatedAt:        paidAt.Add(-time.Minute),
				UpdatedAt:        paidAt,
				Items: []model.OrderItem{
					{
						ID:              8,
						ProductID:       10,
						SKUID:           5,
						ProductName:     "T-shirt",
						Specs:           map[string]string{"color": "black"},
						UnitPriceCents:  1999,
						Quantity:        2,
						LineAmountCents: 3998,
					},
				},
			}, nil)

		rec := doRequest(
			newMerchantOrderRouter(t, orderController, shop),
			http.MethodGet,
			"/merchant/orders/4",
			"Bearer session-token",
			"",
		)
		requireOrderStatus(t, rec, http.StatusOK)

		var body dto.MerchantOrderDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode merchant order detail: %v", err)
		}
		if body.ID != 4 ||
			body.OrderID != 12 ||
			body.ItemCount != 2 ||
			body.Status != "awaiting_shipment" ||
			len(body.Items) != 1 ||
			body.Items[0].ProductName != "T-shirt" {
			t.Fatalf("merchant order detail = %+v", body)
		}
	})

	t.Run("rejects an invalid id", func(t *testing.T) {
		orderController, _ := newTestOrderController(t)
		rec := doRequest(
			newMerchantOrderRouter(t, orderController, shop),
			http.MethodGet,
			"/merchant/orders/0",
			"Bearer session-token",
			"",
		)
		requireOrderStatus(t, rec, http.StatusBadRequest)
		requireOrderError(t, rec, apperr.ErrInvalidOrderID)
	})

	t.Run("returns not found for another shop's order", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		orderService.EXPECT().
			GetMerchant(gomock.Any(), shop.ID, 4).
			Return(nil, apperr.ErrOrderNotFound)

		rec := doRequest(
			newMerchantOrderRouter(t, orderController, shop),
			http.MethodGet,
			"/merchant/orders/4",
			"Bearer session-token",
			"",
		)
		requireOrderStatus(t, rec, http.StatusNotFound)
		requireOrderError(t, rec, apperr.ErrOrderNotFound)
	})
}
