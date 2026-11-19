package service_test

import (
	"context"
	"errors"
	"testing"

	"pinshaoshao/internal/feature/inventory/apperr"
	"pinshaoshao/internal/feature/inventory/mocks"
	"pinshaoshao/internal/feature/inventory/model"
	"pinshaoshao/internal/feature/inventory/service"

	"go.uber.org/mock/gomock"
)

func newTestInventoryService(t *testing.T) (*service.InventoryService, *mocks.MockInventoryRepository) {
	t.Helper()

	repository := mocks.NewMockInventoryRepository(gomock.NewController(t))
	return service.NewInventoryService(repository), repository
}

func TestListInventory(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the inventory of the shop", func(t *testing.T) {
		inventoryService, repository := newTestInventoryService(t)
		items := []model.InventoryItem{
			{SKUID: 5, ProductID: 1, ProductName: "T-shirt", TotalStock: 2, IsEnabled: true},
			{SKUID: 6, ProductID: 1, ProductName: "T-shirt", TotalStock: 50, IsEnabled: false},
		}

		repository.EXPECT().ListInventoryByShopID(ctx, 1).Return(items, nil)

		got, err := inventoryService.List(ctx, 1)
		if err != nil {
			t.Fatalf("list inventory: %v", err)
		}
		if len(got) != len(items) {
			t.Fatalf("inventory count = %d, want %d", len(got), len(items))
		}
	})

	t.Run("repository failure", func(t *testing.T) {
		inventoryService, repository := newTestInventoryService(t)
		dbErr := errors.New("database is down")

		repository.EXPECT().ListInventoryByShopID(ctx, 1).Return(nil, dbErr)

		_, err := inventoryService.List(ctx, 1)
		requireErrorIs(t, err, dbErr)
	})
}

func TestUpdateInventory(t *testing.T) {
	ctx := context.Background()

	t.Run("updates the total stock", func(t *testing.T) {
		inventoryService, repository := newTestInventoryService(t)
		item := &model.InventoryItem{SKUID: 5, TotalStock: 12, FrozenStock: 2, Version: 4}

		repository.EXPECT().
			UpdateInventoryTotalStock(ctx, 1, 5, 12, int64(3)).
			Return(item, nil)

		got, err := inventoryService.Update(ctx, 1, 5, service.UpdateInventoryInput{TotalStock: 12, Version: 3})
		if err != nil {
			t.Fatalf("update inventory: %v", err)
		}
		if got != item {
			t.Fatalf("inventory item = %+v, want %+v", got, item)
		}
	})

	t.Run("rejects a negative update", func(t *testing.T) {
		inventoryService, _ := newTestInventoryService(t)

		_, err := inventoryService.Update(ctx, 1, 5, service.UpdateInventoryInput{TotalStock: -1})
		requireErrorIs(t, err, apperr.ErrInvalidInventoryUpdate)
	})
}
