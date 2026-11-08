package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/repository"
)

func TestShopRepositoryCreateShop(t *testing.T) {
	ctx := context.Background()
	shopRepository := repository.NewShopRepository(testPool)

	t.Run("creates an active shop", func(t *testing.T) {
		resetDatabase(t)

		shop := &model.Shop{
			Username:     testShopUsername,
			PasswordHash: testShopPassword,
			Email:        testShopEmail,
		}
		if err := shopRepository.CreateShop(ctx, shop); err != nil {
			t.Fatalf("create shop: %v", err)
		}
		if shop.ID == 0 {
			t.Fatal("shop ID was not assigned")
		}

		got, err := shopRepository.GetShopByUsername(ctx, testShopUsername)
		if err != nil {
			t.Fatalf("get shop: %v", err)
		}
		if got.ID != shop.ID || got.Username != shop.Username {
			t.Fatalf("shop = %+v, want %+v", got, shop)
		}
		if got.PasswordHash != testShopPassword {
			t.Fatalf("password hash = %q, want %q", got.PasswordHash, testShopPassword)
		}
		if got.Email != testShopEmail {
			t.Fatalf("email = %q, want %q", got.Email, testShopEmail)
		}
		if got.Status != model.ShopStatusActive {
			t.Fatalf("status = %q, want %q", got.Status, model.ShopStatusActive)
		}
		if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
			t.Fatalf("timestamps are not set: %+v", got)
		}
	})

	t.Run("rejects a duplicate username", func(t *testing.T) {
		resetDatabase(t)
		insertShop(t)

		err := shopRepository.CreateShop(ctx, &model.Shop{
			Username:     testShopUsername,
			PasswordHash: testShopPassword,
			Email:        "another@example.com",
		})
		requireErrorIs(t, err, apperr.ErrShopUsernameTaken)
	})

	t.Run("rejects a duplicate email", func(t *testing.T) {
		resetDatabase(t)
		insertShop(t)

		err := shopRepository.CreateShop(ctx, &model.Shop{
			Username:     "shop02",
			PasswordHash: testShopPassword,
			Email:        testShopEmail,
		})
		requireErrorIs(t, err, apperr.ErrShopEmailTaken)
	})

	t.Run("returns not found for an unknown username", func(t *testing.T) {
		resetDatabase(t)

		_, err := shopRepository.GetShopByUsername(ctx, testShopUsername)
		requireErrorIs(t, err, apperr.ErrShopNotFound)
	})
}

func TestShopRepositorySessions(t *testing.T) {
	ctx := context.Background()
	shopRepository := repository.NewShopRepository(testPool)

	t.Run("creates and resolves a session", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		tokenHash := []byte("shop-session-token")

		if err := shopRepository.CreateSession(ctx, shop, tokenHash); err != nil {
			t.Fatalf("create session: %v", err)
		}

		got, err := shopRepository.GetShopBySessionTokenHash(ctx, tokenHash)
		if err != nil {
			t.Fatalf("get shop by session: %v", err)
		}
		if got.ID != shop.ID {
			t.Fatalf("shop ID = %d, want %d", got.ID, shop.ID)
		}
	})

	t.Run("rejects an unknown session", func(t *testing.T) {
		resetDatabase(t)

		_, err := shopRepository.GetShopBySessionTokenHash(ctx, []byte("unknown-token"))
		requireErrorIs(t, err, apperr.ErrInvalidShopSession)
	})

	t.Run("rejects an expired session", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		tokenHash := []byte("expired-session-token")

		if err := shopRepository.CreateSession(ctx, shop, tokenHash); err != nil {
			t.Fatalf("create session: %v", err)
		}
		if _, err := testPool.Exec(ctx, `UPDATE shop_sessions SET expires_at = now() - INTERVAL '1 hour'`); err != nil {
			t.Fatalf("expire session: %v", err)
		}

		_, err := shopRepository.GetShopBySessionTokenHash(ctx, tokenHash)
		requireErrorIs(t, err, apperr.ErrInvalidShopSession)
	})

	t.Run("deletes a session", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		tokenHash := []byte("deleted-session-token")

		if err := shopRepository.CreateSession(ctx, shop, tokenHash); err != nil {
			t.Fatalf("create session: %v", err)
		}
		if err := shopRepository.DeleteSessionByTokenHash(ctx, tokenHash); err != nil {
			t.Fatalf("delete session: %v", err)
		}

		_, err := shopRepository.GetShopBySessionTokenHash(ctx, tokenHash)
		requireErrorIs(t, err, apperr.ErrInvalidShopSession)
	})
}
