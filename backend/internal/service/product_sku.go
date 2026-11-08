package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/model"
)

// ProductReader is the product lookup the SKU service depends on.
type ProductReader interface {
	GetProductByID(ctx context.Context, id int) (*model.Product, error)
}

// ProductSKURepository is the persistence the SKU service depends on.
type ProductSKURepository interface {
	CreateProductSKU(ctx context.Context, sku *model.ProductSKU) error
	ListProductSKUsByProductID(ctx context.Context, productID int) ([]model.ProductSKU, error)
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

func (s *ProductSKUService) Create(ctx context.Context, shopID int, productID int, req *dto.CreateProductSKURequest) (*model.ProductSKU, error) {
	product, err := s.productRepository.GetProductByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("service.CreateProductSKU: get product: %w", err)
	}
	if product.ShopID != shopID {
		return nil, fmt.Errorf("service.CreateProductSKU: %w", apperr.ErrProductAccessDenied)
	}

	specs := make(map[string]string, len(req.Specs))
	for name, value := range req.Specs {
		specs[name] = value
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	sku := &model.ProductSKU{
		ProductID:  productID,
		Specs:      specs,
		PriceCents: req.PriceCents,
		Stock:      req.Stock,
		IsEnabled:  isEnabled,
	}

	if err := s.skuRepository.CreateProductSKU(ctx, sku); err != nil {
		return nil, fmt.Errorf("service.CreateProductSKU: %w", err)
	}
	return sku, nil
}

func (s *ProductSKUService) List(ctx context.Context, productID int) ([]model.ProductSKU, error) {
	if err := s.ensureProductExists(ctx, productID); err != nil {
		return nil, fmt.Errorf("service.ListProductSKUs: %w", err)
	}

	skus, err := s.skuRepository.ListProductSKUsByProductID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("service.ListProductSKUs: %w", err)
	}
	return skus, nil
}

func (s *ProductSKUService) ensureProductExists(ctx context.Context, productID int) error {
	_, err := s.productRepository.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, apperr.ErrProductNotFound) {
			return apperr.ErrProductNotFound
		}
		return fmt.Errorf("get product: %w", err)
	}
	return nil
}
