package service

import (
	"context"
	"fmt"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/model"
)

// ProductCategoryRepository is the persistence the category service depends on.
type ProductCategoryRepository interface {
	ListProductCategoriesByShopID(ctx context.Context, shopID int) ([]model.ProductCategory, error)
	GetProductCategoryByID(ctx context.Context, shopID int, categoryID int) (*model.ProductCategory, error)
	CreateProductCategory(ctx context.Context, category *model.ProductCategory) error
	UpdateProductCategory(ctx context.Context, category *model.ProductCategory) error
	DeleteProductCategory(ctx context.Context, shopID int, categoryID int) error
}

type ProductCategoryService struct {
	categoryRepository ProductCategoryRepository
}

func NewProductCategoryService(categoryRepository ProductCategoryRepository) *ProductCategoryService {
	return &ProductCategoryService{categoryRepository: categoryRepository}
}

// List returns the categories of the shop ordered by name.
func (s *ProductCategoryService) List(ctx context.Context, shopID int) ([]model.ProductCategory, error) {
	categories, err := s.categoryRepository.ListProductCategoriesByShopID(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("service.ListProductCategories: %w", err)
	}
	return categories, nil
}

// Create adds a category to the shop.
func (s *ProductCategoryService) Create(ctx context.Context, shopID int, req *dto.CreateProductCategoryRequest) (*model.ProductCategory, error) {
	category := &model.ProductCategory{
		ShopID: shopID,
		Name:   req.Name,
	}

	if err := s.categoryRepository.CreateProductCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("service.CreateProductCategory: %w", err)
	}
	return category, nil
}

// Update renames a category of the shop and returns it with its product count.
func (s *ProductCategoryService) Update(ctx context.Context, shopID int, categoryID int, req *dto.UpdateProductCategoryRequest) (*model.ProductCategory, error) {
	category := &model.ProductCategory{
		ID:     categoryID,
		ShopID: shopID,
		Name:   req.Name,
	}

	if err := s.categoryRepository.UpdateProductCategory(ctx, category); err != nil {
		return nil, fmt.Errorf("service.UpdateProductCategory: %w", err)
	}

	updated, err := s.categoryRepository.GetProductCategoryByID(ctx, shopID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateProductCategory: get category: %w", err)
	}
	return updated, nil
}

// Delete removes a category of the shop; its products become uncategorized.
func (s *ProductCategoryService) Delete(ctx context.Context, shopID int, categoryID int) error {
	if err := s.categoryRepository.DeleteProductCategory(ctx, shopID, categoryID); err != nil {
		return fmt.Errorf("service.DeleteProductCategory: %w", err)
	}
	return nil
}
