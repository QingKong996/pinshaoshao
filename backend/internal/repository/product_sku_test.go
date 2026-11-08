package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/repository"
)

func TestProductSKURepository(t *testing.T) {
	ctx := context.Background()
	skuRepository := repository.NewProductSKURepository(testPool)

	t.Run("creates a product sku", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)

		sku := &model.ProductSKU{
			ProductID:  product.ID,
			Specs:      map[string]string{"color": "black", "size": "M"},
			PriceCents: 1999,
			Stock:      100,
			IsEnabled:  true,
		}
		if err := skuRepository.CreateProductSKU(ctx, sku); err != nil {
			t.Fatalf("create product sku: %v", err)
		}
		if sku.ID == 0 {
			t.Fatal("product sku ID was not assigned")
		}
		if sku.CreatedAt.IsZero() || sku.UpdatedAt.IsZero() {
			t.Fatalf("timestamps are not set: %+v", sku)
		}

		skus, err := skuRepository.ListProductSKUsByProductID(ctx, product.ID)
		if err != nil {
			t.Fatalf("list product skus: %v", err)
		}
		if len(skus) != 1 {
			t.Fatalf("product sku count = %d, want 1", len(skus))
		}
		got := skus[0]
		if got.ProductID != product.ID || got.PriceCents != 1999 || got.Stock != 100 || !got.IsEnabled {
			t.Fatalf("product sku = %+v, want created values", got)
		}
		if got.Specs["color"] != "black" || got.Specs["size"] != "M" {
			t.Fatalf("specs = %v, want color=black size=M", got.Specs)
		}
	})

	t.Run("lists the skus of a product ordered by id", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		otherProduct := insertProduct(t, shop.ID)

		specsByProduct := map[int][]map[string]string{
			product.ID:      {{"color": "black"}, {"color": "white"}},
			otherProduct.ID: {{"color": "black"}},
		}
		for productID, specs := range specsByProduct {
			for _, spec := range specs {
				sku := &model.ProductSKU{ProductID: productID, Specs: spec}
				if err := skuRepository.CreateProductSKU(ctx, sku); err != nil {
					t.Fatalf("create product sku: %v", err)
				}
			}
		}

		skus, err := skuRepository.ListProductSKUsByProductID(ctx, product.ID)
		if err != nil {
			t.Fatalf("list product skus: %v", err)
		}
		if len(skus) != 2 {
			t.Fatalf("product sku count = %d, want 2", len(skus))
		}
		for i, want := range []string{"black", "white"} {
			if skus[i].Specs["color"] != want {
				t.Fatalf("product sku %d color = %q, want %q", i, skus[i].Specs["color"], want)
			}
		}
	})

	t.Run("rejects duplicate specs", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		specs := map[string]string{"color": "black"}

		if err := skuRepository.CreateProductSKU(ctx, &model.ProductSKU{ProductID: product.ID, Specs: specs}); err != nil {
			t.Fatalf("create product sku: %v", err)
		}

		err := skuRepository.CreateProductSKU(ctx, &model.ProductSKU{ProductID: product.ID, Specs: specs})
		requireErrorIs(t, err, apperr.ErrProductSKUAlreadyExists)
	})

	t.Run("rejects an unknown product", func(t *testing.T) {
		resetDatabase(t)

		err := skuRepository.CreateProductSKU(ctx, &model.ProductSKU{
			ProductID: 404,
			Specs:     map[string]string{"color": "black"},
		})
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})
}
