//go:build integration

package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/feature/product/model"
	productrepository "pinshaoshao/internal/feature/product/repository"
	shopmodel "pinshaoshao/internal/feature/shop/model"
	shoprepository "pinshaoshao/internal/feature/shop/repository"
)

func TestDashboardRepositoryOverview(t *testing.T) {
	ctx := context.Background()
	dashboardRepository := shoprepository.NewDashboardRepository(testPool)
	productRepository := productrepository.NewProductRepository(testPool)
	categoryRepository := productrepository.NewProductCategoryRepository(testPool)
	skuRepository := productrepository.NewProductSKURepository(testPool)

	t.Run("aggregates the counters of the shop", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)

		category := &model.ProductCategory{ShopID: shop.ID, Name: "clothes"}
		if err := categoryRepository.CreateProductCategory(ctx, category); err != nil {
			t.Fatalf("create category: %v", err)
		}

		onShelf := &model.Product{
			ShopID:     shop.ID,
			Name:       "T-shirt",
			Status:     model.ProductStatusOnShelf,
			CategoryID: &category.ID,
		}
		if err := productRepository.CreateProduct(ctx, onShelf); err != nil {
			t.Fatalf("create on-shelf product: %v", err)
		}
		offShelf := &model.Product{
			ShopID: shop.ID,
			Name:   "Hoodie",
			Status: model.ProductStatusOffShelf,
		}
		if err := productRepository.CreateProduct(ctx, offShelf); err != nil {
			t.Fatalf("create off-shelf product: %v", err)
		}

		black := createTestSKU(t, ctx, skuRepository, onShelf.ID, map[string]string{"variant": "black"}, 1000)
		setTestInventory(t, ctx, black.ID, 3, 0)
		createTestSKU(t, ctx, skuRepository, onShelf.ID, map[string]string{"variant": "white"}, 2000)
		if err := skuRepository.CreateProductSKU(ctx, &model.ProductSKU{
			ProductID:  onShelf.ID,
			Specs:      map[string]string{"variant": "disabled"},
			PriceCents: 3000,
			IsEnabled:  false,
		}); err != nil {
			t.Fatalf("create disabled SKU: %v", err)
		}
		offShelfSKU := createTestSKU(t, ctx, skuRepository, offShelf.ID, map[string]string{"size": "M"}, 500)
		setTestInventory(t, ctx, offShelfSKU.ID, 2, 0)

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
		setTestInventory(t, ctx, otherSKU.ID, 100, 0)

		stats, err := dashboardRepository.GetDashboardStats(ctx, shop.ID)
		if err != nil {
			t.Fatalf("get dashboard stats: %v", err)
		}
		if stats.ProductCount != 2 || stats.OnShelfProductCount != 1 {
			t.Fatalf("product counters = %+v, want 2 total and 1 on shelf", stats)
		}
		if stats.SKUCount != 4 || stats.CategoryCount != 1 {
			t.Fatalf("catalog counters = %+v, want 4 SKUs and 1 category", stats)
		}
		if stats.TotalStock != 5 || stats.InventoryValueCents != 4000 {
			t.Fatalf("inventory counters = %+v, want stock 5 and value 4000", stats)
		}
		if stats.OutOfStockSKUCount != 1 {
			t.Fatalf("out-of-stock SKU count = %d, want 1", stats.OutOfStockSKUCount)
		}
	})

	t.Run("returns zero counters for an empty shop", func(t *testing.T) {
		resetDatabase(t)
		shop := insertShop(t)

		stats, err := dashboardRepository.GetDashboardStats(ctx, shop.ID)
		if err != nil {
			t.Fatalf("get dashboard stats: %v", err)
		}
		if *stats != (shopmodel.DashboardStats{}) {
			t.Fatalf("dashboard stats = %+v, want zero counters", stats)
		}
	})
}
