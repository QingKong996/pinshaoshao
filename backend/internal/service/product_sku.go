package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"strings"
	"unicode/utf8"
)

const (
	maxSKUSpecEntries   = 5
	maxSKUSpecKeyLength = 20
	maxSKUSpecValueLen  = 50
)

type CreateProductSKUInput struct {
	Specs      map[string]string
	PriceCents int64
	Stock      int
	IsEnabled  *bool
}

// ProductReader is the product lookup the SKU service depends on.
type ProductReader interface {
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
}

// ProductSKURepository is the persistence the SKU service depends on.
type ProductSKURepository interface {
	CreateProductSKU(ctx context.Context, sku *model.ProductSKU) error
	ListProductSKUsByProductID(ctx context.Context, productID int) ([]model.ProductSKU, error)
	UpdateProductSKU(ctx context.Context, productID int, skuID int, update *model.ProductSKUUpdate) (*model.ProductSKU, error)
	DeleteProductSKU(ctx context.Context, productID int, skuID int) error
}

type ProductSKUService struct {
	productRepository ProductReader
	skuRepository     ProductSKURepository
}

func NewProductSKUService(productRepository ProductReader, skuRepository ProductSKURepository) *ProductSKUService {
	return &ProductSKUService{
		productRepository: productRepository,
		skuRepository:     skuRepository,
	}
}

// Create adds a SKU to a product owned by the shop.
func (s *ProductSKUService) Create(ctx context.Context, shopID int, productID int, input CreateProductSKUInput) (*model.ProductSKU, error) {
	if err := s.ensureOwnedProduct(ctx, shopID, productID); err != nil {
		return nil, fmt.Errorf("service.CreateProductSKU: %w", err)
	}

	specs, err := normalizeSpecs(input.Specs)
	if err != nil {
		return nil, fmt.Errorf("service.CreateProductSKU: %w", err)
	}

	isEnabled := true
	if input.IsEnabled != nil {
		isEnabled = *input.IsEnabled
	}

	sku := &model.ProductSKU{
		ProductID:  productID,
		Specs:      specs,
		PriceCents: input.PriceCents,
		Stock:      input.Stock,
		IsEnabled:  isEnabled,
	}

	if err := s.skuRepository.CreateProductSKU(ctx, sku); err != nil {
		return nil, fmt.Errorf("service.CreateProductSKU: %w", err)
	}
	return sku, nil
}

// List returns the SKUs of a product owned by the shop.
func (s *ProductSKUService) List(ctx context.Context, shopID int, productID int) ([]model.ProductSKU, error) {
	if err := s.ensureOwnedProduct(ctx, shopID, productID); err != nil {
		return nil, fmt.Errorf("service.ListProductSKUs: %w", err)
	}

	skus, err := s.skuRepository.ListProductSKUsByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("service.ListProductSKUs: %w", err)
	}
	return skus, nil
}

// Update partially updates a SKU of a product owned by the shop.
func (s *ProductSKUService) Update(ctx context.Context, shopID int, productID int, skuID int, update *model.ProductSKUUpdate) (*model.ProductSKU, error) {
	if err := s.ensureOwnedProduct(ctx, shopID, productID); err != nil {
		return nil, fmt.Errorf("service.UpdateProductSKU: %w", err)
	}

	if update.Specs == nil && update.PriceCents == nil && update.Stock == nil && update.IsEnabled == nil {
		return nil, fmt.Errorf("service.UpdateProductSKU: %w", apperr.ErrInvalidProductSKUUpdate)
	}

	normalized := *update
	if update.Specs != nil {
		specs, err := normalizeSpecs(update.Specs)
		if err != nil {
			return nil, fmt.Errorf("service.UpdateProductSKU: %w", err)
		}
		normalized.Specs = specs
	}

	sku, err := s.skuRepository.UpdateProductSKU(ctx, productID, skuID, &normalized)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateProductSKU: %w", err)
	}
	return sku, nil
}

// Delete removes a SKU of a product owned by the shop.
func (s *ProductSKUService) Delete(ctx context.Context, shopID int, productID int, skuID int) error {
	if err := s.ensureOwnedProduct(ctx, shopID, productID); err != nil {
		return fmt.Errorf("service.DeleteProductSKU: %w", err)
	}

	if err := s.skuRepository.DeleteProductSKU(ctx, productID, skuID); err != nil {
		return fmt.Errorf("service.DeleteProductSKU: %w", err)
	}
	return nil
}

// ensureOwnedProduct returns ErrProductNotFound for a missing or foreign product.
func (s *ProductSKUService) ensureOwnedProduct(ctx context.Context, shopID int, productID int) error {
	product, err := s.productRepository.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, apperr.ErrProductNotFound) {
			return apperr.ErrProductNotFound
		}
		return fmt.Errorf("get product: %w", err)
	}
	if product.ShopID != shopID {
		return apperr.ErrProductNotFound
	}
	return nil
}

// normalizeSpecs trims the spec keys and values and validates their shape.
func normalizeSpecs(specs map[string]string) (map[string]string, error) {
	if len(specs) > maxSKUSpecEntries {
		return nil, apperr.ErrInvalidProductSKUSpecs
	}

	normalized := make(map[string]string, len(specs))
	for key, value := range specs {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" ||
			utf8.RuneCountInString(key) > maxSKUSpecKeyLength ||
			utf8.RuneCountInString(value) > maxSKUSpecValueLen {
			return nil, apperr.ErrInvalidProductSKUSpecs
		}
		if _, exists := normalized[key]; exists {
			return nil, apperr.ErrInvalidProductSKUSpecs
		}
		normalized[key] = value
	}
	return normalized, nil
}
