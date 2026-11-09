package service_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
	"pinshaoshao/internal/service"

	"go.uber.org/mock/gomock"
)

const (
	testShopUsername = "shop01"
	testShopPassword = "Password123"
	testShopEmail    = "shop01@example.com"
)

func newTestShopService(t *testing.T) (*service.ShopService, *mocks.MockShopRepository) {
	t.Helper()

	shopRepository := mocks.NewMockShopRepository(gomock.NewController(t))
	return service.NewShopService(shopRepository), shopRepository
}

func TestShopAuthenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("valid session", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		token := "session-token"
		shop := &model.Shop{ID: 1, Username: testShopUsername, Status: model.ShopStatusActive}

		shopRepository.EXPECT().
			GetShopBySessionTokenHash(ctx, security.HashSessionToken(token)).
			Return(shop, nil)

		got, err := shopService.Authenticate(ctx, token)
		if err != nil {
			t.Fatalf("authenticate: %v", err)
		}
		if got.Username != testShopUsername {
			t.Fatalf("username = %q, want %q", got.Username, testShopUsername)
		}
	})

	t.Run("unknown token", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		token := "not-a-real-token"

		shopRepository.EXPECT().
			GetShopBySessionTokenHash(ctx, security.HashSessionToken(token)).
			Return(nil, fmt.Errorf("shop repository: %w", apperr.ErrInvalidShopSession))

		_, err := shopService.Authenticate(ctx, token)
		requireErrorIs(t, err, apperr.ErrInvalidShopSession)
	})

	t.Run("empty token", func(t *testing.T) {
		shopService, _ := newTestShopService(t)

		_, err := shopService.Authenticate(ctx, "")
		requireErrorIs(t, err, apperr.ErrUnauthorized)
	})

	t.Run("banned shop", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		token := "session-token"
		shop := &model.Shop{ID: 1, Username: testShopUsername, Status: model.ShopStatusBanned}

		shopRepository.EXPECT().
			GetShopBySessionTokenHash(ctx, security.HashSessionToken(token)).
			Return(shop, nil)

		_, err := shopService.Authenticate(ctx, token)
		requireErrorIs(t, err, apperr.ErrShopBanned)
	})
}

func TestShopLogout(t *testing.T) {
	ctx := context.Background()

	t.Run("revokes session", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		token := "session-token"

		shopRepository.EXPECT().
			DeleteSessionByTokenHash(ctx, security.HashSessionToken(token)).
			Return(nil)

		if err := shopService.Logout(ctx, token); err != nil {
			t.Fatalf("logout: %v", err)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		token := "session-token"

		shopRepository.EXPECT().
			DeleteSessionByTokenHash(ctx, security.HashSessionToken(token)).
			Return(nil).
			Times(2)

		if err := shopService.Logout(ctx, token); err != nil {
			t.Fatalf("first logout: %v", err)
		}
		if err := shopService.Logout(ctx, token); err != nil {
			t.Fatalf("second logout: %v", err)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		shopService, _ := newTestShopService(t)

		requireErrorIs(t, shopService.Logout(ctx, ""), apperr.ErrUnauthorized)
	})
}

func TestShopLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("valid credentials", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		shop := &model.Shop{
			ID:           1,
			Username:     testShopUsername,
			PasswordHash: mustHashPassword(t, testShopPassword),
			Status:       model.ShopStatusActive,
		}

		shopRepository.EXPECT().
			GetShopByUsername(ctx, testShopUsername).
			Return(shop, nil)

		var sessionTokenHash []byte
		shopRepository.EXPECT().
			CreateSession(ctx, shop, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *model.Shop, tokenHash []byte) error {
				sessionTokenHash = tokenHash
				return nil
			})

		token, err := shopService.Login(ctx, &dto.LoginRequest{
			Username: testShopUsername,
			Password: testShopPassword,
		})
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		if !bytes.Equal(sessionTokenHash, security.HashSessionToken(token)) {
			t.Fatal("stored session token hash does not match the issued token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		shop := &model.Shop{
			ID:           1,
			Username:     testShopUsername,
			PasswordHash: mustHashPassword(t, testShopPassword),
			Status:       model.ShopStatusActive,
		}

		shopRepository.EXPECT().
			GetShopByUsername(ctx, testShopUsername).
			Return(shop, nil)

		_, err := shopService.Login(ctx, &dto.LoginRequest{
			Username: testShopUsername,
			Password: "WrongPassword123",
		})
		requireErrorIs(t, err, apperr.ErrInvalidCredentials)
	})

	t.Run("unknown shop", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)

		shopRepository.EXPECT().
			GetShopByUsername(ctx, testShopUsername).
			Return(nil, fmt.Errorf("shop repository: %w", apperr.ErrShopNotFound))

		_, err := shopService.Login(ctx, &dto.LoginRequest{
			Username: testShopUsername,
			Password: testShopPassword,
		})
		requireErrorIs(t, err, apperr.ErrInvalidCredentials)
	})

	t.Run("banned shop", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		shop := &model.Shop{
			ID:           1,
			Username:     testShopUsername,
			PasswordHash: mustHashPassword(t, testShopPassword),
			Status:       model.ShopStatusBanned,
		}

		shopRepository.EXPECT().
			GetShopByUsername(ctx, testShopUsername).
			Return(shop, nil)

		_, err := shopService.Login(ctx, &dto.LoginRequest{
			Username: testShopUsername,
			Password: testShopPassword,
		})
		requireErrorIs(t, err, apperr.ErrShopBanned)
	})
}

func TestShopRegister(t *testing.T) {
	ctx := context.Background()

	t.Run("stores the display name", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)

		shopRepository.EXPECT().
			CreateShopWithSession(ctx, gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, shop *model.Shop, _ []byte) error {
				if shop.Name != "Test Shop" {
					t.Fatalf("display name = %q, want %q", shop.Name, "Test Shop")
				}
				shop.ID = 1
				return nil
			})

		if _, err := shopService.Register(ctx, &dto.ShopRegisterRequest{
			Username: testShopUsername,
			Name:     "Test Shop",
			Password: testShopPassword,
			Email:    testShopEmail,
		}); err != nil {
			t.Fatalf("register: %v", err)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)

		shopRepository.EXPECT().
			CreateShopWithSession(ctx, gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("shop repository: %w", apperr.ErrShopUsernameTaken))

		_, err := shopService.Register(ctx, &dto.ShopRegisterRequest{
			Username: testShopUsername,
			Name:     "Test Shop",
			Password: testShopPassword,
			Email:    "another@example.com",
		})
		requireErrorIs(t, err, apperr.ErrShopUsernameTaken)
	})

	t.Run("duplicate email", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)

		shopRepository.EXPECT().
			CreateShopWithSession(ctx, gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("shop repository: %w", apperr.ErrShopEmailTaken))

		_, err := shopService.Register(ctx, &dto.ShopRegisterRequest{
			Username: "shop02",
			Name:     "Test Shop",
			Password: testShopPassword,
			Email:    testShopEmail,
		})
		requireErrorIs(t, err, apperr.ErrShopEmailTaken)
	})
}

func TestShopGetProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the shop profile", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		shop := &model.Shop{
			ID:          1,
			Username:    testShopUsername,
			Name:        "First shop",
			Description: "Intro",
			Status:      model.ShopStatusActive,
		}

		shopRepository.EXPECT().GetShopByID(ctx, 1).Return(shop, nil)

		got, err := shopService.GetProfile(ctx, 1)
		if err != nil {
			t.Fatalf("get profile: %v", err)
		}
		if got.Name != "First shop" || got.Description != "Intro" {
			t.Fatalf("profile = %+v, want updated fields", got)
		}
	})

	t.Run("unknown shop", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)

		shopRepository.EXPECT().
			GetShopByID(ctx, 99).
			Return(nil, fmt.Errorf("shop repository: %w", apperr.ErrShopNotFound))

		_, err := shopService.GetProfile(ctx, 99)
		requireErrorIs(t, err, apperr.ErrShopNotFound)
	})
}

func TestShopUpdateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("updates the shop profile", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		description := "Intro"
		updated := &model.Shop{
			ID:          1,
			Username:    testShopUsername,
			Name:        "First shop",
			Description: description,
			Status:      model.ShopStatusActive,
		}

		shopRepository.EXPECT().UpdateShopProfile(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, shop *model.Shop) (*model.Shop, error) {
			if shop.ID != 1 || shop.Name != "First shop" || shop.Description != "Intro" {
				t.Fatalf("shop = %+v, want profile fields", shop)
			}
			return updated, nil
		})

		got, err := shopService.UpdateProfile(ctx, 1, &dto.UpdateShopProfileRequest{
			Name:        "First shop",
			Description: &description,
		})
		if err != nil {
			t.Fatalf("update profile: %v", err)
		}
		if got.Name != "First shop" || got.Description != "Intro" {
			t.Fatalf("profile = %+v, want updated fields", got)
		}
	})

	t.Run("repository failure", func(t *testing.T) {
		shopService, shopRepository := newTestShopService(t)
		description := "Intro"

		shopRepository.EXPECT().
			UpdateShopProfile(ctx, gomock.Any()).
			Return(nil, errors.New("database is down"))

		_, err := shopService.UpdateProfile(ctx, 1, &dto.UpdateShopProfileRequest{
			Name:        "First shop",
			Description: &description,
		})
		if err == nil {
			t.Fatal("update profile: expected an error")
		}
	})
}
