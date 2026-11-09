package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

const (
	testShopUsername = "shop01"
	testShopEmail    = "shop01@example.com"
)

func newTestShopController(t *testing.T) (*controller.ShopController, *mocks.MockShopService) {
	t.Helper()

	shopService := mocks.NewMockShopService(gomock.NewController(t))
	return controller.NewShopController(shopService), shopService
}

func newShopLogoutRouter(t *testing.T, shopController *controller.ShopController, shop *model.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/logout", middleware.ShopAuth(stubShopAuthenticator{shop: shop}), shopController.Logout)
	return r
}

func newShopProfileRouter(t *testing.T, shopController *controller.ShopController, shop *model.Shop) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	protected := r.Group("")
	protected.Use(middleware.ShopAuth(stubShopAuthenticator{shop: shop}))
	protected.GET("/shop", shopController.GetProfile)
	protected.PATCH("/shop", shopController.UpdateProfile)
	return r
}

func TestShopLogin(t *testing.T) {
	credentials := fmt.Sprintf(`{"username":%q,"password":%q}`, testShopUsername, testPassword)

	t.Run("returns a session token", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("session-token", nil)

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		shopController.Login(ctx)

		requireStatus(t, rec, http.StatusOK)
		requireToken(t, rec)
	})

	t.Run("rejects invalid credentials", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrInvalidCredentials))

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		shopController.Login(ctx)

		requireStatus(t, rec, http.StatusUnauthorized)
		requireError(t, rec, apperr.ErrInvalidCredentials)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		shopController, _ := newTestShopController(t)

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", `{bad`)
		shopController.Login(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects a banned shop", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrShopBanned))

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		shopController.Login(ctx)

		requireStatus(t, rec, http.StatusForbidden)
		requireError(t, rec, apperr.ErrShopBanned)
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Login(gomock.Any(), gomock.Any()).
			Return("", errors.New("database is down"))

		ctx, rec := newControllerContext(t, http.MethodPost, "/login", credentials)
		shopController.Login(ctx)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestShopRegister(t *testing.T) {
	registration := fmt.Sprintf(`{"username":%q,"name":"Test Shop","password":%q,"email":%q}`, testShopUsername, testPassword, testShopEmail)

	t.Run("returns a session token", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("session-token", nil)

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusCreated)
		requireToken(t, rec)
	})

	t.Run("rejects a taken username", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrShopUsernameTaken))

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusConflict)
		requireError(t, rec, apperr.ErrShopUsernameTaken)
	})

	t.Run("rejects a taken email", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("", fmt.Errorf("service: %w", apperr.ErrShopEmailTaken))

		body := fmt.Sprintf(`{"username":"shop02","name":"Test Shop","password":%q,"email":%q}`, testPassword, testShopEmail)
		ctx, rec := newControllerContext(t, http.MethodPost, "/register", body)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusConflict)
		requireError(t, rec, apperr.ErrShopEmailTaken)
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		shopController, _ := newTestShopController(t)

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", `{bad`)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid email", func(t *testing.T) {
		shopController, _ := newTestShopController(t)

		body := fmt.Sprintf(`{"username":%q,"name":"Test Shop","password":%q,"email":"not-an-email"}`, testShopUsername, testPassword)
		ctx, rec := newControllerContext(t, http.MethodPost, "/register", body)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		shopController, _ := newTestShopController(t)

		body := fmt.Sprintf(`{"username":%q,"name":"Test Shop","password":%q,"email":%q}`, "sh", testPassword, testShopEmail)
		ctx, rec := newControllerContext(t, http.MethodPost, "/register", body)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects a missing shop name", func(t *testing.T) {
		shopController, _ := newTestShopController(t)

		body := fmt.Sprintf(`{"username":%q,"password":%q,"email":%q}`, testShopUsername, testPassword, testShopEmail)
		ctx, rec := newControllerContext(t, http.MethodPost, "/register", body)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects a one-character shop name", func(t *testing.T) {
		shopController, _ := newTestShopController(t)

		body := fmt.Sprintf(`{"username":%q,"name":"A","password":%q,"email":%q}`, testShopUsername, testPassword, testShopEmail)
		ctx, rec := newControllerContext(t, http.MethodPost, "/register", body)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Register(gomock.Any(), gomock.Any()).
			Return("", errors.New("database is down"))

		ctx, rec := newControllerContext(t, http.MethodPost, "/register", registration)
		shopController.Register(ctx)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestShopLogout(t *testing.T) {
	activeShop := &model.Shop{ID: 1, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("revokes the session", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Logout(gomock.Any(), "session-token").
			Return(nil)

		rec := doRequest(newShopLogoutRouter(t, shopController, activeShop), http.MethodPost, "/logout", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusNoContent)
		if rec.Body.Len() != 0 {
			t.Fatalf("body = %q, want empty", rec.Body.String())
		}
	})

	t.Run("returns 500 when the session cannot be revoked", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			Logout(gomock.Any(), "session-token").
			Return(errors.New("database is down"))

		rec := doRequest(newShopLogoutRouter(t, shopController, activeShop), http.MethodPost, "/logout", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestShopGetProfile(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}

	t.Run("returns the shop profile", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			GetProfile(gomock.Any(), activeShop.ID).
			Return(&model.Shop{
				ID:          activeShop.ID,
				Username:    testShopUsername,
				Name:        "First shop",
				Description: "Intro",
				Email:       testShopEmail,
				Status:      model.ShopStatusActive,
			}, nil)

		rec := doRequest(newShopProfileRouter(t, shopController, activeShop), http.MethodGet, "/shop", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusOK)

		var body dto.ShopProfileResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode shop profile: %v", err)
		}
		if body.ID != activeShop.ID || body.Username != testShopUsername {
			t.Fatalf("profile = %+v, want shop %d", body, activeShop.ID)
		}
		if body.Name != "First shop" || body.Description != "Intro" {
			t.Fatalf("profile = %+v, want profile fields", body)
		}
		if body.Status != string(model.ShopStatusActive) {
			t.Fatalf("status = %q, want %q", body.Status, model.ShopStatusActive)
		}
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			GetProfile(gomock.Any(), activeShop.ID).
			Return(nil, errors.New("database is down"))

		rec := doRequest(newShopProfileRouter(t, shopController, activeShop), http.MethodGet, "/shop", "Bearer session-token", "")

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}

func TestShopUpdateProfile(t *testing.T) {
	activeShop := &model.Shop{ID: 7, Username: testShopUsername, Status: model.ShopStatusActive}
	validBody := `{"name":"First shop","description":"Intro"}`

	t.Run("returns the updated profile", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			UpdateProfile(gomock.Any(), activeShop.ID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int, req *dto.UpdateShopProfileRequest) (*model.Shop, error) {
				if req.Description == nil || *req.Description != "Intro" {
					t.Fatalf("description = %v, want %q", req.Description, "Intro")
				}
				return &model.Shop{
					ID:          activeShop.ID,
					Username:    testShopUsername,
					Name:        "First shop",
					Description: *req.Description,
					Status:      model.ShopStatusActive,
				}, nil
			})

		rec := doRequest(newShopProfileRouter(t, shopController, activeShop), http.MethodPatch, "/shop", "Bearer session-token", validBody)

		requireStatus(t, rec, http.StatusOK)

		var body dto.ShopProfileResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode shop profile: %v", err)
		}
		if body.Name != "First shop" || body.Description != "Intro" {
			t.Fatalf("profile = %+v, want updated fields", body)
		}
	})

	t.Run("rejects malformed json", func(t *testing.T) {
		shopController, _ := newTestShopController(t)

		rec := doRequest(newShopProfileRouter(t, shopController, activeShop), http.MethodPatch, "/shop", "Bearer session-token", `{bad`)

		requireStatus(t, rec, http.StatusBadRequest)
		requireError(t, rec, apperr.ErrInvalidJsonFormat)
	})

	t.Run("rejects an invalid payload", func(t *testing.T) {
		payloads := map[string]string{
			"missing name":         `{"description":"Intro"}`,
			"missing description":  `{"name":"First shop"}`,
			"empty name":           `{"name":"","description":"Intro"}`,
			"name too long":        fmt.Sprintf(`{"name":%q,"description":"Intro"}`, strings.Repeat("n", 51)),
			"description too long": fmt.Sprintf(`{"name":"First shop","description":%q}`, strings.Repeat("d", 501)),
		}
		for name, payload := range payloads {
			t.Run(name, func(t *testing.T) {
				shopController, _ := newTestShopController(t)

				rec := doRequest(newShopProfileRouter(t, shopController, activeShop), http.MethodPatch, "/shop", "Bearer session-token", payload)

				requireStatus(t, rec, http.StatusBadRequest)
				requireError(t, rec, apperr.ErrInvalidJsonFormat)
			})
		}
	})

	t.Run("returns 500 on service failure", func(t *testing.T) {
		shopController, shopService := newTestShopController(t)
		shopService.EXPECT().
			UpdateProfile(gomock.Any(), activeShop.ID, gomock.Any()).
			Return(nil, errors.New("database is down"))

		rec := doRequest(newShopProfileRouter(t, shopController, activeShop), http.MethodPatch, "/shop", "Bearer session-token", validBody)

		requireStatus(t, rec, http.StatusInternalServerError)
		requireError(t, rec, apperr.ErrInternalServer)
	})
}
