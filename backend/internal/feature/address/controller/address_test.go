package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pinshaoshao/internal/feature/address/apperr"
	"pinshaoshao/internal/feature/address/controller"
	"pinshaoshao/internal/feature/address/dto"
	"pinshaoshao/internal/feature/address/model"
	"pinshaoshao/internal/feature/address/service"
	usermodule "pinshaoshao/internal/feature/user"
	usermodel "pinshaoshao/internal/feature/user/model"

	"github.com/gin-gonic/gin"
)

type stubAddressService struct {
	listFn       func(context.Context, int) ([]model.Address, error)
	createFn     func(context.Context, int, service.AddressInput) (*model.Address, error)
	updateFn     func(context.Context, int, int, service.AddressInput) (*model.Address, error)
	deleteFn     func(context.Context, int, int) error
	setDefaultFn func(context.Context, int, int) (*model.Address, error)
}

func (s stubAddressService) List(ctx context.Context, userID int) ([]model.Address, error) {
	return s.listFn(ctx, userID)
}

func (s stubAddressService) Create(ctx context.Context, userID int, input service.AddressInput) (*model.Address, error) {
	return s.createFn(ctx, userID, input)
}

func (s stubAddressService) Update(ctx context.Context, userID int, addressID int, input service.AddressInput) (*model.Address, error) {
	return s.updateFn(ctx, userID, addressID, input)
}

func (s stubAddressService) Delete(ctx context.Context, userID int, addressID int) error {
	return s.deleteFn(ctx, userID, addressID)
}

func (s stubAddressService) SetDefault(ctx context.Context, userID int, addressID int) (*model.Address, error) {
	return s.setDefaultFn(ctx, userID, addressID)
}

type stubAddressUserAuthenticator struct {
	user *usermodel.User
}

func (s stubAddressUserAuthenticator) Authenticate(context.Context, string) (*usermodel.User, error) {
	return s.user, nil
}

func TestAddressControllerListAndCreate(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}
	address := testAddress()

	t.Run("lists addresses", func(t *testing.T) {
		addressController := controller.NewAddressController(stubAddressService{
			listFn: func(_ context.Context, userID int) ([]model.Address, error) {
				if userID != user.ID {
					t.Fatalf("user id = %d, want %d", userID, user.ID)
				}
				return []model.Address{address}, nil
			},
		})

		rec := doAddressRequest(
			newAddressRouter(t, addressController, user),
			http.MethodGet,
			"/addresses",
			"",
		)

		requireAddressStatus(t, rec, http.StatusOK)
		var body []dto.AddressResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode addresses: %v", err)
		}
		if len(body) != 1 || body[0].ReceiverName != address.ReceiverName || !body[0].IsDefault {
			t.Fatalf("addresses = %+v, want saved address", body)
		}
	})

	t.Run("creates an address", func(t *testing.T) {
		addressController := controller.NewAddressController(stubAddressService{
			createFn: func(_ context.Context, userID int, input service.AddressInput) (*model.Address, error) {
				if userID != user.ID ||
					input.ReceiverName != "Alice" ||
					input.ReceiverPhone != "13800138000" ||
					input.ShippingAddress != "Shanghai" ||
					!input.IsDefault {
					t.Fatalf("create args = %d/%+v", userID, input)
				}
				return &address, nil
			},
		})

		rec := doAddressRequest(
			newAddressRouter(t, addressController, user),
			http.MethodPost,
			"/addresses",
			`{"receiver_name":"Alice","receiver_phone":"13800138000","shipping_address":"Shanghai","is_default":true}`,
		)

		requireAddressStatus(t, rec, http.StatusCreated)
	})
}

func TestAddressControllerUpdateDeleteAndDefault(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}
	address := testAddress()

	t.Run("updates an address", func(t *testing.T) {
		addressController := controller.NewAddressController(stubAddressService{
			updateFn: func(_ context.Context, userID int, addressID int, input service.AddressInput) (*model.Address, error) {
				if userID != user.ID || addressID != address.ID || input.IsDefault {
					t.Fatalf("update args = %d/%d/%+v", userID, addressID, input)
				}
				return &address, nil
			},
		})

		rec := doAddressRequest(
			newAddressRouter(t, addressController, user),
			http.MethodPut,
			"/addresses/1",
			`{"receiver_name":"Alice","receiver_phone":"123","shipping_address":"Shanghai","is_default":false}`,
		)

		requireAddressStatus(t, rec, http.StatusOK)
	})

	t.Run("sets the default address", func(t *testing.T) {
		addressController := controller.NewAddressController(stubAddressService{
			setDefaultFn: func(context.Context, int, int) (*model.Address, error) {
				return &address, nil
			},
		})

		rec := doAddressRequest(
			newAddressRouter(t, addressController, user),
			http.MethodPost,
			"/addresses/1/default",
			"",
		)

		requireAddressStatus(t, rec, http.StatusOK)
	})

	t.Run("deletes an address", func(t *testing.T) {
		addressController := controller.NewAddressController(stubAddressService{
			deleteFn: func(context.Context, int, int) error {
				return nil
			},
		})

		rec := doAddressRequest(
			newAddressRouter(t, addressController, user),
			http.MethodDelete,
			"/addresses/1",
			"",
		)

		requireAddressStatus(t, rec, http.StatusNoContent)
	})

	t.Run("maps a missing address", func(t *testing.T) {
		addressController := controller.NewAddressController(stubAddressService{
			deleteFn: func(context.Context, int, int) error {
				return apperr.ErrAddressNotFound
			},
		})

		rec := doAddressRequest(
			newAddressRouter(t, addressController, user),
			http.MethodDelete,
			"/addresses/1",
			"",
		)

		requireAddressStatus(t, rec, http.StatusNotFound)
	})
}

func newAddressRouter(t *testing.T, addressController *controller.AddressController, user *usermodel.User) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(usermodule.AuthMiddleware(stubAddressUserAuthenticator{user: user}))
	protected.GET("/addresses", addressController.List)
	protected.POST("/addresses", addressController.Create)
	protected.PUT("/addresses/:addressID", addressController.Update)
	protected.DELETE("/addresses/:addressID", addressController.Delete)
	protected.POST("/addresses/:addressID/default", addressController.SetDefault)
	return r
}

func testAddress() model.Address {
	return model.Address{
		ID:              1,
		UserID:          7,
		ReceiverName:    "Alice",
		ReceiverPhone:   "13800138000",
		ShippingAddress: "Shanghai",
		IsDefault:       true,
		CreatedAt:       time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
	}
}

func doAddressRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer session-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func requireAddressStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}
