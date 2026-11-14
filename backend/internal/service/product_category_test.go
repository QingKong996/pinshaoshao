package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/service"

	"go.uber.org/mock/gomock"
)

func newTestProductCategoryService(t *testing.T) (*service.ProductCategoryService, *mocks.MockProductCategoryRepository) {
	t.Helper()

	repository := mocks.NewMockProductCategoryRepository(gomock.NewController(t))
	return service.NewProductCategoryService(repository), repository
}

func TestListProductCategories(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the categories of the shop", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)

		repository.EXPECT().ListProductCategoriesByShopID(ctx, 1).Return([]model.ProductCategory{{ID: 1, ShopID: 1, Name: "clothes"}}, nil)

		got, err := categoryService.List(ctx, 1)
		if err != nil {
			t.Fatalf("list categories: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("category count = %d, want 1", len(got))
		}
	})

	t.Run("repository failure", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)
		dbErr := errors.New("database is down")

		repository.EXPECT().ListProductCategoriesByShopID(ctx, 1).Return(nil, dbErr)

		_, err := categoryService.List(ctx, 1)
		requireErrorIs(t, err, dbErr)
	})
}

func TestCreateProductCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("creates a category", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)

		repository.EXPECT().CreateProductCategory(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, category *model.ProductCategory) error {
			if category.ShopID != 1 || category.Name != "clothes" {
				t.Fatalf("category = %+v, want shop 1 name clothes", category)
			}
			category.ID = 3
			return nil
		})

		got, err := categoryService.Create(ctx, 1, "clothes")
		if err != nil {
			t.Fatalf("create category: %v", err)
		}
		if got.ID != 3 {
			t.Fatalf("category ID = %d, want 3", got.ID)
		}
		if got.ProductCount != 0 {
			t.Fatalf("product count = %d, want 0", got.ProductCount)
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)

		repository.EXPECT().
			CreateProductCategory(ctx, gomock.Any()).
			Return(fmt.Errorf("category repository: %w", apperr.ErrCategoryNameTaken))

		_, err := categoryService.Create(ctx, 1, "clothes")
		requireErrorIs(t, err, apperr.ErrCategoryNameTaken)
	})
}

func TestUpdateProductCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("renames the category", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)
		updated := &model.ProductCategory{ID: 3, ShopID: 1, Name: "shoes", ProductCount: 2}

		repository.EXPECT().UpdateProductCategory(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, category *model.ProductCategory) error {
			if category.ID != 3 || category.ShopID != 1 || category.Name != "shoes" {
				t.Fatalf("category = %+v, want id 3 of shop 1 named shoes", category)
			}
			return nil
		})
		repository.EXPECT().GetProductCategoryByID(ctx, 1, 3).Return(updated, nil)

		got, err := categoryService.Update(ctx, 1, 3, "shoes")
		if err != nil {
			t.Fatalf("update category: %v", err)
		}
		if got.ProductCount != 2 {
			t.Fatalf("product count = %d, want 2", got.ProductCount)
		}
	})

	t.Run("unknown category", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)

		repository.EXPECT().
			UpdateProductCategory(ctx, gomock.Any()).
			Return(fmt.Errorf("category repository: %w", apperr.ErrCategoryNotFound))

		_, err := categoryService.Update(ctx, 1, 99, "shoes")
		requireErrorIs(t, err, apperr.ErrCategoryNotFound)
	})

	t.Run("duplicate name", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)

		repository.EXPECT().
			UpdateProductCategory(ctx, gomock.Any()).
			Return(fmt.Errorf("category repository: %w", apperr.ErrCategoryNameTaken))

		_, err := categoryService.Update(ctx, 1, 3, "shoes")
		requireErrorIs(t, err, apperr.ErrCategoryNameTaken)
	})
}

func TestDeleteProductCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes the category", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)

		repository.EXPECT().DeleteProductCategory(ctx, 1, 3).Return(nil)

		if err := categoryService.Delete(ctx, 1, 3); err != nil {
			t.Fatalf("delete category: %v", err)
		}
	})

	t.Run("unknown category", func(t *testing.T) {
		categoryService, repository := newTestProductCategoryService(t)

		repository.EXPECT().
			DeleteProductCategory(ctx, 1, 99).
			Return(fmt.Errorf("category repository: %w", apperr.ErrCategoryNotFound))

		err := categoryService.Delete(ctx, 1, 99)
		requireErrorIs(t, err, apperr.ErrCategoryNotFound)
	})
}
