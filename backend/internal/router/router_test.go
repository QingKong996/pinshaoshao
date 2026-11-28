package router_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	addressmodule "pinshaoshao/internal/feature/address"
	cartmodule "pinshaoshao/internal/feature/cart"
	feedmodule "pinshaoshao/internal/feature/feed"
	inventorymodule "pinshaoshao/internal/feature/inventory"
	logisticsmodule "pinshaoshao/internal/feature/logistics"
	ordermodule "pinshaoshao/internal/feature/order"
	paymentmodule "pinshaoshao/internal/feature/payment"
	productmodule "pinshaoshao/internal/feature/product"
	shopmodule "pinshaoshao/internal/feature/shop"
	shopapperr "pinshaoshao/internal/feature/shop/apperr"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	usermodule "pinshaoshao/internal/feature/user"
	userapperr "pinshaoshao/internal/feature/user/apperr"
	usermodel "pinshaoshao/internal/feature/user/model"
	"pinshaoshao/internal/http/httpx"
	mockauth "pinshaoshao/internal/mock/auth"
	mocklogisticsmodule "pinshaoshao/internal/mock/logistics"
	mockpaymentmodule "pinshaoshao/internal/mock/payment"
	"pinshaoshao/internal/outbox"
	"pinshaoshao/internal/router"

	"github.com/gin-gonic/gin"
)

type invalidUserSession struct{}

func (invalidUserSession) Authenticate(context.Context, string) (*usermodel.User, error) {
	return nil, userapperr.ErrInvalidSession
}

type invalidShopSession struct{}

func (invalidShopSession) Authenticate(context.Context, string) (*shopmodel.Shop, error) {
	return nil, shopapperr.ErrInvalidShopSession
}

type routeCase struct {
	method      string
	pattern     string
	requestPath string
	public      bool
	publicCode  int
	authError   string
}

var routeCases = []routeCase{
	{method: http.MethodPost, pattern: "/api/v1/login", requestPath: "/api/v1/login", public: true},
	{method: http.MethodPost, pattern: "/api/v1/register", requestPath: "/api/v1/register", public: true},
	{method: http.MethodPost, pattern: "/api/v1/logout", requestPath: "/api/v1/logout", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/cart", requestPath: "/api/v1/cart", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/cart/items", requestPath: "/api/v1/cart/items", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/cart/items/:itemID", requestPath: "/api/v1/cart/items/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/cart/items/:itemID", requestPath: "/api/v1/cart/items/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/addresses", requestPath: "/api/v1/addresses", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/addresses", requestPath: "/api/v1/addresses", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPut, pattern: "/api/v1/addresses/:addressID", requestPath: "/api/v1/addresses/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/addresses/:addressID", requestPath: "/api/v1/addresses/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/addresses/:addressID/default", requestPath: "/api/v1/addresses/1/default", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/orders", requestPath: "/api/v1/orders", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/orders/direct", requestPath: "/api/v1/orders/direct", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/orders", requestPath: "/api/v1/orders", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/orders/:orderID", requestPath: "/api/v1/orders/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/orders/:orderID/cancel", requestPath: "/api/v1/orders/1/cancel", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/orders", requestPath: "/api/v1/merchant/orders", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/orders/:subOrderID", requestPath: "/api/v1/merchant/orders/1", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/orders/:subOrderID/receiver", requestPath: "/api/v1/merchant/orders/1/receiver", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/payments", requestPath: "/api/v1/payments", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/payments/:paymentNo", requestPath: "/api/v1/payments/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/payments/:paymentNo/pay", requestPath: "/api/v1/payments/1/pay", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/payments/:paymentNo/close", requestPath: "/api/v1/payments/1/close", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/payment/webhook", requestPath: "/api/v1/payment/webhook", public: true, publicCode: http.StatusUnauthorized},
	{method: http.MethodGet, pattern: "/api/v1/shipments", requestPath: "/api/v1/shipments", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/shipments/:shipmentNo", requestPath: "/api/v1/shipments/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/logistics/webhook", requestPath: "/api/v1/logistics/webhook", public: true, publicCode: http.StatusUnauthorized},
	{method: http.MethodPost, pattern: "/api/v1/merchant/shipments", requestPath: "/api/v1/merchant/shipments", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/shipments", requestPath: "/api/v1/merchant/shipments", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/shipments/:shipmentNo", requestPath: "/api/v1/merchant/shipments/1", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/feed", requestPath: "/api/v1/feed", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/products/:productID", requestPath: "/api/v1/products/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/login", requestPath: "/api/v1/merchant/login", public: true},
	{method: http.MethodPost, pattern: "/api/v1/merchant/register", requestPath: "/api/v1/merchant/register", public: true},
	{method: http.MethodPost, pattern: "/api/v1/merchant/logout", requestPath: "/api/v1/merchant/logout", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/shop", requestPath: "/api/v1/merchant/shop", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/shop", requestPath: "/api/v1/merchant/shop", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/categories", requestPath: "/api/v1/merchant/categories", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/categories", requestPath: "/api/v1/merchant/categories", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/categories/:categoryID", requestPath: "/api/v1/merchant/categories/1", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/merchant/categories/:categoryID", requestPath: "/api/v1/merchant/categories/1", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/products", requestPath: "/api/v1/merchant/products", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/products", requestPath: "/api/v1/merchant/products", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/products/:productID", requestPath: "/api/v1/merchant/products/1", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/products/:productID", requestPath: "/api/v1/merchant/products/1", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/merchant/products/:productID", requestPath: "/api/v1/merchant/products/1", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/products/:productID/skus", requestPath: "/api/v1/merchant/products/1/skus", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/products/:productID/skus", requestPath: "/api/v1/merchant/products/1/skus", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/products/:productID/skus/:skuID", requestPath: "/api/v1/merchant/products/1/skus/2", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/merchant/products/:productID/skus/:skuID", requestPath: "/api/v1/merchant/products/1/skus/2", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/inventory", requestPath: "/api/v1/merchant/inventory", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/inventory/:skuID", requestPath: "/api/v1/merchant/inventory/2", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/dashboard", requestPath: "/api/v1/merchant/dashboard", authError: shopapperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/mockpay/api/v1/auth/login", requestPath: "/mockpay/api/v1/auth/login", public: true},
	{method: http.MethodPost, pattern: "/mockpay/api/v1/auth/logout", requestPath: "/mockpay/api/v1/auth/logout", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mockpay/api/v1/payments", requestPath: "/mockpay/api/v1/payments", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/mockpay/api/v1/payments/:paymentNo", requestPath: "/mockpay/api/v1/payments/1", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mockpay/api/v1/payments/:paymentNo/close", requestPath: "/mockpay/api/v1/payments/1/close", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mockpay/api/v1/payments/:paymentNo/refunds", requestPath: "/mockpay/api/v1/payments/1/refunds", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/mockpay/api/v1/payments/:paymentNo/notifications", requestPath: "/mockpay/api/v1/payments/1/notifications", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mockpay/api/v1/payments/:paymentNo/notifications/retry", requestPath: "/mockpay/api/v1/payments/1/notifications/retry", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/mockpay/checkout/:paymentNo", requestPath: "/mockpay/checkout/1"},
	{method: http.MethodPost, pattern: "/mockpay/checkout/:paymentNo/pay", requestPath: "/mockpay/checkout/1/pay"},
	{method: http.MethodPost, pattern: "/mockpay/checkout/:paymentNo/close", requestPath: "/mockpay/checkout/1/close"},
	{method: http.MethodPost, pattern: "/mocklogistics/api/v1/auth/login", requestPath: "/mocklogistics/api/v1/auth/login", public: true},
	{method: http.MethodPost, pattern: "/mocklogistics/api/v1/auth/logout", requestPath: "/mocklogistics/api/v1/auth/logout", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mocklogistics/api/v1/shipments", requestPath: "/mocklogistics/api/v1/shipments", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/mocklogistics/api/v1/shipments/:shipmentNo", requestPath: "/mocklogistics/api/v1/shipments/1", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mocklogistics/api/v1/shipments/:shipmentNo/outbound", requestPath: "/mocklogistics/api/v1/shipments/1/outbound", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mocklogistics/api/v1/shipments/:shipmentNo/advance", requestPath: "/mocklogistics/api/v1/shipments/1/advance", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mocklogistics/api/v1/shipments/:shipmentNo/complete", requestPath: "/mocklogistics/api/v1/shipments/1/complete", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/mocklogistics/api/v1/shipments/:shipmentNo/notifications", requestPath: "/mocklogistics/api/v1/shipments/1/notifications", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mocklogistics/api/v1/shipments/:shipmentNo/notifications/:notificationID/retry", requestPath: "/mocklogistics/api/v1/shipments/1/notifications/2/retry", authError: mockauth.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/mocklogistics/webhook", requestPath: "/mocklogistics/webhook", public: true, publicCode: http.StatusNoContent},
}

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	inventoryModule := inventorymodule.NewModule(nil)
	orderModule := ordermodule.NewModule(nil, inventoryModule.Reservation(), outbox.NewWriter())
	paymentModule, err := paymentmodule.NewModule(
		nil,
		orderModule.PaymentLifecycle(),
		orderModule.OrderService(),
		paymentmodule.Config{
			GatewayBaseURL:  "http://localhost:8080",
			WebhookSecret:   "test-signing-key",
			NotifyURL:       "http://localhost:8080/api/v1/payment/webhook",
			GatewayUsername: "mockpay-test",
			GatewayPassword: "mockpay-test-password",
		},
		outbox.NewWriter(),
	)
	if err != nil {
		t.Fatalf("create payment module: %v", err)
	}
	orderModule.SetPaymentCloser(paymentModule.PaymentCloser())
	logisticsModule, err := logisticsmodule.NewModule(
		nil,
		orderModule.ShippingLifecycle(),
		logisticsmodule.Config{
			GatewayBaseURL:  "http://localhost:8080",
			WebhookSecret:   "test-signing-key",
			NotifyURL:       "http://localhost:8080/api/v1/logistics/webhook",
			GatewayUsername: "mocklogistics-test",
			GatewayPassword: "mocklogistics-test-password",
		},
	)
	if err != nil {
		t.Fatalf("create logistics module: %v", err)
	}
	mockPaymentModule, err := mockpaymentmodule.NewModule(nil, mockpaymentmodule.Config{
		WebhookSecret:     "test-signing-key",
		NotifyAllowedURLs: "http://localhost:8080/api/v1/payment/webhook",
		Auth: mockauth.Config{
			Username: "mockpay-test",
			Password: "mockpay-test-password",
		},
	})
	if err != nil {
		t.Fatalf("create mock payment module: %v", err)
	}
	mockLogisticsModule, err := mocklogisticsmodule.NewModule(nil, mocklogisticsmodule.Config{
		WebhookSecret:     "test-signing-key",
		NotifyAllowedURLs: "http://localhost:8080/api/v1/logistics/webhook",
		Auth: mockauth.Config{
			Username: "mocklogistics-test",
			Password: "mocklogistics-test-password",
		},
	})
	if err != nil {
		t.Fatalf("create mock logistics module: %v", err)
	}
	r := gin.New()
	router.Register(r, router.Modules{
		User:              usermodule.NewModule(nil),
		Shop:              shopmodule.NewModule(nil),
		Product:           productmodule.NewModule(nil, inventoryModule.Lifecycle()),
		Inventory:         inventoryModule,
		Cart:              cartmodule.NewModule(nil),
		Address:           addressmodule.NewModule(nil),
		Feed:              feedmodule.NewModule(nil),
		Order:             orderModule,
		Payment:           paymentModule,
		Logistics:         logisticsModule,
		MockPayment:       mockPaymentModule,
		MockLogistics:     mockLogisticsModule,
		UserAuth:          usermodule.AuthMiddleware(invalidUserSession{}),
		ShopAuth:          shopmodule.AuthMiddleware(invalidShopSession{}),
		MockPaymentAuth:   invalidMockSessionMiddleware(),
		MockLogisticsAuth: invalidMockSessionMiddleware(),
	})
	return r
}

func invalidMockSessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("WWW-Authenticate", "Bearer")
		c.AbortWithStatusJSON(http.StatusUnauthorized, httpx.ErrorResponse{
			Error: mockauth.ErrInvalidSession.Error(),
		})
	}
}

func TestRegisteredRoutes(t *testing.T) {
	r := newTestRouter(t)
	got := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}

	want := make([]string, 0, len(routeCases))
	for _, route := range routeCases {
		want = append(want, route.method+" "+route.pattern)
	}
	slices.Sort(got)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Fatalf("routes = %v, want %v", got, want)
	}
}

func TestPublicAPIRouteSet(t *testing.T) {
	for _, route := range routeCases {
		if !strings.HasPrefix(route.pattern, "/api/") {
			continue
		}
		if got := router.IsPublicAPI(route.method, route.pattern); got != route.public {
			t.Errorf(
				"IsPublicAPI(%s, %s) = %v, want %v",
				route.method,
				route.pattern,
				got,
				route.public,
			)
		}
	}
}

func TestAuthenticationRoutesArePublic(t *testing.T) {
	r := newTestRouter(t)

	for _, route := range routeCases {
		if !route.public {
			continue
		}
		t.Run(route.pattern, func(t *testing.T) {
			rec := doRouterRequest(r, route.method, route.requestPath, "", `{bad`)

			want := route.publicCode
			if want == 0 {
				want = http.StatusBadRequest
			}
			if rec.Code != want {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
			}
			if route.publicCode != 0 {
				return
			}
			requireRouterError(t, rec, httpx.InvalidJSONMessage)
		})
	}
}

func TestProtectedRoutesRejectInvalidSessions(t *testing.T) {
	r := newTestRouter(t)

	for _, route := range routeCases {
		if route.authError == "" {
			continue
		}
		t.Run(route.method+" "+route.pattern, func(t *testing.T) {
			rec := doRouterRequest(r, route.method, route.requestPath, "Bearer invalid-token", "")

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
			}
			if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
				t.Fatalf("WWW-Authenticate = %q, want Bearer", got)
			}
			requireRouterError(t, rec, route.authError)
		})
	}
}

func TestMockLoginRoutesRejectRemoteClients(t *testing.T) {
	router := newTestRouter(t)
	for _, path := range []string{"/mockpay/api/v1/auth/login", "/mocklogistics/api/v1/auth/login"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{bad`))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
			}
		})
	}
}

func doRouterRequest(r http.Handler, method, path, authorization, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:1234"
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

func requireRouterError(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body httpx.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != want {
		t.Fatalf("error = %q, want %q", body.Error, want)
	}
}
