package repository_test

import (
	"context"
	"errors"
	"sync"
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

	t.Run("rejects spec keys that differ from existing skus", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)

		if err := skuRepository.CreateProductSKU(ctx, &model.ProductSKU{
			ProductID: product.ID,
			Specs:     map[string]string{"color": "black"},
		}); err != nil {
			t.Fatalf("create product sku: %v", err)
		}

		err := skuRepository.CreateProductSKU(ctx, &model.ProductSKU{
			ProductID: product.ID,
			Specs:     map[string]string{"size": "M"},
		})
		requireErrorIs(t, err, apperr.ErrProductSKUSpecsMismatch)
	})

	t.Run("serializes concurrent creation of the first skus", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)

		specs := []map[string]string{
			{"color": "black"},
			{"size": "M"},
		}
		errs := make([]error, len(specs))
		var wg sync.WaitGroup
		for i := range specs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				errs[i] = skuRepository.CreateProductSKU(ctx, &model.ProductSKU{
					ProductID: product.ID,
					Specs:     specs[i],
				})
			}(i)
		}
		wg.Wait()

		var created, mismatched int
		for _, err := range errs {
			switch {
			case err == nil:
				created++
			case errors.Is(err, apperr.ErrProductSKUSpecsMismatch):
				mismatched++
			default:
				t.Fatalf("unexpected create error: %v", err)
			}
		}
		if created != 1 || mismatched != 1 {
			t.Fatalf("created = %d, mismatched = %d, want 1 and 1", created, mismatched)
		}

		skus, err := skuRepository.ListProductSKUsByProductID(ctx, product.ID)
		if err != nil {
			t.Fatalf("list product skus: %v", err)
		}
		if len(skus) != 1 {
			t.Fatalf("product sku count = %d, want 1", len(skus))
		}
	})
}

func TestProductSKUUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	skuRepository := repository.NewProductSKURepository(testPool)

	t.Run("updates only the provided fields", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		sku := createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "black"}, 1999, 3)

		stock := 42
		got, err := skuRepository.UpdateProductSKU(ctx, product.ID, sku.ID, &model.ProductSKUUpdate{Stock: &stock})
		if err != nil {
			t.Fatalf("update product sku: %v", err)
		}
		if got.Stock != 42 {
			t.Fatalf("stock = %d, want 42", got.Stock)
		}
		if got.PriceCents != 1999 || got.Specs["color"] != "black" || !got.IsEnabled {
			t.Fatalf("product sku = %+v, want unchanged fields", got)
		}
	})

	t.Run("rejects spec keys that differ from the other skus", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		first := createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "black"}, 1999, 3)
		createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "white"}, 1999, 3)

		specs := map[string]string{"size": "M"}
		_, err := skuRepository.UpdateProductSKU(ctx, product.ID, first.ID, &model.ProductSKUUpdate{Specs: specs})
		requireErrorIs(t, err, apperr.ErrProductSKUSpecsMismatch)
	})

	t.Run("allows changing the spec keys of the only sku", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		sku := createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "black"}, 1999, 3)

		specs := map[string]string{"size": "M"}
		got, err := skuRepository.UpdateProductSKU(ctx, product.ID, sku.ID, &model.ProductSKUUpdate{Specs: specs})
		if err != nil {
			t.Fatalf("update product sku: %v", err)
		}
		if got.Specs["size"] != "M" {
			t.Fatalf("specs = %v, want size=M", got.Specs)
		}
	})

	t.Run("rejects an unknown sku", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)

		stock := 1
		_, err := skuRepository.UpdateProductSKU(ctx, product.ID, 404, &model.ProductSKUUpdate{Stock: &stock})
		requireErrorIs(t, err, apperr.ErrProductSKUNotFound)
	})

	t.Run("deletes the sku", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		sku := createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "black"}, 1999, 3)

		if err := skuRepository.DeleteProductSKU(ctx, product.ID, sku.ID); err != nil {
			t.Fatalf("delete product sku: %v", err)
		}
		skus, err := skuRepository.ListProductSKUsByProductID(ctx, product.ID)
		if err != nil {
			t.Fatalf("list product skus: %v", err)
		}
		if len(skus) != 0 {
			t.Fatalf("product sku count = %d, want 0", len(skus))
		}
	})

	t.Run("rejects deleting an unknown sku", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)

		requireErrorIs(t, skuRepository.DeleteProductSKU(ctx, product.ID, 404), apperr.ErrProductSKUNotFound)
	})
}

func createTestSKU(t *testing.T, ctx context.Context, skuRepository *repository.ProductSKURepository, productID int, specs map[string]string, priceCents int64, stock int) *model.ProductSKU {
	t.Helper()

	sku := &model.ProductSKU{
		ProductID:  productID,
		Specs:      specs,
		PriceCents: priceCents,
		Stock:      stock,
		IsEnabled:  true,
	}
	if err := skuRepository.CreateProductSKU(ctx, sku); err != nil {
		t.Fatalf("create product sku: %v", err)
	}
	return sku
}
