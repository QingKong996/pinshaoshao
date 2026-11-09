package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/model"
)

// ProductRepository is the persistence the product service depends on.
type ProductRepository interface {
	CreateProduct(ctx context.Context, product *model.Product) error
	GetProductSummaryByID(ctx context.Context, shopID int, productID int) (*model.ProductSummary, error)
	ListProductSummariesByShopID(ctx context.Context, shopID int) ([]model.ProductSummary, error)
	UpdateProduct(ctx context.Context, product *model.Product) error
	DeleteProduct(ctx context.Context, shopID int, productID int) error
}

// ProductCategoryReader is the category lookup the product service depends on.
type ProductCategoryReader interface {
	GetProductCategoryByID(ctx context.Context, shopID int, categoryID int) (*model.ProductCategory, error)
}

type ProductService struct {
	productRepository ProductRepository
	categoryReader    ProductCategoryReader
}

func NewProductService(productRepository ProductRepository, categoryReader ProductCategoryReader) *ProductService {
	return &ProductService{
		productRepository: productRepository,
		categoryReader:    categoryReader,
	}
}

// List returns the products of the shop with their aggregates, newest first.
func (s *ProductService) List(ctx context.Context, shopID int) ([]model.ProductSummary, error) {
	products, err := s.productRepository.ListProductSummariesByShopID(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("service.ListProducts: %w", err)
	}
	return products, nil
}

// Get returns a product of the shop with its aggregates.
func (s *ProductService) Get(ctx context.Context, shopID int, productID int) (*model.ProductSummary, error) {
	product, err := s.productRepository.GetProductSummaryByID(ctx, shopID, productID)
	if err != nil {
		return nil, fmt.Errorf("service.GetProduct: %w", err)
	}
	return product, nil
}

// Create adds a product to the shop.
func (s *ProductService) Create(ctx context.Context, shopID int, req *dto.CreateProductRequest) (*model.ProductSummary, error) {
	if err := s.validateCategory(ctx, shopID, req.CategoryID); err != nil {
		return nil, fmt.Errorf("service.CreateProduct: %w", err)
	}

	status := req.Status
	if status == "" {
		status = string(model.ProductStatusOffShelf)
	}

	product := &model.Product{
		ShopID:      shopID,
		Name:        req.Name,
		Description: req.Description,
		Status:      model.ProductStatus(status),
		CategoryID:  req.CategoryID,
	}

	if err := s.productRepository.CreateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("service.CreateProduct: %w", err)
	}

	created, err := s.productRepository.GetProductSummaryByID(ctx, shopID, product.ID)
	if err != nil {
		return nil, fmt.Errorf("service.CreateProduct: get created product: %w", err)
	}
	return created, nil
}

// Update replaces the editable fields of a product of the shop.
func (s *ProductService) Update(ctx context.Context, shopID int, productID int, req *dto.UpdateProductRequest) (*model.ProductSummary, error) {
	if err := s.validateCategory(ctx, shopID, req.CategoryID); err != nil {
		return nil, fmt.Errorf("service.UpdateProduct: %w", err)
	}

	product := &model.Product{
		ID:          productID,
		ShopID:      shopID,
		Name:        req.Name,
		Description: req.Description,
		Status:      model.ProductStatus(req.Status),
		CategoryID:  req.CategoryID,
	}

	if err := s.productRepository.UpdateProduct(ctx, product); err != nil {
		return nil, fmt.Errorf("service.UpdateProduct: %w", err)
	}

	updated, err := s.productRepository.GetProductSummaryByID(ctx, shopID, productID)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateProduct: get updated product: %w", err)
	}
	return updated, nil
}

// Delete removes a product of the shop; its SKUs are removed with it.
func (s *ProductService) Delete(ctx context.Context, shopID int, productID int) error {
	if err := s.productRepository.DeleteProduct(ctx, shopID, productID); err != nil {
		return fmt.Errorf("service.DeleteProduct: %w", err)
	}
	return nil
}

func (s *ProductService) validateCategory(ctx context.Context, shopID int, categoryID *int) error {
	if categoryID == nil {
		return nil
	}

	if _, err := s.categoryReader.GetProductCategoryByID(ctx, shopID, *categoryID); err != nil {
		if errors.Is(err, apperr.ErrCategoryNotFound) {
			return apperr.ErrInvalidCategoryID
		}
		return fmt.Errorf("get category: %w", err)
	}
	return nil
}
