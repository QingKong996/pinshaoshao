package service_test

import (
	"context"
	"errors"
	"testing"

	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/service"

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
			{SKUID: 5, ProductID: 1, ProductName: "T-shirt", Stock: 2, IsEnabled: true},
			{SKUID: 6, ProductID: 1, ProductName: "T-shirt", Stock: 50, IsEnabled: false},
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
