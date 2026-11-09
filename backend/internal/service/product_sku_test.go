package service_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/mocks"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/service"

	"go.uber.org/mock/gomock"
)

var testProduct = &model.Product{ID: 1, ShopID: 1, Name: "T-shirt"}

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
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().CreateProductSKU(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, sku *model.ProductSKU) error {
			sku.ID = 1
			return nil
		})

		sku, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{
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
		skuService, productReader, skuRepository := newTestProductSKUService(t)
		disabled := false

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().CreateProductSKU(ctx, gomock.Any()).Return(nil)

		sku, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{
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
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().CreateProductSKU(ctx, gomock.Any()).Return(nil)

		sku, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{})
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

	t.Run("trims the spec keys and values", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().CreateProductSKU(ctx, gomock.Any()).Return(nil)

		sku, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{
			Specs: map[string]string{" color ": " black "},
		})
		if err != nil {
			t.Fatalf("create sku: %v", err)
		}
		if got := sku.Specs["color"]; got != "black" {
			t.Fatalf("spec value = %q, want %q", got, "black")
		}
		if _, ok := sku.Specs[" color "]; ok {
			t.Fatalf("spec key was not trimmed: %v", sku.Specs)
		}
	})

	t.Run("rejects invalid specs", func(t *testing.T) {
		tests := map[string]map[string]string{
			"empty key":   {"": "black"},
			"empty value": {"color": " "},
			"long key":    {strings.Repeat("k", 21): "black"},
			"long value":  {"color": strings.Repeat("v", 51)},
			"too many specs": {
				"a": "1", "b": "2", "c": "3", "d": "4", "e": "5", "f": "6",
			},
			"duplicate after trim": {"color": "black", " color ": "white"},
		}
		for name, specs := range tests {
			t.Run(name, func(t *testing.T) {
				skuService, productReader, _ := newTestProductSKUService(t)

				productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)

				_, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{Specs: specs})
				requireErrorIs(t, err, apperr.ErrInvalidProductSKUSpecs)
			})
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)

		productReader.EXPECT().
			GetProductByID(ctx, 99).
			Return(nil, fmt.Errorf("product repository: %w", apperr.ErrProductNotFound))

		_, err := skuService.Create(ctx, testProduct.ShopID, 99, &dto.CreateProductSKURequest{
			Specs: map[string]string{"color": "black"},
		})
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("rejects a product owned by another shop", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)
		otherShopProduct := &model.Product{ID: 1, ShopID: testProduct.ShopID + 1, Name: "T-shirt"}

		productReader.EXPECT().GetProductByID(ctx, 1).Return(otherShopProduct, nil)

		_, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{
			Specs: map[string]string{"color": "black"},
		})
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("duplicate specs", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			CreateProductSKU(ctx, gomock.Any()).
			Return(fmt.Errorf("sku repository: %w", apperr.ErrProductSKUAlreadyExists))

		_, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{
			Specs:      map[string]string{"color": "black", "size": "M"},
			PriceCents: 1999,
		})
		requireErrorIs(t, err, apperr.ErrProductSKUAlreadyExists)
	})

	t.Run("rejects spec keys that differ from existing skus", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			CreateProductSKU(ctx, gomock.Any()).
			Return(fmt.Errorf("sku repository: %w", apperr.ErrProductSKUSpecsMismatch))

		_, err := skuService.Create(ctx, testProduct.ShopID, 1, &dto.CreateProductSKURequest{
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

		got, err := skuService.List(ctx, testProduct.ShopID, 1)
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

		_, err := skuService.List(ctx, testProduct.ShopID, 99)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("does not expose a product owned by another shop", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)
		otherShopProduct := &model.Product{ID: 1, ShopID: testProduct.ShopID + 1, Name: "T-shirt"}

		productReader.EXPECT().GetProductByID(ctx, 1).Return(otherShopProduct, nil)

		_, err := skuService.List(ctx, testProduct.ShopID, 1)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})
}

func TestUpdateProductSKU(t *testing.T) {
	ctx := context.Background()

	t.Run("updates only the provided fields", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)
		stock := 42

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			UpdateProductSKU(ctx, 1, 5, gomock.Any()).
			DoAndReturn(func(_ context.Context, productID, skuID int, update *model.ProductSKUUpdate) (*model.ProductSKU, error) {
				if productID != 1 || skuID != 5 {
					t.Fatalf("product/sku = %d/%d, want 1/5", productID, skuID)
				}
				if update.Stock == nil || *update.Stock != 42 {
					t.Fatalf("stock update = %v, want 42", update.Stock)
				}
				if update.PriceCents != nil || update.Specs != nil || update.IsEnabled != nil {
					t.Fatalf("update = %+v, want only the stock", update)
				}
				return &model.ProductSKU{ID: 5, ProductID: 1, Specs: map[string]string{"color": "black"}, PriceCents: 1999, Stock: 42, IsEnabled: true}, nil
			})

		got, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{Stock: &stock})
		if err != nil {
			t.Fatalf("update sku: %v", err)
		}
		if got.Stock != 42 {
			t.Fatalf("stock = %d, want 42", got.Stock)
		}
		if got.PriceCents != 1999 || got.Specs["color"] != "black" || !got.IsEnabled {
			t.Fatalf("sku = %+v, want unchanged fields", got)
		}
	})

	t.Run("normalizes the specs", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			UpdateProductSKU(ctx, 1, 5, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ int, update *model.ProductSKUUpdate) (*model.ProductSKU, error) {
				if update.Specs == nil || update.Specs["color"] != "black" {
					t.Fatalf("specs = %v, want color=black", update.Specs)
				}
				if _, ok := update.Specs[" color "]; ok {
					t.Fatalf("spec key was not trimmed: %v", update.Specs)
				}
				return &model.ProductSKU{ID: 5, ProductID: 1, Specs: update.Specs}, nil
			})

		if _, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{
			Specs: map[string]string{" color ": " black "},
		}); err != nil {
			t.Fatalf("update sku: %v", err)
		}
	})

	t.Run("updates all fields", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)
		price := int64(2999)
		stock := 0
		enabled := false

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			UpdateProductSKU(ctx, 1, 5, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ int, update *model.ProductSKUUpdate) (*model.ProductSKU, error) {
				if update.PriceCents == nil || *update.PriceCents != 2999 {
					t.Fatalf("price update = %v, want 2999", update.PriceCents)
				}
				if update.Stock == nil || *update.Stock != 0 {
					t.Fatalf("stock update = %v, want 0", update.Stock)
				}
				if update.IsEnabled == nil || *update.IsEnabled {
					t.Fatalf("enabled update = %v, want false", update.IsEnabled)
				}
				if update.Specs == nil || update.Specs["size"] != "L" {
					t.Fatalf("specs = %v, want size=L", update.Specs)
				}
				return &model.ProductSKU{ID: 5, ProductID: 1, Specs: update.Specs, PriceCents: 2999, Stock: 0, IsEnabled: false}, nil
			})

		got, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{
			Specs:      map[string]string{"size": "L"},
			PriceCents: &price,
			Stock:      &stock,
			IsEnabled:  &enabled,
		})
		if err != nil {
			t.Fatalf("update sku: %v", err)
		}
		if got.PriceCents != 2999 || got.Stock != 0 || got.IsEnabled {
			t.Fatalf("sku = %+v, want updated fields", got)
		}
	})

	t.Run("rejects an empty update", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)
		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)

		_, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{})
		requireErrorIs(t, err, apperr.ErrInvalidProductSKUUpdate)
	})

	t.Run("rejects invalid specs", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)

		_, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{
			Specs: map[string]string{"": "black"},
		})
		requireErrorIs(t, err, apperr.ErrInvalidProductSKUSpecs)
	})

	t.Run("unknown sku", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			UpdateProductSKU(ctx, 1, 99, gomock.Any()).
			Return(nil, fmt.Errorf("sku repository: %w", apperr.ErrProductSKUNotFound))

		stock := 1
		_, err := skuService.Update(ctx, testProduct.ShopID, 1, 99, &dto.UpdateProductSKURequest{Stock: &stock})
		requireErrorIs(t, err, apperr.ErrProductSKUNotFound)
	})

	t.Run("does not expose a product owned by another shop", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)
		otherShopProduct := &model.Product{ID: 1, ShopID: testProduct.ShopID + 1, Name: "T-shirt"}

		productReader.EXPECT().GetProductByID(ctx, 1).Return(otherShopProduct, nil)

		stock := 1
		_, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{Stock: &stock})
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("rejects spec keys that differ from the other skus", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			UpdateProductSKU(ctx, 1, 5, gomock.Any()).
			Return(nil, fmt.Errorf("sku repository: %w", apperr.ErrProductSKUSpecsMismatch))

		_, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{
			Specs: map[string]string{"size": "L"},
		})
		requireErrorIs(t, err, apperr.ErrProductSKUSpecsMismatch)
	})

	t.Run("duplicate specs", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			UpdateProductSKU(ctx, 1, 5, gomock.Any()).
			Return(nil, fmt.Errorf("sku repository: %w", apperr.ErrProductSKUAlreadyExists))

		stock := 0
		_, err := skuService.Update(ctx, testProduct.ShopID, 1, 5, &dto.UpdateProductSKURequest{Stock: &stock})
		requireErrorIs(t, err, apperr.ErrProductSKUAlreadyExists)
	})
}
func TestDeleteProductSKU(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes the SKU", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().DeleteProductSKU(ctx, 1, 5).Return(nil)

		if err := skuService.Delete(ctx, testProduct.ShopID, 1, 5); err != nil {
			t.Fatalf("delete sku: %v", err)
		}
	})

	t.Run("unknown product", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)

		productReader.EXPECT().
			GetProductByID(ctx, 99).
			Return(nil, fmt.Errorf("product repository: %w", apperr.ErrProductNotFound))

		err := skuService.Delete(ctx, testProduct.ShopID, 99, 5)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("does not expose a product owned by another shop", func(t *testing.T) {
		skuService, productReader, _ := newTestProductSKUService(t)
		otherShopProduct := &model.Product{ID: 1, ShopID: testProduct.ShopID + 1, Name: "T-shirt"}

		productReader.EXPECT().GetProductByID(ctx, 1).Return(otherShopProduct, nil)

		err := skuService.Delete(ctx, testProduct.ShopID, 1, 5)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("unknown sku", func(t *testing.T) {
		skuService, productReader, skuRepository := newTestProductSKUService(t)

		productReader.EXPECT().GetProductByID(ctx, 1).Return(testProduct, nil)
		skuRepository.EXPECT().
			DeleteProductSKU(ctx, 1, 99).
			Return(fmt.Errorf("sku repository: %w", apperr.ErrProductSKUNotFound))

		err := skuService.Delete(ctx, testProduct.ShopID, 1, 99)
		requireErrorIs(t, err, apperr.ErrProductSKUNotFound)
	})
}
