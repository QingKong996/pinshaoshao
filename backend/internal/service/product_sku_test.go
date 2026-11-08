package service_test

import (
	"context"
	"fmt"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/service"

	"go.uber.org/mock/gomock"
)

var testProduct = &model.Product{ID: 1, Name: "T-shirt"}

func newTestProductSKUService(t *testing.T) (*service.ProductSKUService, *mocks.MockProductReader, *mocks.MockProductSKURepository) {
	t.Helper()

	ctrl := gomock.NewController(t)
	productReader := mocks.NewMockProductReader(ctrl)
	skuRepository := mocks.NewMockProductSKURepository(ctrl)
	return service.NewProductSKUService(productReader, skuRepository), productReader, skuRepository
}

func TestCreateProductSKU(t *testing.T) {
	ctx := context.Background()

	t.Run("creates an enabled SKU by default", func(t *testing.T) {
		skuService, _, skuRepository := newTestProductSKUService(t)

		skuRepository.EXPECT().CreateProductSKU(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, sku *model.ProductSKU) error {
			sku.ID = 1
			return nil
		})

		sku, err := skuService.Create(ctx, 1, &dto.CreateProductSKURequest{
			Specs:      map[string]string{"color": "black", "size": "M"},
			PriceCents: 1999,
			Stock:      100,
		})
		if err != nil {
			t.Fatalf("create sku: %v", err)
		}
		if sku.ID == 0 {
			t.Fatal("sku ID was not assigned")
		}
		if sku.ProductID != 1 {
			t.Fatalf("product ID = %d, want 1", sku.ProductID)
		}
		if !sku.IsEnabled {
			t.Fatal("sku is disabled, want enabled by default")
		}
		if got := sku.Specs; got["color"] != "black" || got["size"] != "M" {
			t.Fatalf("specs = %v", got)
		}
	})

	t.Run("accepts an explicit disabled flag", func(t *testing.T) {
		skuService, _, skuRepository := newTestProductSKUService(t)
		disabled := false

		skuRepository.EXPECT().CreateProductSKU(ctx, gomock.Any()).Return(nil)

		sku, err := skuService.Create(ctx, 1, &dto.CreateProductSKURequest{
			Specs:     map[string]string{"color": "black"},
			IsEnabled: &disabled,
		})
		if err != nil {
			t.Fatalf("create sku: %v", err)
		}
		if sku.IsEnabled {
			t.Fatal("sku is enabled, want disabled")
		}
	})

	t.Run("normalizes missing specs to an empty object", func(t *testing.T) {
		skuService, _, skuRepository := newTestProductSKUService(t)

		skuRepository.EXPECT().CreateProductSKU(ctx, gomock.Any()).Return(nil)

		sku, err := skuService.Create(ctx, 1, &dto.CreateProductSKURequest{})
		if err != nil {
			t.Fatalf("create sku: %v", err)
		}
		if sku.Specs == nil {
			t.Fatal("specs is nil, want empty object")
		}
		if len(sku.Specs) != 0 {
			t.Fatalf("specs = %v, want empty", sku.Specs)
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		skuService, _, skuRepository := newTestProductSKUService(t)

		skuRepository.EXPECT().
			CreateProductSKU(ctx, gomock.Any()).
			Return(fmt.Errorf("product repository: %w", apperr.ErrProductNotFound))

		_, err := skuService.Create(ctx, 99, &dto.CreateProductSKURequest{
			Specs: map[string]string{"color": "black"},
		})
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("duplicate specs", func(t *testing.T) {
		skuService, _, skuRepository := newTestProductSKUService(t)

		skuRepository.EXPECT().
			CreateProductSKU(ctx, gomock.Any()).
			Return(fmt.Errorf("sku repository: %w", apperr.ErrProductSKUAlreadyExists))

		_, err := skuService.Create(ctx, 1, &dto.CreateProductSKURequest{
			Specs:      map[string]string{"color": "black", "size": "M"},
			PriceCents: 1999,
		})
		requireErrorIs(t, err, apperr.ErrProductSKUAlreadyExists)
	})

	t.Run("rejects spec keys that differ from existing skus", func(t *testing.T) {
		skuService, _, skuRepository := newTestProductSKUService(t)

		skuRepository.EXPECT().
			CreateProductSKU(ctx, gomock.Any()).
			Return(fmt.Errorf("sku repository: %w", apperr.ErrProductSKUSpecsMismatch))

		_, err := skuService.Create(ctx, 1, &dto.CreateProductSKURequest{
			Specs: map[string]string{"colour": "white", "size": "L"},
		})
		requireErrorIs(t, err, apperr.ErrProductSKUSpecsMismatch)
	})
}

func TestListProductSKUs(t *testing.T) {
	ctx := context.Background()

	t.Run("returns the SKUs of the product", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)
		skus := []model.ProductSKU{
			{
				ID:         1,
				ProductID:  1,
				Specs:      map[string]string{"color": "black", "size": "M"},
				PriceCents: 1999,
				Stock:      10,
			},
			{
				ID:         2,
				ProductID:  1,
				Specs:      map[string]string{"color": "white", "size": "L"},
				PriceCents: 1999,
				Stock:      10,
			},
		}

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().ListProductSKUsByProductID(ctx, 1).Return(skus, nil)

		got, err := skuService.List(ctx, 1)
		if err != nil {
			t.Fatalf("list skus: %v", err)
		}
		if len(got) != len(skus) {
			t.Fatalf("sku count = %d, want %d", len(got), len(skus))
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)

		productReader.EXPECT().
			GetProductByID(ctx, 99).
			Return(nil, fmt.Errorf("product repository: %w", apperr.ErrProductNotFound))

		_, err := skuService.List(ctx, 99)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})
}
