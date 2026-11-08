package router

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/controller"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func newTestProductSKURouter(t *testing.T) (http.Handler, *mocks.MockUserRepository, *mocks.MockProductReader, *mocks.MockProductSKURepository) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	userRepository := mocks.NewMockUserRepository(gomock.NewController(t))
	userService := service.NewUserService(userRepository)
	userController := controller.NewUserController(userService)
	authMiddleware := middleware.UserAuth(userService)

	productReader := mocks.NewMockProductReader(gomock.NewController(t))
	skuRepository := mocks.NewMockProductSKURepository(gomock.NewController(t))
	skuService := service.NewProductSKUService(productReader, skuRepository)
	skuController := controller.NewProductSKUController(skuService)

	r := gin.New()
	UserRegister(r, userController, authMiddleware)
	ProductSKURegister(r, skuController, authMiddleware)
	return r, userRepository, productReader, skuRepository
}

func TestProductSKURoutes(t *testing.T) {
	r, userRepository, productReader, skuRepository := newTestProductSKURouter(t)

	credentials := `{"username":"` + testUsername + `","password":"` + testPassword + `"}`
	passwordHash := mustHashPassword(t, testPassword)

	activeUser := &model.User{
		ID:           1,
		Username:     testUsername,
		PasswordHash: passwordHash,
		Status:       model.UserStatusActive,
	}
	product := &model.Product{ID: 1, Name: "T-shirt"}
	createdSKU := model.ProductSKU{
		ID:         1,
		ProductID:  1,
		Specs:      map[string]string{"color": "black", "size": "M"},
		PriceCents: 1999,
		Stock:      100,
		IsEnabled:  true,
	}

	userRepository.EXPECT().
		CreateUserWithSession(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user *model.User, _ []byte) error {
			user.ID = 1
			user.Status = model.UserStatusActive
			return nil
		})

	rec := doRequest(r, http.MethodPost, "/api/v1/register", "", credentials)
	requireStatus(t, rec, http.StatusCreated)
	token := responseToken(t, rec)

	skuBody := `{"specs":{"color":"black","size":"M"},"price_cents":1999,"stock":100}`

	t.Run("list is public and starts empty", func(t *testing.T) {
		productReader.EXPECT().
			GetProductByID(gomock.Any(), 1).
			Return(product, nil)
		skuRepository.EXPECT().
			ListProductSKUsByProductID(gomock.Any(), 1).
			Return([]model.ProductSKU{}, nil)

		rec := doRequest(r, http.MethodGet, "/api/v1/products/1/skus", "", "")
		requireStatus(t, rec, http.StatusOK)

		var skus []dto.ProductSKUResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &skus); err != nil {
			t.Fatalf("decode sku list: %v", err)
		}
		if len(skus) != 0 {
			t.Fatalf("sku count = %d, want 0", len(skus))
		}
	})

	t.Run("create requires authentication", func(t *testing.T) {
		requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/products/1/skus", "", skuBody), http.StatusUnauthorized)
	})

	t.Run("create rejects a malformed product id", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)

		rec := doRequest(r, http.MethodPost, "/api/v1/products/abc/skus", "Bearer "+token, skuBody)
		requireStatus(t, rec, http.StatusBadRequest)
		if got := responseError(t, rec); got != apperr.ErrInvalidProductID.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidProductID.Error())
		}
	})

	t.Run("list rejects a malformed product id", func(t *testing.T) {
		rec := doRequest(r, http.MethodGet, "/api/v1/products/abc/skus", "", "")
		requireStatus(t, rec, http.StatusBadRequest)
		if got := responseError(t, rec); got != apperr.ErrInvalidProductID.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrInvalidProductID.Error())
		}
	})

	t.Run("list rejects an unknown product", func(t *testing.T) {
		productReader.EXPECT().
			GetProductByID(gomock.Any(), 99).
			Return(nil, fmt.Errorf("repository.GetProductByID: %w", apperr.ErrProductNotFound))

		rec := doRequest(r, http.MethodGet, "/api/v1/products/99/skus", "", "")
		requireStatus(t, rec, http.StatusNotFound)
		if got := responseError(t, rec); got != apperr.ErrProductNotFound.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrProductNotFound.Error())
		}
	})

	t.Run("create rejects an unknown product", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)
		skuRepository.EXPECT().
			CreateProductSKU(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("repository.CreateProductSKU: %w", apperr.ErrProductNotFound))

		rec := doRequest(r, http.MethodPost, "/api/v1/products/99/skus", "Bearer "+token, skuBody)
		requireStatus(t, rec, http.StatusNotFound)
		if got := responseError(t, rec); got != apperr.ErrProductNotFound.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrProductNotFound.Error())
		}
	})

	t.Run("create rejects an invalid payload", func(t *testing.T) {
		bodies := map[string]string{
			"malformed json": `{bad`,
			"negative price": `{"price_cents":-1}`,
			"negative stock": `{"stock":-1}`,
			"stock overflow": `{"stock":3000000000}`,
		}
		for name, body := range bodies {
			t.Run(name, func(t *testing.T) {
				userRepository.EXPECT().
					GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
					Return(activeUser, nil)

				requireStatus(t, doRequest(r, http.MethodPost, "/api/v1/products/1/skus", "Bearer "+token, body), http.StatusBadRequest)
			})
		}
	})

	t.Run("create returns the new sku", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)
		skuRepository.EXPECT().
			CreateProductSKU(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, sku *model.ProductSKU) error {
				sku.ID = 1
				return nil
			})

		rec := doRequest(r, http.MethodPost, "/api/v1/products/1/skus", "Bearer "+token, skuBody)
		requireStatus(t, rec, http.StatusCreated)

		var sku dto.ProductSKUResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &sku); err != nil {
			t.Fatalf("decode sku: %v", err)
		}
		if sku.ID == 0 {
			t.Fatal("sku ID was not assigned")
		}
		if sku.ProductID != 1 {
			t.Fatalf("product ID = %d, want 1", sku.ProductID)
		}
		if !sku.IsEnabled {
			t.Fatal("sku is disabled, want enabled by default")
		}
		if sku.Specs["color"] != "black" || sku.Specs["size"] != "M" {
			t.Fatalf("specs = %v", sku.Specs)
		}
	})

	t.Run("create rejects specs keys that differ from existing skus", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)
		skuRepository.EXPECT().
			CreateProductSKU(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("repository.CreateProductSKU: %w", apperr.ErrProductSKUSpecsMismatch))

		body := `{"specs":{"color":"white"},"price_cents":1999,"stock":10}`
		rec := doRequest(r, http.MethodPost, "/api/v1/products/1/skus", "Bearer "+token, body)
		requireStatus(t, rec, http.StatusBadRequest)
		if got := responseError(t, rec); got != apperr.ErrProductSKUSpecsMismatch.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrProductSKUSpecsMismatch.Error())
		}
	})

	t.Run("duplicate specs conflict", func(t *testing.T) {
		userRepository.EXPECT().
			GetUserBySessionTokenHash(gomock.Any(), security.HashSessionToken(token)).
			Return(activeUser, nil)
		skuRepository.EXPECT().
			CreateProductSKU(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("repository.CreateProductSKU: %w", apperr.ErrProductSKUAlreadyExists))

		rec := doRequest(r, http.MethodPost, "/api/v1/products/1/skus", "Bearer "+token, skuBody)
		requireStatus(t, rec, http.StatusConflict)
		if got := responseError(t, rec); got != apperr.ErrProductSKUAlreadyExists.Error() {
			t.Fatalf("error = %q, want %q", got, apperr.ErrProductSKUAlreadyExists.Error())
		}
	})

	t.Run("list returns the created sku", func(t *testing.T) {
		productReader.EXPECT().
			GetProductByID(gomock.Any(), 1).
			Return(product, nil)
		skuRepository.EXPECT().
			ListProductSKUsByProductID(gomock.Any(), 1).
			Return([]model.ProductSKU{createdSKU}, nil)

		rec := doRequest(r, http.MethodGet, "/api/v1/products/1/skus", "", "")
		requireStatus(t, rec, http.StatusOK)

		var skus []dto.ProductSKUResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &skus); err != nil {
			t.Fatalf("decode sku list: %v", err)
		}
		if len(skus) != 1 {
			t.Fatalf("sku count = %d, want 1", len(skus))
		}
	})
}
