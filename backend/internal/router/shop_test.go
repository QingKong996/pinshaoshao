package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

const (
	testShopUsername = "shop01"
	testShopPassword = "Password123"
	testShopEmail    = "shop01@example.com"
)

func newTestShopRouter(t *testing.T) (http.Handler, *mocks.MockShopRepository) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	shopRepository := mocks.NewMockShopRepository(gomock.NewController(t))
	shopService := service.NewShopService(shopRepository)
	shopController := controller.NewShopController(shopService)
	shopAuthMiddleware := middleware.ShopAuth(shopService)

	r := gin.New()
	ShopRegister(r, shopController, shopAuthMiddleware)

	// Test-only route to observe the shop injected by ShopAuth.
	r.GET("/api/v1/merchant/whoami", middleware.ShopAuth(shopService), func(c *gin.Context) {
		shop, ok := middleware.CurrentShop(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "shop missing from context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"username": shop.Username})
	})

	return r, shopRepository
}

func TestShopRoutes(t *testing.T) {
	r, shopRepository := newTestShopRouter(t)
	credentials := `{"username":"` + testShopUsername + `","password":"` + testShopPassword + `"}`
	registration := `{"username":"` + testShopUsername + `","password":"` + testShopPassword + `","email":"` + testShopEmail + `"}`
	passwordHash := mustHashPassword(t, testShopPassword)

	activeShop := &model.Shop{
		ID:           1,
		Username:     testShopUsername,
		PasswordHash: passwordHash,
		Email:        testShopEmail,
		Status:       model.ShopStatusActive,
	}
	bannedShop := &model.Shop{
		ID:           1,
		Username:     testShopUsername,
		PasswordHash: passwordHash,
		Email:        testShopEmail,
		Status:       model.ShopStatusBanned,
	}

	var token string

	t.Run("register issues a session token", func(t *testing.T) {
		shopRepository.EXPECT().
			CreateShopWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, shop *model.Shop, _ []byte) error {
				shop.ID = 1
				shop.Status = model.ShopStatusActive
				return nil
			})

		rec := doRequest(r, http.MethodPost, "/api/v1/merchant/register", "", registration)
		requireStatus(t, rec, http.StatusCreated)
		token = responseToken(t, rec)
	})

	t.Run("register rejects a duplicate username", func(t *testing.T) {
		shopRepository.EXPECT().
			CreateShopWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("repository.CreateShop: %w", apperr.ErrShopUsernameTaken))

		rec := doRequest(r, http.MethodPost, "/api/v1/merchant/register", "", registration)
		requireStatus(t, rec, http.StatusConflict)
		if got := responseError(t, rec); got != apperr.ErrShopUsernameTaken.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrShopUsernameTaken.Error())
		}
	})

	t.Run("register rejects a duplicate email", func(t *testing.T) {
		shopRepository.EXPECT().
			CreateShopWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("repository.CreateShop: %w", apperr.ErrShopEmailTaken))

		body := `{"username":"shop02","password":"` + testShopPassword + `","email":"` + testShopEmail + `"}`
		rec := doRequest(r, http.MethodPost, "/api/v1/merchant/register", "", body)
		requireStatus(t, rec, http.StatusConflict)
		if got := responseError(t, rec); got != apperr.ErrShopEmailTaken.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrShopEmailTaken.Error())
		}
	})

	t.Run("register rejects an invalid email", func(t *testing.T) {
		body := `{"username":"shop02","password":"` + testShopPassword + `","email":"not-an-email"}`
		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/merchant/register", "", body), http.StatusBadRequest)
	})

	t.Run("middleware rejects missing credentials", func(t *testing.T) {
		rec := doRequest(r, http.MethodPost, "/api/v1/merchant/logout", "", "")
		requireStatus(t, rec, http.StatusUnauthorized)
		if got := responseError(t, rec); got != apperr.ErrUnauthorized.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrUnauthorized.Error())
		}
		if got := rec.Header().Get("WWW-Authenticate"); got != "Bearer" {
			t.Fatalf("WWW-Authenticate = %q, want %q", got, "Bearer")
		}
	})

	t.Run("middleware rejects non-bearer scheme", func(t *testing.T) {
		requireStatus(t, doRequest(r, http.MethodGet, "/api/v1/merchant/whoami", "Basic abc", ""), http.StatusUnauthorized)
	})

	t.Run("middleware rejects unknown token", func(t *testing.T) {
		shopRepository.EXPECT().
			GetShopBySessionTokenHash(gomock.Any(), security.HashSessionToken("not-a-real-token")).
			Return(nil, fmt.Errorf("repository.GetShopBySessionTokenHash: %w", apperr.ErrInvalidShopSession))

		rec := doRequest(r, http.MethodGet, "/api/v1/merchant/whoami", "Bearer not-a-real-token", "")
		requireStatus(t, rec, http.StatusUnauthorized)
		if got := responseError(t, rec); got != apperr.ErrInvalidShopSession.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidShopSession.Error())
		}
	})

	t.Run("middleware injects shop for valid token", func(t *testing.T) {
		shopRepository.EXPECT().
			GetShopBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeShop, nil)

		rec := doRequest(r, http.MethodGet, "/api/v1/merchant/whoami", "Bearer "+token, "")
		requireStatus(t, rec, http.StatusOK)

		var whoami struct {
			Username string `json:"username"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &whoami); err != nil {
			t.Fatalf("decode whoami response: %v", err)
		}
		if whoami.Username != testShopUsername {
			t.Fatalf("username = %q, want %q", whoami.Username, testShopUsername)
		}
	})

	t.Run("logout revokes the session token", func(t *testing.T) {
		shopRepository.EXPECT().
			GetShopBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeShop, nil)
		shopRepository.EXPECT().
			DeleteSessionByTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(nil)
		shopRepository.EXPECT().
			GetShopBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(nil, fmt.Errorf("repository.GetShopBySessionTokenHash: %w", apperr.ErrInvalidShopSession))

		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/merchant/logout", "Bearer "+token, ""), http.StatusNoContent)
		requireStatus(t, doRequest(r, http.MethodGet, "/api/v1/merchant/whoami", "Bearer "+token, ""), http.StatusUnauthorized)
	})

	t.Run("login issues a new session token", func(t *testing.T) {
		shopRepository.EXPECT().
			GetShopByUsername(gomock.Any(), testShopUsername).
			Return(activeShop, nil)
		shopRepository.EXPECT().
			CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		rec := doRequest(r, http.MethodPost, "/api/v1/merchant/login", "", credentials)
		requireStatus(t, rec, http.StatusOK)
		token = responseToken(t, rec)

		shopRepository.EXPECT().
			GetShopBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeShop, nil)
		requireStatus(t, doRequest(r, http.MethodGet, "/api/v1/merchant/whoami", "Bearer "+token, ""), http.StatusOK)
	})

	t.Run("middleware rejects banned shop", func(t *testing.T) {
		shopRepository.EXPECT().
			GetShopBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(bannedShop, nil).
			Times(2)

		rec := doRequest(r, http.MethodGet, "/api/v1/merchant/whoami", "Bearer "+token, "")
		requireStatus(t, rec, http.StatusForbidden)
		if got := responseError(t, rec); got != apperr.ErrShopBanned.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrShopBanned.Error())
		}
		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/merchant/logout", "Bearer "+token, ""), http.StatusForbidden)
	})
}

func TestMerchantAndUserTokensAreNotInterchangeable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userRepository := mocks.NewMockUserRepository(gomock.NewController(t))
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)

	shopRepository := mocks.NewMockShopRepository(gomock.NewController(t))
	shopService := service.NewShopService(shopRepository)
	shopController := controller.NewShopController(shopService)

	r := gin.New()
	UserRegister(r, userController, middleware.UserAuth(userService))
	ShopRegister(r, shopController, middleware.ShopAuth(shopService))

	userRepository.EXPECT().
		CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *model.User, _ []byte) error {
			user.ID = 1
			user.Status = model.UserStatusActive
			return nil
		})

	rec := doRequest(r, http.MethodPost, "/api/v1/register", "", `{"username":"alice01","password":"Password123"}`)
	requireStatus(t, rec, http.StatusCreated)
	userToken := responseToken(t, rec)

	shopRepository.EXPECT().
		GetShopBySessionTokenHash(gomock.Any(), security.HashSessionToken(userToken)).
		Return(nil, fmt.Errorf("repository.GetShopBySessionTokenHash: %w", apperr.ErrInvalidShopSession))

	rec = doRequest(r, http.MethodPost, "/api/v1/merchant/logout", "Bearer "+userToken, "")
	requireStatus(t, rec, http.StatusUnauthorized)
	if got := responseError(t, rec); got != apperr.ErrInvalidShopSession.Error() {
		t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidShopSession.Error())
	}

	shopRepository.EXPECT().
		CreateShopWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, shop *model.Shop, _ []byte) error {
			shop.ID = 1
			shop.Status = model.ShopStatusActive
			return nil
		})

	rec = doRequest(r, http.MethodPost, "/api/v1/merchant/register", "", `{"username":"shop02","password":"Password123","email":"shop02@example.com"}`)
	requireStatus(t, rec, http.StatusCreated)
	shopToken := responseToken(t, rec)

	userRepository.EXPECT().
		GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(shopToken)).
		Return(nil, fmt.Errorf("repository.GetUserBySessionTokenHash: %w", apperr.ErrInvalidSession))

	rec = doRequest(r, http.MethodPost, "/api/v1/logout", "Bearer "+shopToken, "")
	requireStatus(t, rec, http.StatusUnauthorized)
	if got := responseError(t, rec); got != apperr.ErrInvalidSession.Error() {
		t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidSession.Error())
	}
}
