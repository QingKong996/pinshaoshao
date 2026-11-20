package service

import (
	"context"
	"fmt"
	"pinshaoshao/internal/feature/cart/apperr"
	"pinshaoshao/internal/feature/cart/model"
)

type CartRepository interface {
	ListByUserID(ctx context.Context, userID int) ([]model.CartItem, error)
	AddItem(ctx context.Context, userID int, skuID int, quantity int) (*model.CartItem, error)
	UpdateQuantity(ctx context.Context, userID int, itemID int, quantity int) (*model.CartItem, error)
	DeleteItem(ctx context.Context, userID int, itemID int) error
}

type CartService struct {
	cartRepository CartRepository
}

func NewCartService(cartRepository CartRepository) *CartService {
	return &CartService{cartRepository: cartRepository}
}

func (s *CartService) List(ctx context.Context, userID int) ([]model.CartItem, error) {
	items, err := s.cartRepository.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.ListCart: %w", err)
	}
	return items, nil
}

func (s *CartService) Add(ctx context.Context, userID int, skuID int, quantity int) (*model.CartItem, error) {
	if skuID <= 0 {
		return nil, fmt.Errorf("service.AddCartItem: %w", apperr.ErrInvalidSKUID)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("service.AddCartItem: %w", apperr.ErrInvalidCartQuantity)
	}

	item, err := s.cartRepository.AddItem(ctx, userID, skuID, quantity)
	if err != nil {
		return nil, fmt.Errorf("service.AddCartItem: %w", err)
	}
	return item, nil
}

func (s *CartService) UpdateQuantity(ctx context.Context, userID int, itemID int, quantity int) (*model.CartItem, error) {
	if itemID <= 0 {
		return nil, fmt.Errorf("service.UpdateCartItem: %w", apperr.ErrInvalidCartItemID)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("service.UpdateCartItem: %w", apperr.ErrInvalidCartQuantity)
	}

	item, err := s.cartRepository.UpdateQuantity(ctx, userID, itemID, quantity)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateCartItem: %w", err)
	}
	return item, nil
}

func (s *CartService) Delete(ctx context.Context, userID int, itemID int) error {
	if itemID <= 0 {
		return fmt.Errorf("service.DeleteCartItem: %w", apperr.ErrInvalidCartItemID)
	}
	if err := s.cartRepository.DeleteItem(ctx, userID, itemID); err != nil {
		return fmt.Errorf("service.DeleteCartItem: %w", err)
	}
	return nil
}
