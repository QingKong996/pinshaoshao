package service_test

import (
	"context"
	"errors"
	"testing"

	"pinshaoshao/internal/feature/cart/apperr"
	"pinshaoshao/internal/feature/cart/mocks"
	"pinshaoshao/internal/feature/cart/model"
	"pinshaoshao/internal/feature/cart/service"

	"go.uber.org/mock/gomock"
)

func newTestCartService(t *testing.T) (*service.CartService, *mocks.MockCartRepository) {
	t.Helper()

	repository := mocks.NewMockCartRepository(gomock.NewController(t))
	return service.NewCartService(repository), repository
}

func TestCartServiceList(t *testing.T) {
	ctx := context.Background()
	cartService, repository := newTestCartService(t)
	items := []model.CartItem{{ID: 1, UserID: 7, SKUID: 3}}

	repository.EXPECT().ListByUserID(ctx, 7).Return(items, nil)

	got, err := cartService.List(ctx, 7)
	if err != nil {
		t.Fatalf("list cart: %v", err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("items = %+v, want one item", got)
	}
}

func TestCartServiceAdd(t *testing.T) {
	ctx := context.Background()
	item := &model.CartItem{ID: 1, SKUID: 3, Quantity: 2}

	t.Run("adds an item", func(t *testing.T) {
		cartService, repository := newTestCartService(t)
		repository.EXPECT().AddItem(ctx, 7, 3, 2).Return(item, nil)

		got, err := cartService.Add(ctx, 7, 3, 2)
		if err != nil {
			t.Fatalf("add cart item: %v", err)
		}
		if got != item {
			t.Fatalf("item = %+v, want %+v", got, item)
		}
	})

	t.Run("rejects an invalid sku", func(t *testing.T) {
		cartService, _ := newTestCartService(t)

		_, err := cartService.Add(ctx, 7, 0, 1)
		requireErrorIs(t, err, apperr.ErrInvalidSKUID)
	})

	t.Run("rejects a non-positive quantity", func(t *testing.T) {
		cartService, _ := newTestCartService(t)

		_, err := cartService.Add(ctx, 7, 3, 0)
		requireErrorIs(t, err, apperr.ErrInvalidCartQuantity)
	})

	t.Run("preserves repository failures", func(t *testing.T) {
		cartService, repository := newTestCartService(t)
		dbErr := errors.New("database is down")
		repository.EXPECT().AddItem(ctx, 7, 3, 1).Return(nil, dbErr)

		_, err := cartService.Add(ctx, 7, 3, 1)
		requireErrorIs(t, err, dbErr)
	})
}

func TestCartServiceUpdateAndDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("updates quantity", func(t *testing.T) {
		cartService, repository := newTestCartService(t)
		item := &model.CartItem{ID: 4, Quantity: 5}
		repository.EXPECT().UpdateQuantity(ctx, 7, 4, 5).Return(item, nil)

		got, err := cartService.UpdateQuantity(ctx, 7, 4, 5)
		if err != nil {
			t.Fatalf("update cart item: %v", err)
		}
		if got != item {
			t.Fatalf("item = %+v, want %+v", got, item)
		}
	})

	t.Run("rejects an invalid item id", func(t *testing.T) {
		cartService, _ := newTestCartService(t)

		_, err := cartService.UpdateQuantity(ctx, 7, 0, 1)
		requireErrorIs(t, err, apperr.ErrInvalidCartItemID)
	})

	t.Run("deletes an item", func(t *testing.T) {
		cartService, repository := newTestCartService(t)
		repository.EXPECT().DeleteItem(ctx, 7, 4).Return(nil)

		if err := cartService.Delete(ctx, 7, 4); err != nil {
			t.Fatalf("delete cart item: %v", err)
		}
	})
}

func requireErrorIs(t *testing.T, err error, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}
