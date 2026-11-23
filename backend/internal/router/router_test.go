package router_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	cartmodule "pinshaoshao/internal/feature/cart"
	feedmodule "pinshaoshao/internal/feature/feed"
	inventorymodule "pinshaoshao/internal/feature/inventory"
	ordermodule "pinshaoshao/internal/feature/order"
	productmodule "pinshaoshao/internal/feature/product"
	shopmodule "pinshaoshao/internal/feature/shop"
	shopapperr "pinshaoshao/internal/feature/shop/apperr"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	usermodule "pinshaoshao/internal/feature/user"
	userapperr "pinshaoshao/internal/feature/user/apperr"
	usermodel "pinshaoshao/internal/feature/user/model"
	"pinshaoshao/internal/http/httpx"
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
	{method: http.MethodPost, pattern: "/api/v1/orders", requestPath: "/api/v1/orders", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/orders/direct", requestPath: "/api/v1/orders/direct", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/orders", requestPath: "/api/v1/orders", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/orders/:orderID", requestPath: "/api/v1/orders/1", authError: userapperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/orders/:orderID/cancel", requestPath: "/api/v1/orders/1/cancel", authError: userapperr.ErrInvalidSession.Error()},
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
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	inventoryModule := inventorymodule.NewModule(nil)
	r := gin.New()
	router.Register(r, router.Modules{
		User:      usermodule.NewModule(nil),
		Shop:      shopmodule.NewModule(nil),
		Product:   productmodule.NewModule(nil, inventoryModule.Lifecycle()),
		Inventory: inventoryModule,
		Cart:      cartmodule.NewModule(nil),
		Feed:      feedmodule.NewModule(nil),
		Order:     ordermodule.NewModule(nil, inventoryModule.Reservation()),
		UserAuth:  usermodule.AuthMiddleware(invalidUserSession{}),
		ShopAuth:  shopmodule.AuthMiddleware(invalidShopSession{}),
	})
	return r
}

func TestRegisteredRoutes(t *testing.T) {
	r := newTestRouter()
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

func TestAuthenticationRoutesArePublic(t *testing.T) {
	r := newTestRouter()

	for _, route := range routeCases {
		if !route.public {
			continue
		}
		t.Run(route.pattern, func(t *testing.T) {
			rec := doRouterRequest(r, route.method, route.requestPath, "", `{bad`)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
			}
			requireRouterError(t, rec, httpx.InvalidJSONMessage)
		})
	}
}

func TestProtectedRoutesRejectInvalidSessions(t *testing.T) {
	r := newTestRouter()

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

func doRouterRequest(r http.Handler, method, path, authorization, body string) *httptest.ResponseRecorder {
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
