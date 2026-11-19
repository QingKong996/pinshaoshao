//go:build integration

package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/feature/product/apperr"
	"pinshaoshao/internal/feature/product/model"
	"pinshaoshao/internal/feature/product/repository"
)

func TestProductCategoryRepository(t *testing.T) {
	ctx := context.Background()
	categoryRepository := repository.NewProductCategoryRepository(testPool)
	productRepository := repository.NewProductRepository(testPool)

	createCategory := func(t *testing.T, shopID int, name string) *model.ProductCategory {
		t.Helper()

		category := &model.ProductCategory{ShopID: shopID, Name: name}
		if err := categoryRepository.CreateProductCategory(ctx, category); err != nil {
			t.Fatalf("create category %q: %v", name, err)
		}
		return category
	}

	createProduct := func(t *testing.T, shopID int, name string, categoryID *int) *model.Product {
		t.Helper()

		product := &model.Product{
			ShopID:     shopID,
			Name:       name,
			Status:     model.ProductStatusOffShelf,
			CategoryID: categoryID,
		}
		if err := productRepository.CreateProduct(ctx, product); err != nil {
			t.Fatalf("create product %q: %v", name, err)
		}
		return product
	}

	t.Run("lists categories with product counts", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)

		shoes := createCategory(t, shop.ID, "shoes")
		clothes := createCategory(t, shop.ID, "clothes")
		createCategory(t, shop.ID, "accessories")
		createProduct(t, shop.ID, "T-shirt", &clothes.ID)
		createProduct(t, shop.ID, "Hoodie", &clothes.ID)
		createProduct(t, shop.ID, "Sneakers", &shoes.ID)

		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")
		createCategory(t, otherShop.ID, "other shop category")

		categories, err := categoryRepository.ListProductCategoriesByShopID(ctx, shop.ID)
		if err != nil {
			t.Fatalf("list categories: %v", err)
		}

		wantNames := []string{"accessories", "clothes", "shoes"}
		wantCounts := []int64{0, 2, 1}
		if len(categories) != len(wantNames) {
			t.Fatalf("category count = %d, want %d", len(categories), len(wantNames))
		}
		for i := range categories {
			if categories[i].Name != wantNames[i] {
				t.Fatalf("category %d name = %q, want %q", i, categories[i].Name, wantNames[i])
			}
			if categories[i].ProductCount != wantCounts[i] {
				t.Fatalf("category %q product count = %d, want %d", categories[i].Name, categories[i].ProductCount, wantCounts[i])
			}
			if categories[i].ShopID != shop.ID {
				t.Fatalf("category %q shop ID = %d, want %d", categories[i].Name, categories[i].ShopID, shop.ID)
			}
		}

		got, err := categoryRepository.GetProductCategoryByID(ctx, shop.ID, clothes.ID)
		if err != nil {
			t.Fatalf("get category: %v", err)
		}
		if got.ProductCount != 2 {
			t.Fatalf("product count = %d, want 2", got.ProductCount)
		}
	})

	t.Run("rejects duplicate names within a shop", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		createCategory(t, shop.ID, "clothes")

		err := categoryRepository.CreateProductCategory(ctx, &model.ProductCategory{ShopID: shop.ID, Name: "clothes"})
		requireErrorIs(t, err, apperr.ErrCategoryNameTaken)

		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")
		createCategory(t, otherShop.ID, "clothes")
	})

	t.Run("renames a category", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		createCategory(t, shop.ID, "clothes")
		shoes := createCategory(t, shop.ID, "shoes")

		err := categoryRepository.UpdateProductCategory(ctx, &model.ProductCategory{
			ID:     shoes.ID,
			ShopID: shop.ID,
			Name:   "clothes",
		})
		requireErrorIs(t, err, apperr.ErrCategoryNameTaken)

		shoes.Name = "footwear"
		if err := categoryRepository.UpdateProductCategory(ctx, shoes); err != nil {
			t.Fatalf("rename category: %v", err)
		}

		got, err := categoryRepository.GetProductCategoryByID(ctx, shop.ID, shoes.ID)
		if err != nil {
			t.Fatalf("get renamed category: %v", err)
		}
		if got.Name != "footwear" {
			t.Fatalf("category name = %q, want footwear", got.Name)
		}
	})

	t.Run("does not mutate a category from another shop", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		category := createCategory(t, shop.ID, "clothes")

		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")

		_, err := categoryRepository.GetProductCategoryByID(ctx, otherShop.ID, category.ID)
		requireErrorIs(t, err, apperr.ErrCategoryNotFound)

		err = categoryRepository.UpdateProductCategory(ctx, &model.ProductCategory{
			ID:     category.ID,
			ShopID: otherShop.ID,
			Name:   "stolen",
		})
		requireErrorIs(t, err, apperr.ErrCategoryNotFound)

		err = categoryRepository.DeleteProductCategory(ctx, otherShop.ID, category.ID)
		requireErrorIs(t, err, apperr.ErrCategoryNotFound)

		got, err := categoryRepository.GetProductCategoryByID(ctx, shop.ID, category.ID)
		if err != nil {
			t.Fatalf("get original category: %v", err)
		}
		if got.Name != "clothes" {
			t.Fatalf("category name = %q, want clothes", got.Name)
		}
	})

	t.Run("deletes a category and uncategorizes its products", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		category := createCategory(t, shop.ID, "clothes")
		product := createProduct(t, shop.ID, "T-shirt", &category.ID)

		if err := categoryRepository.DeleteProductCategory(ctx, shop.ID, category.ID); err != nil {
			t.Fatalf("delete category: %v", err)
		}

		got, err := productRepository.GetProductByID(ctx, product.ID)
		if err != nil {
			t.Fatalf("get product after category delete: %v", err)
		}
		if got.CategoryID != nil {
			t.Fatalf("product category ID = %v, want nil", got.CategoryID)
		}

		err = categoryRepository.DeleteProductCategory(ctx, shop.ID, category.ID)
		requireErrorIs(t, err, apperr.ErrCategoryNotFound)
	})
}
