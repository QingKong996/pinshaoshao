package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pinshaoshao/internal/feature/cart/apperr"
	"pinshaoshao/internal/feature/cart/controller"
	"pinshaoshao/internal/feature/cart/dto"
	"pinshaoshao/internal/feature/cart/mocks"
	"pinshaoshao/internal/feature/cart/model"
	usermodule "pinshaoshao/internal/feature/user"
	usermodel "pinshaoshao/internal/feature/user/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

type stubUserAuthenticator struct {
	user *usermodel.User
}

func (s stubUserAuthenticator) Authenticate(context.Context, string) (*usermodel.User, error) {
	return s.user, nil
}

func newTestCartController(t *testing.T) (*controller.CartController, *mocks.MockCartService) {
	t.Helper()

	cartService := mocks.NewMockCartService(gomock.NewController(t))
	return controller.NewCartController(cartService), cartService
}

func newCartRouter(t *testing.T, cartController *controller.CartController, user *usermodel.User) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(usermodule.AuthMiddleware(stubUserAuthenticator{user: user}))
	protected.GET("/cart", cartController.List)
	protected.POST("/cart/items", cartController.Add)
	protected.PATCH("/cart/items/:itemID", cartController.UpdateQuantity)
	protected.DELETE("/cart/items/:itemID", cartController.Delete)
	return r
}

func newTestCartItem() model.CartItem {
	status := "on_shelf"
	enabled := true
	stock := 8
	return model.CartItem{
		ID:             1,
		UserID:         7,
		ProductID:      10,
		SKUID:          5,
		ProductName:    "T-shirt",
		Specs:          map[string]string{"color": "black"},
		UnitPriceCents: 1999,
		Quantity:       2,
		ProductStatus:  &status,
		SKUEnabled:     &enabled,
		AvailableStock: &stock,
	}
}

func TestCartList(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}
	cartController, cartService := newTestCartController(t)
	cartService.EXPECT().List(gomock.Any(), user.ID).Return([]model.CartItem{newTestCartItem()}, nil)

	rec := doRequest(newCartRouter(t, cartController, user), http.MethodGet, "/cart", "Bearer session-token", "")

	requireStatus(t, rec, http.StatusOK)

	var body []dto.CartItemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode cart response: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("item count = %d, want 1", len(body))
	}
	if body[0].ID != 1 || body[0].LineTotalCents != 3998 || body[0].Availability != "available" {
		t.Fatalf("item = %+v, want cart snapshot values", body[0])
	}
	if strings.Contains(rec.Body.String(), "available_stock") {
		t.Fatalf("cart response exposes available stock: %s", rec.Body.String())
	}
}

func TestCartAdd(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}

	t.Run("adds an item", func(t *testing.T) {
		cartController, cartService := newTestCartController(t)
		item := newTestCartItem()
		cartService.EXPECT().Add(gomock.Any(), user.ID, 5, 2).Return(&item, nil)

		rec := doRequest(
			newCartRouter(t, cartController, user),
			http.MethodPost,
			"/cart/items",
			"Bearer session-token",
			`{"sku_id":5,"quantity":2}`,
		)

		requireStatus(t, rec, http.StatusCreated)
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		cartController, _ := newTestCartController(t)
		rec := doRequest(
			newCartRouter(t, cartController, user),
			http.MethodPost,
			"/cart/items",
			"Bearer session-token",
			`{"sku_id":5,"quantity":0}`,
		)

		requireStatus(t, rec, http.StatusBadRequest)
	})

	t.Run("maps availability errors", func(t *testing.T) {
		tests := []struct {
			name       string
			err        error
			wantStatus int
			wantError  error
		}{
			{"missing product", apperr.ErrProductNotFound, http.StatusNotFound, apperr.ErrProductNotFound},
			{"unavailable product", apperr.ErrProductUnavailable, http.StatusConflict, apperr.ErrProductUnavailable},
			{"wrapped unavailable product", fmt.Errorf("service.AddCartItem: %w", apperr.ErrProductUnavailable), http.StatusConflict, apperr.ErrProductUnavailable},
			{"service failure", errors.New("database is down"), http.StatusInternalServerError, apperr.ErrInternalServer},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cartController, cartService := newTestCartController(t)
				cartService.EXPECT().Add(gomock.Any(), user.ID, 5, 2).Return(nil, tt.err)

				rec := doRequest(
					newCartRouter(t, cartController, user),
					http.MethodPost,
					"/cart/items",
					"Bearer session-token",
					`{"sku_id":5,"quantity":2}`,
				)

				requireStatus(t, rec, tt.wantStatus)
				requireError(t, rec, tt.wantError)
			})
		}
	})
}

func TestCartUpdateAndDelete(t *testing.T) {
	user := &usermodel.User{ID: 7, Username: "alice01", Status: usermodel.UserStatusActive}

	t.Run("updates quantity", func(t *testing.T) {
		cartController, cartService := newTestCartController(t)
		item := newTestCartItem()
		cartService.EXPECT().UpdateQuantity(gomock.Any(), user.ID, 1, 4).Return(&item, nil)

		rec := doRequest(
			newCartRouter(t, cartController, user),
			http.MethodPatch,
			"/cart/items/1",
			"Bearer session-token",
			`{"quantity":4}`,
		)

		requireStatus(t, rec, http.StatusOK)
	})

	t.Run("deletes an item", func(t *testing.T) {
		cartController, cartService := newTestCartController(t)
		cartService.EXPECT().Delete(gomock.Any(), user.ID, 1).Return(nil)

		rec := doRequest(
			newCartRouter(t, cartController, user),
			http.MethodDelete,
			"/cart/items/1",
			"Bearer session-token",
			"",
		)

		requireStatus(t, rec, http.StatusNoContent)
	})

	t.Run("maps a missing item", func(t *testing.T) {
		cartController, cartService := newTestCartController(t)
		cartService.EXPECT().Delete(gomock.Any(), user.ID, 1).Return(apperr.ErrCartItemNotFound)

		rec := doRequest(
			newCartRouter(t, cartController, user),
			http.MethodDelete,
			"/cart/items/1",
			"Bearer session-token",
			"",
		)

		requireStatus(t, rec, http.StatusNotFound)
		requireError(t, rec, apperr.ErrCartItemNotFound)
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

func requireStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func requireError(t *testing.T, rec *httptest.ResponseRecorder, want error) {
	t.Helper()

	var body dto.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != want.Error() {
		t.Fatalf("error = %q, want %q", body.Error, want.Error())
	}
}
