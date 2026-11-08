package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/repository"
)

func TestProductRepository(t *testing.T) {
	ctx := context.Background()
	productRepository := repository.NewProductRepository(testPool)
	shopRepository := repository.NewShopRepository(testPool)

	t.Run("creates a product", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)

		product := &model.Product{
			ShopID:      shop.ID,
			Name:        "T-shirt",
			Description: "Cotton T-shirt",
		}
		if err := productRepository.CreateProduct(ctx, product); err != nil {
			t.Fatalf("create product: %v", err)
		}
		if product.ID == 0 {
			t.Fatal("product ID was not assigned")
		}

		got, err := productRepository.GetProductByID(ctx, product.ID)
		if err != nil {
			t.Fatalf("get product: %v", err)
		}
		if got.ShopID != shop.ID {
			t.Fatalf("shop ID = %d, want %d", got.ShopID, shop.ID)
		}
		if got.Name != product.Name || got.Description != product.Description {
			t.Fatalf("product = %+v, want %+v", got, product)
		}
		if got.Status != model.ProductStatusOffShelf {
			t.Fatalf("status = %q, want %q", got.Status, model.ProductStatusOffShelf)
		}
		if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
			t.Fatalf("timestamps are not set: %+v", got)
		}
	})

	t.Run("returns not found for an unknown product", func(t *testing.T) {
		resetDatabase(t)

		_, err := productRepository.GetProductByID(ctx, 404)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("lists the products of a shop ordered by id", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		otherShop := &model.Shop{
			Username:     "shop02",
			PasswordHash: testShopPassword,
			Email:        "shop02@example.com",
		}
		if err := shopRepository.CreateShop(ctx, otherShop); err != nil {
			t.Fatalf("create other shop: %v", err)
		}

		for _, name := range []string{"first", "second"} {
			if err := productRepository.CreateProduct(ctx, &model.Product{ShopID: shop.ID, Name: name}); err != nil {
				t.Fatalf("create product %q: %v", name, err)
			}
		}
		if err := productRepository.CreateProduct(ctx, &model.Product{ShopID: otherShop.ID, Name: "other"}); err != nil {
			t.Fatalf("create other product: %v", err)
		}

		products, err := productRepository.ListProductsByShopID(ctx, shop.ID)
		if err != nil {
			t.Fatalf("list products: %v", err)
		}
		if len(products) != 2 {
			t.Fatalf("product count = %d, want 2", len(products))
		}
		for i, want := range []string{"first", "second"} {
			if products[i].Name != want {
				t.Fatalf("product %d name = %q, want %q", i, products[i].Name, want)
			}
		}
	})

	t.Run("updates a product", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)

		product.Name = "Hoodie"
		product.Description = "Warm hoodie"
		product.Status = model.ProductStatusOnShelf
		if err := productRepository.UpdateProduct(ctx, product); err != nil {
			t.Fatalf("update product: %v", err)
		}

		got, err := productRepository.GetProductByID(ctx, product.ID)
		if err != nil {
			t.Fatalf("get product: %v", err)
		}
		if got.Name != "Hoodie" || got.Description != "Warm hoodie" {
			t.Fatalf("product = %+v, want updated fields", got)
		}
		if got.Status != model.ProductStatusOnShelf {
			t.Fatalf("status = %q, want %q", got.Status, model.ProductStatusOnShelf)
		}
	})

	t.Run("returns not found when updating an unknown product", func(t *testing.T) {
		resetDatabase(t)

		err := productRepository.UpdateProduct(ctx, &model.Product{
			ID:     404,
			Name:   "ghost",
			Status: model.ProductStatusOffShelf,
		})
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})
}
