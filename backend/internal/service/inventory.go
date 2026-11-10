package service

import (
	"context"
	"fmt"
	"pinshaoshao/internal/model"
)

// InventoryRepository is the persistence the inventory service depends on.
type InventoryRepository interface {
	ListInventoryByShopID(ctx context.Context, shopID int) ([]model.InventoryItem, error)
}

type InventoryService struct {
	inventoryRepository InventoryRepository
}

func NewInventoryService(inventoryRepository InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepository: inventoryRepository}
}

// List returns every SKU of the shop ordered by stock ascending.
func (s *InventoryService) List(ctx context.Context, shopID int) ([]model.InventoryItem, error) {
	items, err := s.inventoryRepository.ListInventoryByShopID(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("service.ListInventory: %w", err)
	}
	return items, nil
}
