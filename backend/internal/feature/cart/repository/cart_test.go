//go:build integration

package repository_test

import (
	"context"
	"math"
	"sync"
	"testing"

	"pinshaoshao/internal/feature/cart/apperr"
	productmodel "pinshaoshao/internal/feature/product/model"
)

func TestCartRepositoryAddAndList(t *testing.T) {
	ctx := context.Background()
	cartRepository := newTestCartRepository()

	t.Run("snapshots catalog data and accumulates quantity", func(t *testing.T) {
		resetDatabase(t)
		user := insertUser(t, "buyer01")
		product, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, true, 10)

		first, err := cartRepository.AddItem(ctx, user.ID, sku.ID, 2)
		if err != nil {
			t.Fatalf("add cart item: %v", err)
		}
		if first.ProductName != product.Name || first.Specs["color"] != "black" || first.UnitPriceCents != 1999 {
			t.Fatalf("cart item = %+v, want catalog snapshot", first)
		}

		if _, err := testPool.Exec(ctx, `UPDATE products SET name = 'Renamed' WHERE id = $1`, product.ID); err != nil {
			t.Fatalf("rename product: %v", err)
		}
		if _, err := testPool.Exec(ctx, `UPDATE product_skus SET price_cents = 2999 WHERE id = $1`, sku.ID); err != nil {
			t.Fatalf("change price: %v", err)
		}

		second, err := cartRepository.AddItem(ctx, user.ID, sku.ID, 3)
		if err != nil {
			t.Fatalf("add cart item again: %v", err)
		}
		if second.ID != first.ID || second.Quantity != 5 {
			t.Fatalf("cart item = %+v, want one item with quantity 5", second)
		}
		if second.ProductName != "T-shirt" || second.Specs["color"] != "black" || second.UnitPriceCents != 1999 {
			t.Fatalf("cart item = %+v, want original snapshot", second)
		}

		items, err := cartRepository.ListByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("list cart: %v", err)
		}
		if len(items) != 1 || items[0].ID != first.ID {
			t.Fatalf("items = %+v, want one cart item", items)
		}
	})

	t.Run("rejects quantity overflow without changing the cart", func(t *testing.T) {
		resetDatabase(t)
		user := insertUser(t, "buyer01")
		_, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, true, math.MaxInt32)

		if _, err := cartRepository.AddItem(ctx, user.ID, sku.ID, math.MaxInt32); err != nil {
			t.Fatalf("add cart item: %v", err)
		}
		_, err := cartRepository.AddItem(ctx, user.ID, sku.ID, 1)
		requireErrorIs(t, err, apperr.ErrInvalidCartQuantity)

		items, err := cartRepository.ListByUserID(ctx, user.ID)
		if err != nil {
			t.Fatalf("list cart: %v", err)
		}
		if len(items) != 1 || items[0].Quantity != math.MaxInt32 {
			t.Fatalf("items = %+v, want one item at the quantity limit", items)
		}
	})

	t.Run("rejects unavailable catalog items", func(t *testing.T) {
		tests := []struct {
			name  string
			setup func(*testing.T) int
		}{
			{
				name: "missing sku",
				setup: func(t *testing.T) int {
					return 404
				},
			},
			{
				name: "off shelf",
				setup: func(t *testing.T) int {
					_, sku := insertCatalogSKU(t, productmodel.ProductStatusOffShelf, true, 10)
					return sku.ID
				},
			},
			{
				name: "disabled sku",
				setup: func(t *testing.T) int {
					_, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, false, 10)
					return sku.ID
				},
			},
			{
				name: "no stock",
				setup: func(t *testing.T) int {
					_, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, true, 0)
					return sku.ID
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resetDatabase(t)
				user := insertUser(t, "buyer01")
				skuID := tt.setup(t)

				_, err := cartRepository.AddItem(ctx, user.ID, skuID, 1)
				if tt.name == "missing sku" {
					requireErrorIs(t, err, apperr.ErrProductNotFound)
					return
				}
				requireErrorIs(t, err, apperr.ErrProductUnavailable)
			})
		}
	})
}

func TestCartRepositoryAvailability(t *testing.T) {
	ctx := context.Background()
	cartRepository := newTestCartRepository()

	resetDatabase(t)
	user := insertUser(t, "buyer01")
	product, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, true, 5)

	if _, err := cartRepository.AddItem(ctx, user.ID, sku.ID, 3); err != nil {
		t.Fatalf("add cart item: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE products SET status = 'off_shelf' WHERE id = $1`, product.ID); err != nil {
		t.Fatalf("off shelf product: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE inventory_items SET total_stock = 2 WHERE sku_id = $1`, sku.ID); err != nil {
		t.Fatalf("reduce stock: %v", err)
	}

	items, err := cartRepository.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list cart: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %+v, want one item", items)
	}
	item := items[0]
	if got := item.Availability(); got != "product_off_shelf" {
		t.Fatalf("availability = %q, want off shelf", got)
	}
	if item.AvailableStock == nil || *item.AvailableStock != 2 {
		t.Fatalf("available stock = %v, want 2", item.AvailableStock)
	}
}

func TestCartRepositoryConcurrentAdd(t *testing.T) {
	ctx := context.Background()
	cartRepository := newTestCartRepository()

	resetDatabase(t)
	user := insertUser(t, "buyer01")
	_, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, true, 20)

	const additions = 8
	errs := make(chan error, additions)
	var wg sync.WaitGroup
	for range additions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cartRepository.AddItem(ctx, user.ID, sku.ID, 1)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent add: %v", err)
		}
	}

	items, err := cartRepository.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list cart: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("item count = %d, want 1", len(items))
	}
	if items[0].Quantity != additions {
		t.Fatalf("quantity = %d, want %d", items[0].Quantity, additions)
	}
}

func TestCartRepositoryUpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	cartRepository := newTestCartRepository()

	resetDatabase(t)
	user := insertUser(t, "buyer01")
	otherUser := insertUser(t, "buyer02")
	_, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, true, 10)
	item, err := cartRepository.AddItem(ctx, user.ID, sku.ID, 1)
	if err != nil {
		t.Fatalf("add cart item: %v", err)
	}

	_, err = cartRepository.UpdateQuantity(ctx, otherUser.ID, item.ID, 2)
	requireErrorIs(t, err, apperr.ErrCartItemNotFound)

	updated, err := cartRepository.UpdateQuantity(ctx, user.ID, item.ID, 4)
	if err != nil {
		t.Fatalf("update quantity: %v", err)
	}
	if updated.Quantity != 4 {
		t.Fatalf("quantity = %d, want 4", updated.Quantity)
	}

	if err := cartRepository.DeleteItem(ctx, otherUser.ID, item.ID); err == nil {
		t.Fatal("expected another user's delete to fail")
	}
	if err := cartRepository.DeleteItem(ctx, user.ID, item.ID); err != nil {
		t.Fatalf("delete cart item: %v", err)
	}
	err = cartRepository.DeleteItem(ctx, user.ID, item.ID)
	requireErrorIs(t, err, apperr.ErrCartItemNotFound)
}

func TestCartRepositoryUsesCurrentProductStatus(t *testing.T) {
	ctx := context.Background()
	cartRepository := newTestCartRepository()

	resetDatabase(t)
	user := insertUser(t, "buyer01")
	product, sku := insertCatalogSKU(t, productmodel.ProductStatusOnShelf, true, 10)
	if _, err := cartRepository.AddItem(ctx, user.ID, sku.ID, 1); err != nil {
		t.Fatalf("add cart item: %v", err)
	}

	if _, err := testPool.Exec(ctx, `UPDATE product_skus SET is_enabled = false WHERE id = $1`, sku.ID); err != nil {
		t.Fatalf("disable sku: %v", err)
	}
	items, err := cartRepository.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list cart: %v", err)
	}
	if got := items[0].Availability(); got != "sku_disabled" {
		t.Fatalf("availability = %q, want sku disabled", got)
	}

	if _, err := testPool.Exec(ctx, `DELETE FROM inventory_items WHERE sku_id = $1`, sku.ID); err != nil {
		t.Fatalf("delete inventory item: %v", err)
	}
	if _, err := testPool.Exec(ctx, `DELETE FROM products WHERE id = $1`, product.ID); err != nil {
		t.Fatalf("delete product: %v", err)
	}
	items, err = cartRepository.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list cart after sku deletion: %v", err)
	}
	if got := items[0].Availability(); got != "product_deleted" {
		t.Fatalf("availability = %q, want deleted product", got)
	}
}
