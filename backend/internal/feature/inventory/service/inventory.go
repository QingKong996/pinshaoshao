package service

import (
	"context"
	"fmt"
	"pinshaoshao/internal/feature/inventory/apperr"
	"pinshaoshao/internal/feature/inventory/model"
)

// InventoryRepository is the persistence the inventory service depends on.
type InventoryRepository interface {
	ListInventoryByShopID(ctx context.Context, shopID int) ([]model.InventoryItem, error)
	UpdateInventoryTotalStock(ctx context.Context, shopID int, skuID int, totalStock int, version int64) (*model.InventoryItem, error)
}

type UpdateInventoryInput struct {
	TotalStock int
	Version    int64
}

type InventoryService struct {
	inventoryRepository InventoryRepository
}

func NewInventoryService(inventoryRepository InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepository: inventoryRepository}
}

// List returns every inventory record of the shop ordered by available stock ascending.
func (s *InventoryService) List(ctx context.Context, shopID int) ([]model.InventoryItem, error) {
	items, err := s.inventoryRepository.ListInventoryByShopID(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("service.ListInventory: %w", err)
	}
	return items, nil
}

func (s *InventoryService) Update(ctx context.Context, shopID int, skuID int, input UpdateInventoryInput) (*model.InventoryItem, error) {
	if input.TotalStock < 0 || input.Version < 0 {
		return nil, fmt.Errorf("service.UpdateInventory: %w", apperr.ErrInvalidInventoryUpdate)
	}

	item, err := s.inventoryRepository.UpdateInventoryTotalStock(ctx, shopID, skuID, input.TotalStock, input.Version)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateInventory: %w", err)
	}
	return item, nil
}
