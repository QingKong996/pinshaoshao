package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/controller"
	"pinshaoshao/internal/feature/order/dto"
	"pinshaoshao/internal/feature/order/mocks"
	"pinshaoshao/internal/feature/order/model"
	"pinshaoshao/internal/feature/order/service"
	usermodule "pinshaoshao/internal/feature/user"
	usermodel "pinshaoshao/internal/feature/user/model"
	"pinshaoshao/internal/http/httpx"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

type stubUserAuthenticator struct {
	user *usermodel.User
}

func (s stubUserAuthenticator) Authenticate(context.Context, string) (*usermodel.User, error) {
	return s.user, nil
}

func newTestOrderController(t *testing.T) (*controller.OrderController, *mocks.MockOrderService) {
	t.Helper()

	orderService := mocks.NewMockOrderService(gomock.NewController(t))
	return controller.NewOrderController(orderService), orderService
}

func newOrderRouter(t *testing.T, orderController *controller.OrderController, user *usermodel.User) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(usermodule.AuthMiddleware(stubUserAuthenticator{user: user}))
	protected.POST("/orders", orderController.Create)
	protected.POST("/orders/direct", orderController.CreateDirect)
	protected.GET("/orders", orderController.List)
	protected.GET("/orders/:orderID", orderController.Get)
	protected.POST("/orders/:orderID/cancel", orderController.Cancel)
	return r
}

func newTestOrder() *model.Order {
	return &model.Order{
		ID:               1,
		Status:           model.OrderStatusPendingPayment,
		TotalAmountCents: 3998,
		ReceiverName:     "Alice",
		SubOrders: []model.SubOrder{
			{
				ID:               2,
				ShopID:           3,
				ShopName:         "Example shop",
				Status:           model.SubOrderStatusPendingPayment,
				TotalAmountCents: 3998,
				Items: []model.OrderItem{
					{
						ID:              4,
						ProductID:       10,
						SKUID:           5,
						ProductName:     "T-shirt",
						Specs:           map[string]string{"color": "black"},
						UnitPriceCents:  1999,
						Quantity:        2,
						LineAmountCents: 3998,
					},
				},
			},
		},
	}
}

func TestOrderCreate(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}

	t.Run("creates an order", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		order := newTestOrder()
		orderService.EXPECT().
			Create(gomock.Any(), user.ID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int, input service.CreateOrderInput) (*model.Order, error) {
				if input.CheckoutKey != "4f2614df-f674-4b5e-a9e4-3fc5e7f62013" {
					t.Fatalf("checkout key = %q", input.CheckoutKey)
				}
				if len(input.Items) != 1 || input.Items[0].CartItemID != 1 {
					t.Fatalf("items = %+v", input.Items)
				}
				return order, nil
			})

		rec := doRequest(
			newOrderRouter(t, orderController, user),
			http.MethodPost,
			"/orders",
			"Bearer session-token",
			`{"checkout_key":"4f2614df-f674-4b5e-a9e4-3fc5e7f62013","items":[{"cart_item_id":1,"quantity":2}],"receiver_name":"Alice","receiver_phone":"13800138000","shipping_address":"Shanghai"}`,
		)

		requireOrderStatus(t, rec, http.StatusCreated)

		var body dto.OrderResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode order response: %v", err)
		}
		if body.ID != 1 || len(body.SubOrders) != 1 || len(body.SubOrders[0].Items) != 1 {
			t.Fatalf("order response = %+v", body)
		}
		if body.SubOrders[0].Items[0].LineAmountCents != 3998 {
			t.Fatalf("line amount = %d, want 3998", body.SubOrders[0].Items[0].LineAmountCents)
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		orderController, _ := newTestOrderController(t)
		rec := doRequest(newOrderRouter(t, orderController, user), http.MethodPost, "/orders", "Bearer session-token", `{bad`)
		requireOrderStatus(t, rec, http.StatusBadRequest)
	})
}

func TestOrderCreateDirect(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}

	t.Run("creates a direct order", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		orderService.EXPECT().
			CreateDirect(gomock.Any(), user.ID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int, input service.CreateDirectOrderInput) (*model.Order, error) {
				if input.CheckoutKey != "4f2614df-f674-4b5e-a9e4-3fc5e7f62013" {
					t.Fatalf("checkout key = %q", input.CheckoutKey)
				}
				if input.SKUID != 5 || input.Quantity != 2 {
					t.Fatalf("item = %d/%d, want 5/2", input.SKUID, input.Quantity)
				}
				return newTestOrder(), nil
			})

		rec := doRequest(
			newOrderRouter(t, orderController, user),
			http.MethodPost,
			"/orders/direct",
			"Bearer session-token",
			`{"checkout_key":"4f2614df-f674-4b5e-a9e4-3fc5e7f62013","sku_id":5,"quantity":2,"receiver_name":"Alice","receiver_phone":"13800138000","shipping_address":"Shanghai"}`,
		)

		requireOrderStatus(t, rec, http.StatusCreated)
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		orderController, _ := newTestOrderController(t)
		rec := doRequest(newOrderRouter(t, orderController, user), http.MethodPost, "/orders/direct", "Bearer session-token", `{bad`)
		requireOrderStatus(t, rec, http.StatusBadRequest)
	})
}

func TestOrderCreateDirectMapsErrors(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  error
	}{
		{"invalid checkout", apperr.ErrInvalidCheckout, http.StatusBadRequest, apperr.ErrInvalidCheckout},
		{"checkout conflict", apperr.ErrCheckoutConflict, http.StatusConflict, apperr.ErrCheckoutConflict},
		{"product unavailable", apperr.ErrProductUnavailable, http.StatusConflict, apperr.ErrProductUnavailable},
		{"insufficient stock", apperr.ErrInsufficientStock, http.StatusConflict, apperr.ErrInsufficientStock},
		{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderController, orderService := newTestOrderController(t)
			orderService.EXPECT().
				CreateDirect(gomock.Any(), user.ID, gomock.Any()).
				Return(nil, tt.err)

			rec := doRequest(
				newOrderRouter(t, orderController, user),
				http.MethodPost,
				"/orders/direct",
				"Bearer session-token",
				`{"checkout_key":"4f2614df-f674-4b5e-a9e4-3fc5e7f62013","sku_id":5,"quantity":2,"receiver_name":"Alice","receiver_phone":"13800138000","shipping_address":"Shanghai"}`,
			)

			requireOrderStatus(t, rec, tt.wantStatus)
			requireOrderError(t, rec, tt.wantError)
		})
	}
}

func TestOrderListAndGet(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}

	t.Run("lists orders", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		orderService.EXPECT().
			List(gomock.Any(), user.ID, model.OrderListFilter{Page: 2, PageSize: 5}).
			DoAndReturn(func(_ context.Context, _ int, filter model.OrderListFilter) (*model.OrderListResult, error) {
				return &model.OrderListResult{
					Orders: []model.OrderSummary{
						{ID: 1, Status: model.OrderStatusPendingPayment, TotalAmountCents: 3998, ItemCount: 2, ShopCount: 1},
					},
					Page:     filter.Page,
					PageSize: filter.PageSize,
					Total:    6,
				}, nil
			})

		rec := doRequest(newOrderRouter(t, orderController, user), http.MethodGet, "/orders?page=2&page_size=5", "Bearer session-token", "")
		requireOrderStatus(t, rec, http.StatusOK)

		var body dto.OrderListResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode order list: %v", err)
		}
		if len(body.Items) != 1 || body.Items[0].ItemCount != 2 || body.TotalPages != 2 {
			t.Fatalf("orders = %+v", body)
		}
	})

	t.Run("rejects invalid pagination", func(t *testing.T) {
		orderController, _ := newTestOrderController(t)
		rec := doRequest(newOrderRouter(t, orderController, user), http.MethodGet, "/orders?page_size=101", "Bearer session-token", "")
		requireOrderStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("gets an order", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		orderService.EXPECT().Get(gomock.Any(), user.ID, 1).Return(newTestOrder(), nil)

		rec := doRequest(newOrderRouter(t, orderController, user), http.MethodGet, "/orders/1", "Bearer session-token", "")
		requireOrderStatus(t, rec, http.StatusOK)
	})

	t.Run("rejects invalid order id", func(t *testing.T) {
		orderController, _ := newTestOrderController(t)
		rec := doRequest(newOrderRouter(t, orderController, user), http.MethodGet, "/orders/abc", "Bearer session-token", "")
		requireOrderStatus(t, rec, http.StatusBadRequest)
	})
}

func TestOrderCancelMapsErrors(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  error
	}{
		{"not found", apperr.ErrOrderNotFound, http.StatusNotFound, apperr.ErrOrderNotFound},
		{"cannot cancel", apperr.ErrOrderCannotCancel, http.StatusConflict, apperr.ErrOrderCannotCancel},
		{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderController, orderService := newTestOrderController(t)
			orderService.EXPECT().Cancel(gomock.Any(), user.ID, 1).Return(nil, tt.err)

			rec := doRequest(newOrderRouter(t, orderController, user), http.MethodPost, "/orders/1/cancel", "Bearer session-token", "")
			requireOrderStatus(t, rec, tt.wantStatus)
			requireOrderError(t, rec, tt.wantError)
		})
	}

	t.Run("cancels an order", func(t *testing.T) {
		orderController, orderService := newTestOrderController(t)
		order := newTestOrder()
		order.Status = model.OrderStatusCancelled
		orderService.EXPECT().Cancel(gomock.Any(), user.ID, 1).Return(order, nil)

		rec := doRequest(newOrderRouter(t, orderController, user), http.MethodPost, "/orders/1/cancel", "Bearer session-token", "")
		requireOrderStatus(t, rec, http.StatusOK)
	})
}

func doRequest(r http.Handler, method, path, authorization, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func requireOrderStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func requireOrderError(t *testing.T, rec *httptest.ResponseRecorder, want error) {
	t.Helper()

	var body httpx.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != want.Error() {
		t.Fatalf("error = %q, want %q", body.Error, want.Error())
	}
}
