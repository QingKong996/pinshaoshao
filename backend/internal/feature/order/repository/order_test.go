//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"
	productmodel "pinshaoshao/internal/feature/product/model"
)

const testCheckoutKey = "4f2614df-f674-4b5e-a9e4-3fc5e7f62013"

func testRequestHash(suffix string) string {
	return strings.Repeat(suffix, 64)[:64]
}

func testOrderInput(itemIDs []int, quantities []int, hash string) model.CreateOrderInput {
	return testOrderInputWithKey(itemIDs, quantities, hash, testCheckoutKey)
}

func testOrderInputWithKey(itemIDs []int, quantities []int, hash string, checkoutKey string) model.CreateOrderInput {
	items := make([]model.CreateOrderItem, 0, len(itemIDs))
	for i := range itemIDs {
		items = append(items, model.CreateOrderItem{
			CartItemID: itemIDs[i],
			Quantity:   quantities[i],
		})
	}
	return model.CreateOrderInput{
		CheckoutKey:     checkoutKey,
		RequestHash:     hash,
		Items:           items,
		ReceiverName:    "Alice",
		ReceiverPhone:   "13800138000",
		ShippingAddress: "Shanghai",
	}
}

func testCheckoutKeyFor(index int) string {
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", index)
}

func TestOrderRepositoryCreateSplitsShopsAndFreezesStock(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	firstShop, _, firstSKU := insertCatalogSKU(t, "shop01", "First shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	secondShop, _, secondSKU := insertCatalogSKU(t, "shop02", "Second shop", productmodel.ProductStatusOnShelf, true, 2999, 10)
	firstItem := addCartItem(t, user.ID, firstSKU.ID, 2)
	secondItem := addCartItem(t, user.ID, secondSKU.ID, 1)

	order, err := newTestOrderRepository().Create(ctx, user.ID, testOrderInput(
		[]int{firstItem.ID, secondItem.ID},
		[]int{2, 1},
		testRequestHash("a"),
	))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.Status != model.OrderStatusPendingPayment {
		t.Fatalf("status = %q, want pending payment", order.Status)
	}
	if order.TotalAmountCents != 6997 {
		t.Fatalf("total = %d, want 6997", order.TotalAmountCents)
	}
	if len(order.SubOrders) != 2 {
		t.Fatalf("sub-order count = %d, want 2", len(order.SubOrders))
	}
	if order.SubOrders[0].ShopID != firstShop.ID || order.SubOrders[0].ShopName != "First shop" {
		t.Fatalf("first sub-order = %+v", order.SubOrders[0])
	}
	if order.SubOrders[1].ShopID != secondShop.ID || order.SubOrders[1].ShopName != "Second shop" {
		t.Fatalf("second sub-order = %+v", order.SubOrders[1])
	}

	requireFrozenStock(t, firstSKU.ID, 2)
	requireFrozenStock(t, secondSKU.ID, 1)
	requireCartItemCount(t, user.ID, 0)
}

func TestOrderRepositoryCreateIsIdempotent(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	input := testOrderInput([]int{item.ID}, []int{2}, testRequestHash("b"))

	first, err := newTestOrderRepository().Create(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	second, err := newTestOrderRepository().Create(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("replay order: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("order IDs = %d/%d, want the same order", first.ID, second.ID)
	}
	requireFrozenStock(t, sku.ID, 2)
	requireCartItemCount(t, user.ID, 0)

	input.RequestHash = testRequestHash("c")
	_, err = newTestOrderRepository().Create(ctx, user.ID, input)
	requireErrorIs(t, err, apperr.ErrCheckoutConflict)
}

func TestOrderRepositoryCreateIsConcurrentlyIdempotent(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	input := testOrderInput([]int{item.ID}, []int{2}, testRequestHash("b"))
	repository := newTestOrderRepository()

	const attempts = 8
	start := make(chan struct{})
	ids := make(chan int, attempts)
	errs := make(chan error, attempts)

	var wg sync.WaitGroup
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			order, err := repository.Create(ctx, user.ID, input)
			if err != nil {
				errs <- err
				return
			}
			ids <- order.ID
		}()
	}
	close(start)
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		t.Fatalf("concurrent create: %v", err)
	}

	var orderID int
	for id := range ids {
		if orderID == 0 {
			orderID = id
			continue
		}
		if id != orderID {
			t.Fatalf("order IDs = %d/%d, want the same order", orderID, id)
		}
	}
	requireFrozenStock(t, sku.ID, 2)
	requireCartItemCount(t, user.ID, 0)
	requireOrderCount(t, user.ID, 1)
}

func TestOrderRepositoryCreateReplayAfterCancelKeepsCancelledState(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	input := testOrderInput([]int{item.ID}, []int{2}, testRequestHash("b"))
	repository := newTestOrderRepository()

	first, err := repository.Create(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if _, err := repository.Cancel(ctx, user.ID, first.ID); err != nil {
		t.Fatalf("cancel order: %v", err)
	}

	replayed, err := repository.Create(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("replay cancelled order: %v", err)
	}
	if replayed.ID != first.ID || replayed.Status != model.OrderStatusCancelled {
		t.Fatalf("replayed order = %d/%q, want %d/cancelled", replayed.ID, replayed.Status, first.ID)
	}
	requireFrozenStock(t, sku.ID, 0)
}

func TestOrderRepositoryCreateRollsBackOnInsufficientStock(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 2)
	item := addCartItem(t, user.ID, sku.ID, 3)

	_, err := newTestOrderRepository().Create(ctx, user.ID, testOrderInput(
		[]int{item.ID},
		[]int{3},
		testRequestHash("d"),
	))
	requireErrorIs(t, err, apperr.ErrInsufficientStock)
	requireFrozenStock(t, sku.ID, 0)
	requireCartItemCount(t, user.ID, 1)
	requireOrderCount(t, user.ID, 0)
}

func TestOrderRepositoryCreateUsesCartSnapshot(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	if _, err := testPool.Exec(
		ctx,
		`UPDATE product_skus SET price_cents = 2999, specs = '{"color":"red"}' WHERE id = $1`,
		sku.ID,
	); err != nil {
		t.Fatalf("change sku: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE products SET name = 'Updated T-shirt' WHERE id = $1`, sku.ProductID); err != nil {
		t.Fatalf("change product: %v", err)
	}

	order, err := newTestOrderRepository().Create(ctx, user.ID, testOrderInput(
		[]int{item.ID},
		[]int{2},
		testRequestHash("e"),
	))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order.TotalAmountCents != 3998 {
		t.Fatalf("total = %d, want 3998", order.TotalAmountCents)
	}
	itemSnapshot := order.SubOrders[0].Items[0]
	if itemSnapshot.UnitPriceCents != 1999 {
		t.Fatalf("unit price = %d, want 1999", itemSnapshot.UnitPriceCents)
	}
	if itemSnapshot.ProductName != "T-shirt" || itemSnapshot.Specs["color"] != "black" {
		t.Fatalf("item snapshot = %+v, want cart snapshot", itemSnapshot)
	}
}

func TestOrderRepositoryCancelReleasesStockOnce(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	order, err := newTestOrderRepository().Create(ctx, user.ID, testOrderInput(
		[]int{item.ID},
		[]int{2},
		testRequestHash("f"),
	))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	cancelled, err := newTestOrderRepository().Cancel(ctx, user.ID, order.ID)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if cancelled.Status != model.OrderStatusCancelled {
		t.Fatalf("status = %q, want cancelled", cancelled.Status)
	}
	if cancelled.SubOrders[0].Status != model.SubOrderStatusCancelled {
		t.Fatalf("sub-order status = %q, want cancelled", cancelled.SubOrders[0].Status)
	}
	requireFrozenStock(t, sku.ID, 0)

	again, err := newTestOrderRepository().Cancel(ctx, user.ID, order.ID)
	if err != nil {
		t.Fatalf("cancel order again: %v", err)
	}
	if again.Status != model.OrderStatusCancelled {
		t.Fatalf("status = %q, want cancelled", again.Status)
	}
	requireFrozenStock(t, sku.ID, 0)
}

func TestOrderRepositoryListCancelsDueOrder(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	repository := newTestOrderRepository()

	order, err := repository.Create(ctx, user.ID, testOrderInput(
		[]int{item.ID},
		[]int{2},
		testRequestHash("f"),
	))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	markOrderCreatedMinutesAgo(t, order.ID, 31)

	orders, err := repository.List(ctx, user.ID, model.OrderListFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(orders.Orders) != 1 || orders.Orders[0].Status != model.OrderStatusCancelled {
		t.Fatalf("orders = %+v, want one cancelled order", orders.Orders)
	}
	requireFrozenStock(t, sku.ID, 0)
}

func TestOrderRepositoryGetCancelsDueOrder(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	repository := newTestOrderRepository()

	order, err := repository.Create(ctx, user.ID, testOrderInput(
		[]int{item.ID},
		[]int{2},
		testRequestHash("f"),
	))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	markOrderCreatedMinutesAgo(t, order.ID, 31)

	expired, err := repository.Get(ctx, user.ID, order.ID)
	if err != nil {
		t.Fatalf("get expired order: %v", err)
	}
	if expired.Status != model.OrderStatusCancelled || expired.SubOrders[0].Status != model.SubOrderStatusCancelled {
		t.Fatalf("expired order = %q/%q, want cancelled", expired.Status, expired.SubOrders[0].Status)
	}
	requireFrozenStock(t, sku.ID, 0)

	expired, err = repository.Get(ctx, user.ID, order.ID)
	if err != nil {
		t.Fatalf("get expired order again: %v", err)
	}
	if expired.Status != model.OrderStatusCancelled {
		t.Fatalf("status = %q, want cancelled", expired.Status)
	}
	requireFrozenStock(t, sku.ID, 0)
}

func TestOrderRepositoryListContinuesWhenDueOrderCancellationFails(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, firstSKU := insertCatalogSKU(t, "shop01", "First shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	_, _, secondSKU := insertCatalogSKU(t, "shop02", "Second shop", productmodel.ProductStatusOnShelf, true, 2999, 10)
	firstItem := addCartItem(t, user.ID, firstSKU.ID, 1)
	secondItem := addCartItem(t, user.ID, secondSKU.ID, 1)
	repository := newTestOrderRepository()

	first, err := repository.Create(ctx, user.ID, testOrderInputWithKey(
		[]int{firstItem.ID},
		[]int{1},
		testRequestHash("a"),
		testCheckoutKeyFor(1),
	))
	if err != nil {
		t.Fatalf("create first order: %v", err)
	}
	second, err := repository.Create(ctx, user.ID, testOrderInputWithKey(
		[]int{secondItem.ID},
		[]int{1},
		testRequestHash("b"),
		testCheckoutKeyFor(2),
	))
	if err != nil {
		t.Fatalf("create second order: %v", err)
	}
	markOrderCreatedMinutesAgo(t, first.ID, 32)
	markOrderCreatedMinutesAgo(t, second.ID, 31)
	if _, err := testPool.Exec(ctx, `UPDATE inventory_items SET frozen_stock = 0 WHERE sku_id = $1`, firstSKU.ID); err != nil {
		t.Fatalf("corrupt frozen stock: %v", err)
	}

	orders, err := repository.List(ctx, user.ID, model.OrderListFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if len(orders.Orders) != 2 {
		t.Fatalf("order count = %d, want 2", len(orders.Orders))
	}

	statuses := make(map[int]model.OrderStatus, len(orders.Orders))
	for _, order := range orders.Orders {
		statuses[order.ID] = order.Status
	}
	if statuses[first.ID] != model.OrderStatusPendingPayment {
		t.Fatalf("first order status = %q, want pending_payment", statuses[first.ID])
	}
	if statuses[second.ID] != model.OrderStatusCancelled {
		t.Fatalf("second order status = %q, want cancelled", statuses[second.ID])
	}
	requireFrozenStock(t, secondSKU.ID, 0)
}

func markOrderCreatedMinutesAgo(t *testing.T, orderID int, minutesAgo int) {
	t.Helper()

	if _, err := testPool.Exec(context.Background(), `
UPDATE orders
SET created_at = now() - make_interval(mins => $2)
WHERE id = $1`, orderID, minutesAgo); err != nil {
		t.Fatalf("expire order: %v", err)
	}
}

func TestOrderRepositoryListPaginates(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	repository := newTestOrderRepository()

	for index := 1; index <= 3; index++ {
		_, _, sku := insertCatalogSKU(
			t,
			fmt.Sprintf("shop%02d", index),
			fmt.Sprintf("Shop %d", index),
			productmodel.ProductStatusOnShelf,
			true,
			1999,
			10,
		)
		item := addCartItem(t, user.ID, sku.ID, 1)
		if _, err := repository.Create(ctx, user.ID, testOrderInputWithKey(
			[]int{item.ID},
			[]int{1},
			testRequestHash("a"),
			testCheckoutKeyFor(index),
		)); err != nil {
			t.Fatalf("create order %d: %v", index, err)
		}
	}

	firstPage, err := repository.List(ctx, user.ID, model.OrderListFilter{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if firstPage.Total != 3 || len(firstPage.Orders) != 2 {
		t.Fatalf("first page = %+v, want total 3 with 2 items", firstPage)
	}

	secondPage, err := repository.List(ctx, user.ID, model.OrderListFilter{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if secondPage.Total != 3 || len(secondPage.Orders) != 1 {
		t.Fatalf("second page = %+v, want total 3 with 1 item", secondPage)
	}
}

func TestOrderRepositoryRejectsOtherUser(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	owner := insertBuyer(t, "buyer01")
	other := insertBuyer(t, "buyer02")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, owner.ID, sku.ID, 1)
	repository := newTestOrderRepository()

	order, err := repository.Create(ctx, owner.ID, testOrderInput(
		[]int{item.ID},
		[]int{1},
		testRequestHash("a"),
	))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	_, err = repository.Get(ctx, other.ID, order.ID)
	requireErrorIs(t, err, apperr.ErrOrderNotFound)
	_, err = repository.Cancel(ctx, other.ID, order.ID)
	requireErrorIs(t, err, apperr.ErrOrderNotFound)
}

func TestOrderRepositoryCancelReportsInventoryInconsistency(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	_, _, sku := insertCatalogSKU(t, "shop01", "Shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	item := addCartItem(t, user.ID, sku.ID, 2)
	repository := newTestOrderRepository()

	order, err := repository.Create(ctx, user.ID, testOrderInput(
		[]int{item.ID},
		[]int{2},
		testRequestHash("a"),
	))
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE inventory_items SET frozen_stock = 0 WHERE sku_id = $1`, sku.ID); err != nil {
		t.Fatalf("corrupt frozen stock: %v", err)
	}

	_, err = repository.Cancel(ctx, user.ID, order.ID)
	requireErrorIs(t, err, apperr.ErrOrderStateInvalid)

	pending, err := repository.Get(ctx, user.ID, order.ID)
	if err != nil {
		t.Fatalf("get order after failed cancel: %v", err)
	}
	if pending.Status != model.OrderStatusPendingPayment {
		t.Fatalf("status = %q, want pending_payment", pending.Status)
	}
}

func requireFrozenStock(t *testing.T, skuID int, want int) {
	t.Helper()

	var got int
	if err := testPool.QueryRow(context.Background(), `SELECT frozen_stock FROM inventory_items WHERE sku_id = $1`, skuID).Scan(&got); err != nil {
		t.Fatalf("read frozen stock: %v", err)
	}
	if got != want {
		t.Fatalf("frozen stock = %d, want %d", got, want)
	}
}

func requireCartItemCount(t *testing.T, userID int, want int) {
	t.Helper()

	var got int
	if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*)::int FROM cart_items WHERE user_id = $1`, userID).Scan(&got); err != nil {
		t.Fatalf("count cart items: %v", err)
	}
	if got != want {
		t.Fatalf("cart item count = %d, want %d", got, want)
	}
}

func requireOrderCount(t *testing.T, userID int, want int) {
	t.Helper()

	var got int
	if err := testPool.QueryRow(context.Background(), `SELECT COUNT(*)::int FROM orders WHERE user_id = $1`, userID).Scan(&got); err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if got != want {
		t.Fatalf("order count = %d, want %d", got, want)
	}
}
