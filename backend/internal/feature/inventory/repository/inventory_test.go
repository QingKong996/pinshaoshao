//go:build integration

package repository_test

import (
	"context"
	"testing"

	inventoryrepository "pinshaoshao/internal/feature/inventory/repository"
	"pinshaoshao/internal/feature/product/model"
	productrepository "pinshaoshao/internal/feature/product/repository"
)

func TestInventoryRepositoryList(t *testing.T) {
	ctx := context.Background()
	inventoryRepository := inventoryrepository.NewInventoryRepository(testPool)
	productRepository := productrepository.NewProductRepository(testPool)
	categoryRepository := productrepository.NewProductCategoryRepository(testPool)
	skuRepository := productrepository.NewProductSKURepository(testPool)

	t.Run("returns the SKUs of the shop ordered by stock", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)

		category := &model.ProductCategory{ShopID: shop.ID, Name: "clothes"}
		if err := categoryRepository.CreateProductCategory(ctx, category); err != nil {
			t.Fatalf("create category: %v", err)
		}

		tshirt := &model.Product{
			ShopID:     shop.ID,
			Name:       "T-shirt",
			Status:     model.ProductStatusOnShelf,
			CategoryID: &category.ID,
		}
		if err := productRepository.CreateProduct(ctx, tshirt); err != nil {
			t.Fatalf("create T-shirt: %v", err)
		}
		hoodie := &model.Product{
			ShopID: shop.ID,
			Name:   "Hoodie",
			Status: model.ProductStatusOffShelf,
		}
		if err := productRepository.CreateProduct(ctx, hoodie); err != nil {
			t.Fatalf("create Hoodie: %v", err)
		}

		tshirtSKU := createTestSKU(t, ctx, skuRepository, tshirt.ID, map[string]string{"color": "black"}, 1999, 4)
		hoodieSKU := createTestSKU(t, ctx, skuRepository, hoodie.ID, map[string]string{"size": "M"}, 2999, 1)

		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")
		otherProduct := &model.Product{
			ShopID: otherShop.ID,
			Name:   "Other shop product",
			Status: model.ProductStatusOnShelf,
		}
		if err := productRepository.CreateProduct(ctx, otherProduct); err != nil {
			t.Fatalf("create other shop product: %v", err)
		}
		createTestSKU(t, ctx, skuRepository, otherProduct.ID, map[string]string{"variant": "other"}, 100, 0)

		items, err := inventoryRepository.ListInventoryByShopID(ctx, shop.ID)
		if err != nil {
			t.Fatalf("list inventory: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("inventory count = %d, want 2", len(items))
		}

		if items[0].SKUID != hoodieSKU.ID || items[0].ProductName != "Hoodie" || items[0].Stock != 1 {
			t.Fatalf("first item = %+v, want the hoodie SKU", items[0])
		}
		if items[0].CategoryID != nil || items[0].CategoryName != nil {
			t.Fatalf("first item category = %+v, want uncategorized", items[0])
		}

		if items[1].SKUID != tshirtSKU.ID || items[1].ProductName != "T-shirt" || items[1].Stock != 4 {
			t.Fatalf("second item = %+v, want the T-shirt SKU", items[1])
		}
		if items[1].CategoryID == nil || *items[1].CategoryID != category.ID {
			t.Fatalf("second item category ID = %v, want %d", items[1].CategoryID, category.ID)
		}
		if items[1].CategoryName == nil || *items[1].CategoryName != "clothes" {
			t.Fatalf("second item category name = %v, want clothes", items[1].CategoryName)
		}
		if items[1].Specs["color"] != "black" || items[1].PriceCents != 1999 || !items[1].IsEnabled {
			t.Fatalf("second item = %+v, want persisted SKU fields", items[1])
		}
	})

	t.Run("returns an empty slice for a shop without SKUs", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)

		items, err := inventoryRepository.ListInventoryByShopID(ctx, shop.ID)
		if err != nil {
			t.Fatalf("list inventory: %v", err)
		}
		if items == nil {
			t.Fatal("inventory is nil, want an empty slice")
		}
		if len(items) != 0 {
			t.Fatalf("inventory count = %d, want 0", len(items))
		}
	})
}
