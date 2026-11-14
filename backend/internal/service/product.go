package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"strings"
)

const (
	defaultProductPageSize = 10
	maxProductPageSize     = 100
	maxProductPage         = 1_000_000
)

type CreateProductInput struct {
	Name        string
	Description string
	CategoryID  *int
	Status      model.ProductStatus
}

type UpdateProductInput struct {
	Name        string
	Description string
	CategoryID  *int
	Status      model.ProductStatus
}

// ProductRepository is the persistence the product service depends on.
type ProductRepository interface {
	CreateProduct(ctx context.Context, product *model.Product) error
	GetProductSummaryByID(ctx context.Context, shopID int, productID int) (*model.ProductSummary, error)
	ListProductSummaries(ctx context.Context, shopID int, filter model.ProductListFilter) (*model.ProductListResult, error)
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

// List returns a filtered and sorted page of products with their aggregates.
func (s *ProductService) List(ctx context.Context, shopID int, filter model.ProductListFilter) (*model.ProductListResult, error) {
	normalized, err := normalizeProductListFilter(filter)
	if err != nil {
		return nil, err
	}

	result, err := s.productRepository.ListProductSummaries(ctx, shopID, normalized)
	if err != nil {
		return nil, fmt.Errorf("service.ListProducts: %w", err)
	}
	result.Page = normalized.Page
	result.PageSize = normalized.PageSize
	return result, nil
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
func (s *ProductService) Create(ctx context.Context, shopID int, input CreateProductInput) (*model.ProductSummary, error) {
	if err := s.validateCategory(ctx, shopID, input.CategoryID); err != nil {
		return nil, fmt.Errorf("service.CreateProduct: %w", err)
	}

	status := input.Status
	if status == "" {
		status = model.ProductStatusOffShelf
	}

	product := &model.Product{
		ShopID:      shopID,
		Name:        input.Name,
		Description: input.Description,
		Status:      status,
		CategoryID:  input.CategoryID,
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
func (s *ProductService) Update(ctx context.Context, shopID int, productID int, input UpdateProductInput) (*model.ProductSummary, error) {
	if err := s.validateCategory(ctx, shopID, input.CategoryID); err != nil {
		return nil, fmt.Errorf("service.UpdateProduct: %w", err)
	}

	product := &model.Product{
		ID:          productID,
		ShopID:      shopID,
		Name:        input.Name,
		Description: input.Description,
		Status:      input.Status,
		CategoryID:  input.CategoryID,
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

func normalizeProductListFilter(filter model.ProductListFilter) (model.ProductListFilter, error) {
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = defaultProductPageSize
	}
	if filter.Page < 1 || filter.Page > maxProductPage || filter.PageSize < 1 || filter.PageSize > maxProductPageSize {
		return model.ProductListFilter{}, apperr.ErrInvalidProductListQuery
	}

	if filter.Sort == "" {
		filter.Sort = model.ProductSortCreatedAt
	}
	switch filter.Sort {
	case model.ProductSortCreatedAt,
		model.ProductSortUpdatedAt,
		model.ProductSortName,
		model.ProductSortCategory,
		model.ProductSortStatus,
		model.ProductSortSKUCount,
		model.ProductSortTotalStock,
		model.ProductSortMinPrice:
	default:
		return model.ProductListFilter{}, apperr.ErrInvalidProductListQuery
	}

	if filter.Order == "" {
		filter.Order = model.ProductSortOrderDesc
	}
	if filter.Order != model.ProductSortOrderAsc && filter.Order != model.ProductSortOrderDesc {
		return model.ProductListFilter{}, apperr.ErrInvalidProductListQuery
	}

	if filter.Status != "" && filter.Status != model.ProductStatusOnShelf && filter.Status != model.ProductStatusOffShelf {
		return model.ProductListFilter{}, apperr.ErrInvalidProductListQuery
	}
	if filter.CategoryID != nil && *filter.CategoryID <= 0 {
		return model.ProductListFilter{}, apperr.ErrInvalidProductListQuery
	}
	if filter.CategoryID != nil && filter.Uncategorized {
		return model.ProductListFilter{}, apperr.ErrInvalidProductListQuery
	}

	filter.Search = strings.TrimSpace(filter.Search)
	return filter, nil
}
