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

func newTestProductService(t *testing.T) (*service.ProductService, *mocks.MockProductRepository, *mocks.MockProductCategoryReader) {
	t.Helper()

	ctrl := gomock.NewController(t)
	productRepository := mocks.NewMockProductRepository(ctrl)
	categoryReader := mocks.NewMockProductCategoryReader(ctrl)
	return service.NewProductService(productRepository, categoryReader), productRepository, categoryReader
}

func testProductSummary() *model.ProductSummary {
	return &model.ProductSummary{
		Product: model.Product{
			ID:          1,
			ShopID:      1,
			Name:        "T-shirt",
			Description: "Cotton T-shirt",
			Status:      model.ProductStatusOnShelf,
		},
		SKUCount:   2,
		TotalStock: 30,
	}
}

func TestListProducts(t *testing.T) {
	ctx := context.Background()

	t.Run("returns a page of products with defaults", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)
		products := []model.ProductSummary{*testProductSummary()}

		productRepository.EXPECT().
			ListProductSummaries(ctx, 1, model.ProductListFilter{
				Sort:     model.ProductSortCreatedAt,
				Order:    model.ProductSortOrderDesc,
				Page:     1,
				PageSize: 10,
			}).
			Return(&model.ProductListResult{Products: products, Total: 1}, nil)

		got, err := productService.List(ctx, 1, model.ProductListFilter{})
		if err != nil {
			t.Fatalf("list products: %v", err)
		}
		if len(got.Products) != 1 {
			t.Fatalf("product count = %d, want 1", len(got.Products))
		}
		if got.Page != 1 || got.PageSize != 10 || got.Total != 1 {
			t.Fatalf("page metadata = %+v, want page 1 size 10 total 1", got)
		}
	})

	t.Run("forwards filters and sorting", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)
		categoryID := 2
		input := model.ProductListFilter{
			Search:     "  shirt  ",
			Status:     model.ProductStatusOnShelf,
			CategoryID: &categoryID,
			Sort:       model.ProductSortTotalStock,
			Order:      model.ProductSortOrderAsc,
			Page:       3,
			PageSize:   5,
		}
		want := input
		want.Search = "shirt"

		productRepository.EXPECT().
			ListProductSummaries(ctx, 1, want).
			DoAndReturn(func(_ context.Context, _ int, got model.ProductListFilter) (*model.ProductListResult, error) {
				if got.Search != "shirt" {
					t.Fatalf("search = %q, want trimmed search", got.Search)
				}
				return &model.ProductListResult{Products: []model.ProductSummary{}, Total: 0}, nil
			})

		if _, err := productService.List(ctx, 1, input); err != nil {
			t.Fatalf("list products: %v", err)
		}
	})

	t.Run("rejects an invalid sort", func(t *testing.T) {
		productService, _, _ := newTestProductService(t)

		_, err := productService.List(ctx, 1, model.ProductListFilter{Sort: "price"})
		requireErrorIs(t, err, apperr.ErrInvalidProductListQuery)
	})

	t.Run("repository failure", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)
		dbErr := errors.New("database is down")

		productRepository.EXPECT().
			ListProductSummaries(ctx, 1, model.ProductListFilter{
				Sort:     model.ProductSortCreatedAt,
				Order:    model.ProductSortOrderDesc,
				Page:     1,
				PageSize: 10,
			}).
			Return(nil, dbErr)

		_, err := productService.List(ctx, 1, model.ProductListFilter{})
		requireErrorIs(t, err, dbErr)
	})
}

func TestGetProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the product of the shop", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)

		productRepository.EXPECT().GetProductSummaryByID(ctx, 1, 2).Return(testProductSummary(), nil)

		got, err := productService.Get(ctx, 1, 2)
		if err != nil {
			t.Fatalf("get product: %v", err)
		}
		if got.ID != 1 {
			t.Fatalf("product ID = %d, want 1", got.ID)
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)

		productRepository.EXPECT().
			GetProductSummaryByID(ctx, 1, 99).
			Return(nil, fmt.Errorf("product repository: %w", apperr.ErrProductNotFound))

		_, err := productService.Get(ctx, 1, 99)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})
}

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults to off the shelf without a category", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)
		created := testProductSummary()

		productRepository.EXPECT().CreateProduct(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, product *model.Product) error {
			if product.Status != model.ProductStatusOffShelf {
				t.Fatalf("status = %q, want %q", product.Status, model.ProductStatusOffShelf)
			}
			if product.CategoryID != nil {
				t.Fatalf("category ID = %v, want nil", product.CategoryID)
			}
			product.ID = created.ID
			return nil
		})
		productRepository.EXPECT().GetProductSummaryByID(ctx, 1, created.ID).Return(created, nil)

		got, err := productService.Create(ctx, 1, service.CreateProductInput{
			Name:        "T-shirt",
			Description: "Cotton T-shirt",
		})
		if err != nil {
			t.Fatalf("create product: %v", err)
		}
		if got.ID != created.ID {
			t.Fatalf("product ID = %d, want %d", got.ID, created.ID)
		}
	})

	t.Run("creates an on-shelf product in a category", func(t *testing.T) {
		productService, productRepository, categoryReader := newTestProductService(t)
		categoryID := 2
		created := testProductSummary()

		categoryReader.EXPECT().GetProductCategoryByID(ctx, 1, categoryID).Return(&model.ProductCategory{ID: categoryID, ShopID: 1}, nil)
		productRepository.EXPECT().CreateProduct(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, product *model.Product) error {
			if product.Status != model.ProductStatusOnShelf {
				t.Fatalf("status = %q, want %q", product.Status, model.ProductStatusOnShelf)
			}
			if product.CategoryID == nil || *product.CategoryID != categoryID {
				t.Fatalf("category ID = %v, want %d", product.CategoryID, categoryID)
			}
			product.ID = created.ID
			return nil
		})
		productRepository.EXPECT().GetProductSummaryByID(ctx, 1, created.ID).Return(created, nil)

		_, err := productService.Create(ctx, 1, service.CreateProductInput{
			Name:       "T-shirt",
			Status:     model.ProductStatusOnShelf,
			CategoryID: &categoryID,
		})
		if err != nil {
			t.Fatalf("create product: %v", err)
		}
	})

	t.Run("rejects a category of another shop", func(t *testing.T) {
		productService, _, categoryReader := newTestProductService(t)
		categoryID := 2

		categoryReader.EXPECT().
			GetProductCategoryByID(ctx, 1, categoryID).
			Return(nil, fmt.Errorf("category repository: %w", apperr.ErrCategoryNotFound))

		_, err := productService.Create(ctx, 1, service.CreateProductInput{Name: "T-shirt", CategoryID: &categoryID})
		requireErrorIs(t, err, apperr.ErrInvalidCategoryID)
	})

	t.Run("repository failure", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)
		dbErr := errors.New("database is down")

		productRepository.EXPECT().CreateProduct(ctx, gomock.Any()).Return(dbErr)

		_, err := productService.Create(ctx, 1, service.CreateProductInput{Name: "T-shirt"})
		requireErrorIs(t, err, dbErr)
	})
}

func TestUpdateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("updates the product", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)
		updated := testProductSummary()

		productRepository.EXPECT().UpdateProduct(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, product *model.Product) error {
			if product.ID != 2 || product.ShopID != 1 {
				t.Fatalf("product = %+v, want id 2 of shop 1", product)
			}
			if product.Status != model.ProductStatusOnShelf {
				t.Fatalf("status = %q, want %q", product.Status, model.ProductStatusOnShelf)
			}
			return nil
		})
		productRepository.EXPECT().GetProductSummaryByID(ctx, 1, 2).Return(updated, nil)

		got, err := productService.Update(ctx, 1, 2, service.UpdateProductInput{
			Name:   "Hoodie",
			Status: model.ProductStatusOnShelf,
		})
		if err != nil {
			t.Fatalf("update product: %v", err)
		}
		if got.ID != updated.ID {
			t.Fatalf("product ID = %d, want %d", got.ID, updated.ID)
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)

		productRepository.EXPECT().
			UpdateProduct(ctx, gomock.Any()).
			Return(fmt.Errorf("product repository: %w", apperr.ErrProductNotFound))

		_, err := productService.Update(ctx, 1, 99, service.UpdateProductInput{Name: "Hoodie", Status: model.ProductStatusOnShelf})
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("rejects a category of another shop", func(t *testing.T) {
		productService, _, categoryReader := newTestProductService(t)
		categoryID := 2

		categoryReader.EXPECT().
			GetProductCategoryByID(ctx, 1, categoryID).
			Return(nil, fmt.Errorf("category repository: %w", apperr.ErrCategoryNotFound))

		_, err := productService.Update(ctx, 1, 2, service.UpdateProductInput{
			Name:       "Hoodie",
			Status:     model.ProductStatusOnShelf,
			CategoryID: &categoryID,
		})
		requireErrorIs(t, err, apperr.ErrInvalidCategoryID)
	})
}

func TestDeleteProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes the product", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)

		productRepository.EXPECT().DeleteProduct(ctx, 1, 2).Return(nil)

		if err := productService.Delete(ctx, 1, 2); err != nil {
			t.Fatalf("delete product: %v", err)
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		productService, productRepository, _ := newTestProductService(t)

		productRepository.EXPECT().
			DeleteProduct(ctx, 1, 99).
			Return(fmt.Errorf("product repository: %w", apperr.ErrProductNotFound))

		err := productService.Delete(ctx, 1, 99)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})
}
