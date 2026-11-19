//go:build integration

package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/feature/inventory/apperr"
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

	t.Run("returns the inventory of the shop ordered by available stock", func(t *testing.T) {
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

		tshirtSKU := createTestSKU(t, ctx, skuRepository, tshirt.ID, map[string]string{"color": "black"}, 1999)
		setTestInventory(t, ctx, tshirtSKU.ID, 5, 1)
		hoodieSKU := createTestSKU(t, ctx, skuRepository, hoodie.ID, map[string]string{"size": "M"}, 2999)
		setTestInventory(t, ctx, hoodieSKU.ID, 2, 1)

		otherShop := insertShopWithUsername(t, ctx, "shop02", "shop02@example.com")
		otherProduct := &model.Product{
			ShopID: otherShop.ID,
			Name:   "Other shop product",
			Status: model.ProductStatusOnShelf,
		}
		if err := productRepository.CreateProduct(ctx, otherProduct); err != nil {
			t.Fatalf("create other shop product: %v", err)
		}
		otherSKU := createTestSKU(t, ctx, skuRepository, otherProduct.ID, map[string]string{"variant": "other"}, 100)
		setTestInventory(t, ctx, otherSKU.ID, 10, 0)

		items, err := inventoryRepository.ListInventoryByShopID(ctx, shop.ID)
		if err != nil {
			t.Fatalf("list inventory: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("inventory count = %d, want 2", len(items))
		}

		if items[0].SKUID != hoodieSKU.ID || items[0].ProductName != "Hoodie" ||
			items[0].TotalStock != 2 || items[0].FrozenStock != 1 || items[0].AvailableStock() != 1 {
			t.Fatalf("first item = %+v, want the hoodie SKU", items[0])
		}
		if items[0].CategoryID != nil || items[0].CategoryName != nil {
			t.Fatalf("first item category = %+v, want uncategorized", items[0])
		}

		if items[1].SKUID != tshirtSKU.ID || items[1].ProductName != "T-shirt" ||
			items[1].TotalStock != 5 || items[1].FrozenStock != 1 || items[1].AvailableStock() != 4 {
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

func TestUpdateInventoryTotalStock(t *testing.T) {
	ctx := context.Background()
	repository := inventoryrepository.NewInventoryRepository(testPool)
	productRepository := productrepository.NewProductRepository(testPool)
	skuRepository := productrepository.NewProductSKURepository(testPool)

	resetDatabase(t)
	shop := insertShop(t)
	product := &model.Product{ShopID: shop.ID, Name: "T-shirt", Status: model.ProductStatusOnShelf}
	if err := productRepository.CreateProduct(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
	}
	sku := createTestSKU(t, ctx, skuRepository, product.ID, map[string]string{"color": "black"}, 1999)
	setTestInventory(t, ctx, sku.ID, 5, 2)

	item, err := repository.UpdateInventoryTotalStock(ctx, shop.ID, sku.ID, 8, 0)
	if err != nil {
		t.Fatalf("update inventory: %v", err)
	}
	if item.TotalStock != 8 || item.FrozenStock != 2 || item.Version != 1 {
		t.Fatalf("inventory item = %+v, want total 8 frozen 2 version 1", item)
	}

	_, err = repository.UpdateInventoryTotalStock(ctx, shop.ID, sku.ID, 9, 0)
	requireErrorIs(t, err, apperr.ErrInventoryVersion)

	_, err = repository.UpdateInventoryTotalStock(ctx, shop.ID, sku.ID, 1, 1)
	requireErrorIs(t, err, apperr.ErrInventoryBelowFrozen)
}
