//go:build integration

package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/feature/product/apperr"
	"pinshaoshao/internal/feature/product/model"
	"pinshaoshao/internal/feature/product/repository"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	shoprepository "pinshaoshao/internal/feature/shop/repository"
)

func TestProductRepository(t *testing.T) {
	ctx := context.Background()
	productRepository := repository.NewProductRepository(testPool)

	t.Run("creates a product", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)

		product := &model.Product{
			ShopID:      shop.ID,
			Name:        "T-shirt",
			Description: "Cotton T-shirt",
			Status:      model.ProductStatusOffShelf,
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

	t.Run("does not expose a product summary from another shop", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")
		otherProduct := insertProduct(t, otherShop.ID)

		_, err := productRepository.GetProductSummaryByID(ctx, shop.ID, otherProduct.ID)
		requireErrorIs(t, err, apperr.ErrProductNotFound)
	})

	t.Run("deletes a product and its skus", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		skuRepository := repository.NewProductSKURepository(testPool)
		createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "black"}, 1999)

		if err := productRepository.DeleteProduct(ctx, shop.ID, product.ID); err != nil {
			t.Fatalf("delete product: %v", err)
		}

		_, err := productRepository.GetProductByID(ctx, product.ID)
		requireErrorIs(t, err, apperr.ErrProductNotFound)

		skus, err := skuRepository.ListProductSKUsByProductID(ctx, product.ID)
		if err != nil {
			t.Fatalf("list product skus: %v", err)
		}
		if len(skus) != 0 {
			t.Fatalf("product sku count = %d, want 0", len(skus))
		}
	})

	t.Run("rejects deleting a product with inventory", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		product := insertProduct(t, shop.ID)
		skuRepository := repository.NewProductSKURepository(testPool)
		sku := createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "black"}, 1999)

		if _, err := testPool.Exec(ctx, `UPDATE inventory_items SET total_stock = 3 WHERE sku_id = $1`, sku.ID); err != nil {
			t.Fatalf("set inventory: %v", err)
		}

		err := productRepository.DeleteProduct(ctx, shop.ID, product.ID)
		requireErrorIs(t, err, apperr.ErrProductHasInventory)
	})

	t.Run("does not delete a product from another shop", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")
		otherProduct := insertProduct(t, otherShop.ID)

		err := productRepository.DeleteProduct(ctx, shop.ID, otherProduct.ID)
		requireErrorIs(t, err, apperr.ErrProductNotFound)

		if _, err := productRepository.GetProductByID(ctx, otherProduct.ID); err != nil {
			t.Fatalf("get other shop product after rejected delete: %v", err)
		}
	})

	t.Run("does not update a product from another shop", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)
		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")
		otherProduct := insertProduct(t, otherShop.ID)

		err := productRepository.UpdateProduct(ctx, &model.Product{
			ID:         otherProduct.ID,
			ShopID:     shop.ID,
			Name:       "Stolen",
			Status:     model.ProductStatusOnShelf,
			CategoryID: nil,
		})
		requireErrorIs(t, err, apperr.ErrProductNotFound)

		persisted, err := productRepository.GetProductByID(ctx, otherProduct.ID)
		if err != nil {
			t.Fatalf("get other shop product after rejected update: %v", err)
		}
		if persisted.Name != otherProduct.Name || persisted.Status != otherProduct.Status {
			t.Fatalf("product = %+v, want unchanged values", persisted)
		}
	})
}

func TestProductSummaryList(t *testing.T) {
	ctx := context.Background()
	productRepository := repository.NewProductRepository(testPool)
	categoryRepository := repository.NewProductCategoryRepository(testPool)
	skuRepository := repository.NewProductSKURepository(testPool)

	setupProducts := func(t *testing.T) (*shopmodel.Shop, *model.Product, *model.Product, *model.Product) {
		t.Helper()

		resetDatabase(t)
		shop := insertShop(t)

		categoryZ := &model.ProductCategory{ShopID: shop.ID, Name: "z-category"}
		if err := categoryRepository.CreateProductCategory(ctx, categoryZ); err != nil {
			t.Fatalf("create category: %v", err)
		}
		categoryA := &model.ProductCategory{ShopID: shop.ID, Name: "a-category"}
		if err := categoryRepository.CreateProductCategory(ctx, categoryA); err != nil {
			t.Fatalf("create category: %v", err)
		}

		createProduct := func(name, description string, status model.ProductStatus, categoryID *int) *model.Product {
			product := &model.Product{
				ShopID:      shop.ID,
				Name:        name,
				Description: description,
				Status:      status,
				CategoryID:  categoryID,
			}
			if err := productRepository.CreateProduct(ctx, product); err != nil {
				t.Fatalf("create product %q: %v", name, err)
			}
			return product
		}
		createSKU := func(product *model.Product, spec string, price int64) {
			sku := &model.ProductSKU{
				ProductID:  product.ID,
				Specs:      map[string]string{"variant": spec},
				PriceCents: price,
				IsEnabled:  true,
			}
			if err := skuRepository.CreateProductSKU(ctx, sku); err != nil {
				t.Fatalf("create sku for %q: %v", product.Name, err)
			}
		}

		alpha := createProduct("Alpha", "cotton shirt", model.ProductStatusOnShelf, &categoryZ.ID)
		bravo := createProduct("Bravo", "warm hoodie", model.ProductStatusOffShelf, &categoryA.ID)
		charlie := createProduct("Charlie", "cotton pants", model.ProductStatusOnShelf, nil)

		createSKU(alpha, "1", 1000)
		createSKU(bravo, "1", 500)
		createSKU(bravo, "2", 800)
		createSKU(charlie, "1", 2000)

		otherShop := &shopmodel.Shop{
			Username:     "shop02",
			PasswordHash: testShopPassword,
			Email:        "shop02@example.com",
		}
		if err := shoprepository.NewShopRepository(testPool).CreateShopWithSession(ctx, otherShop, testTokenHash(otherShop.Username)); err != nil {
			t.Fatalf("create other shop: %v", err)
		}
		if err := productRepository.CreateProduct(ctx, &model.Product{
			ShopID: otherShop.ID,
			Name:   "Zulu",
			Status: model.ProductStatusOnShelf,
		}); err != nil {
			t.Fatalf("create other shop product: %v", err)
		}

		return shop, alpha, bravo, charlie
	}

	t.Run("filters and paginates summaries", func(t *testing.T) {
		shop, _, _, charlie := setupProducts(t)

		result, err := productRepository.ListProductSummaries(ctx, shop.ID, model.ProductListFilter{
			Sort:     model.ProductSortName,
			Order:    model.ProductSortOrderAsc,
			Page:     2,
			PageSize: 2,
		})
		if err != nil {
			t.Fatalf("list products: %v", err)
		}
		if result.Total != 3 || len(result.Products) != 1 {
			t.Fatalf("total = %d, products = %d, want total 3 and one product", result.Total, len(result.Products))
		}
		if result.Products[0].ID != charlie.ID {
			t.Fatalf("product = %q, want Charlie", result.Products[0].Name)
		}

		result, err = productRepository.ListProductSummaries(ctx, shop.ID, model.ProductListFilter{
			Search: "COTTON",
			Status: model.ProductStatusOnShelf,
			Sort:   model.ProductSortCreatedAt,
			Order:  model.ProductSortOrderDesc,
			Page:   1, PageSize: 10,
		})
		if err != nil {
			t.Fatalf("filter products: %v", err)
		}
		if result.Total != 2 || len(result.Products) != 2 {
			t.Fatalf("filtered products = %+v, want two cotton on-shelf products", result.Products)
		}
		if result.Products[0].Name != "Charlie" || result.Products[1].Name != "Alpha" {
			t.Fatalf("filtered products = %+v, want Charlie then Alpha", result.Products)
		}
	})

	t.Run("supports category and uncategorized filters", func(t *testing.T) {
		shop, alpha, bravo, charlie := setupProducts(t)

		result, err := productRepository.ListProductSummaries(ctx, shop.ID, model.ProductListFilter{
			CategoryID: alpha.CategoryID,
			Sort:       model.ProductSortName,
			Order:      model.ProductSortOrderAsc,
			Page:       1,
			PageSize:   10,
		})
		if err != nil {
			t.Fatalf("filter by category: %v", err)
		}
		if result.Total != 1 || len(result.Products) != 1 || result.Products[0].ID != alpha.ID {
			t.Fatalf("category products = %+v, want Alpha", result.Products)
		}

		result, err = productRepository.ListProductSummaries(ctx, shop.ID, model.ProductListFilter{
			Uncategorized: true,
			Sort:          model.ProductSortName,
			Order:         model.ProductSortOrderAsc,
			Page:          1,
			PageSize:      10,
		})
		if err != nil {
			t.Fatalf("filter uncategorized: %v", err)
		}
		if result.Total != 1 || len(result.Products) != 1 || result.Products[0].ID != charlie.ID {
			t.Fatalf("uncategorized products = %+v, want Charlie", result.Products)
		}

		result, err = productRepository.ListProductSummaries(ctx, shop.ID, model.ProductListFilter{
			Status:   model.ProductStatusOffShelf,
			Sort:     model.ProductSortName,
			Order:    model.ProductSortOrderAsc,
			Page:     1,
			PageSize: 10,
		})
		if err != nil {
			t.Fatalf("filter by status: %v", err)
		}
		if result.Total != 1 || len(result.Products) != 1 || result.Products[0].ID != bravo.ID {
			t.Fatalf("off-shelf products = %+v, want Bravo", result.Products)
		}
	})

	t.Run("sorts by product fields and aggregates", func(t *testing.T) {
		shop, _, _, _ := setupProducts(t)

		tests := []struct {
			name  string
			sort  model.ProductSort
			order model.ProductSortOrder
			want  []string
		}{
			{"name ascending", model.ProductSortName, model.ProductSortOrderAsc, []string{"Alpha", "Bravo", "Charlie"}},
			{"name descending", model.ProductSortName, model.ProductSortOrderDesc, []string{"Charlie", "Bravo", "Alpha"}},
			{"created newest first", model.ProductSortCreatedAt, model.ProductSortOrderDesc, []string{"Charlie", "Bravo", "Alpha"}},
			{"sku count descending", model.ProductSortSKUCount, model.ProductSortOrderDesc, []string{"Bravo", "Charlie", "Alpha"}},
			{"price ascending", model.ProductSortMinPrice, model.ProductSortOrderAsc, []string{"Bravo", "Alpha", "Charlie"}},
			{"category ascending", model.ProductSortCategory, model.ProductSortOrderAsc, []string{"Bravo", "Alpha", "Charlie"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := productRepository.ListProductSummaries(ctx, shop.ID, model.ProductListFilter{
					Sort:     tt.sort,
					Order:    tt.order,
					Page:     1,
					PageSize: 10,
				})
				if err != nil {
					t.Fatalf("list products: %v", err)
				}
				if len(result.Products) != len(tt.want) {
					t.Fatalf("product count = %d, want %d", len(result.Products), len(tt.want))
				}
				for i, want := range tt.want {
					if result.Products[i].Name != want {
						t.Fatalf("product %d = %q, want %q", i, result.Products[i].Name, want)
					}
				}
			})
		}
	})
}
