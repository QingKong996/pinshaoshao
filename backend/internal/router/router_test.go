package router_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/router"
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
)

type invalidUserSession struct{}

func (invalidUserSession) Authenticate(context.Context, string) (*model.User, error) {
	return nil, apperr.ErrInvalidSession
}

type invalidShopSession struct{}

func (invalidShopSession) Authenticate(context.Context, string) (*model.Shop, error) {
	return nil, apperr.ErrInvalidShopSession
}

type unusedUserService struct{}

func (unusedUserService) Login(context.Context, service.LoginInput) (string, error) {
	panic("unexpected user login")
}

func (unusedUserService) Register(context.Context, service.RegisterUserInput) (string, error) {
	panic("unexpected user registration")
}

func (unusedUserService) Logout(context.Context, string) error {
	panic("unexpected user logout")
}

type unusedShopService struct{}

func (unusedShopService) Login(context.Context, service.LoginInput) (string, error) {
	panic("unexpected shop login")
}

func (unusedShopService) Register(context.Context, service.RegisterShopInput) (string, error) {
	panic("unexpected shop registration")
}

func (unusedShopService) Logout(context.Context, string) error {
	panic("unexpected shop logout")
}

func (unusedShopService) GetProfile(context.Context, int) (*model.Shop, error) {
	panic("unexpected shop profile lookup")
}

func (unusedShopService) UpdateProfile(context.Context, int, service.UpdateShopProfileInput) (*model.Shop, error) {
	panic("unexpected shop profile update")
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
	{method: http.MethodPost, pattern: "/api/v1/logout", requestPath: "/api/v1/logout", authError: apperr.ErrInvalidSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/login", requestPath: "/api/v1/merchant/login", public: true},
	{method: http.MethodPost, pattern: "/api/v1/merchant/register", requestPath: "/api/v1/merchant/register", public: true},
	{method: http.MethodPost, pattern: "/api/v1/merchant/logout", requestPath: "/api/v1/merchant/logout", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/shop", requestPath: "/api/v1/merchant/shop", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/shop", requestPath: "/api/v1/merchant/shop", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/categories", requestPath: "/api/v1/merchant/categories", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/categories", requestPath: "/api/v1/merchant/categories", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/categories/:categoryID", requestPath: "/api/v1/merchant/categories/1", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/merchant/categories/:categoryID", requestPath: "/api/v1/merchant/categories/1", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/products", requestPath: "/api/v1/merchant/products", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/products", requestPath: "/api/v1/merchant/products", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/products/:productID", requestPath: "/api/v1/merchant/products/1", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/products/:productID", requestPath: "/api/v1/merchant/products/1", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/merchant/products/:productID", requestPath: "/api/v1/merchant/products/1", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/products/:productID/skus", requestPath: "/api/v1/merchant/products/1/skus", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPost, pattern: "/api/v1/merchant/products/:productID/skus", requestPath: "/api/v1/merchant/products/1/skus", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodPatch, pattern: "/api/v1/merchant/products/:productID/skus/:skuID", requestPath: "/api/v1/merchant/products/1/skus/2", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodDelete, pattern: "/api/v1/merchant/products/:productID/skus/:skuID", requestPath: "/api/v1/merchant/products/1/skus/2", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/inventory", requestPath: "/api/v1/merchant/inventory", authError: apperr.ErrInvalidShopSession.Error()},
	{method: http.MethodGet, pattern: "/api/v1/merchant/dashboard", requestPath: "/api/v1/merchant/dashboard", authError: apperr.ErrInvalidShopSession.Error()},
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	userController := controller.NewUserController(unusedUserService{})
	shopController := controller.NewShopController(unusedShopService{})
	categoryController := controller.NewProductCategoryController(nil)
	productController := controller.NewProductController(nil)
	skuController := controller.NewProductSKUController(nil)
	inventoryController := controller.NewInventoryController(nil)
	dashboardController := controller.NewDashboardController(nil)

	router.UserRegister(r, userController, middleware.UserAuth(invalidUserSession{}))
	router.ShopRegister(r, shopController, middleware.ShopAuth(invalidShopSession{}))
	router.MerchantRegister(
		r,
		shopController,
		categoryController,
		productController,
		skuController,
		inventoryController,
		dashboardController,
		middleware.ShopAuth(invalidShopSession{}),
	)
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
			requireRouterError(t, rec, apperr.ErrInvalidJsonFormat.Error())
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

	var body dto.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != want {
		t.Fatalf("error = %q, want %q", body.Error, want)
	}
}
