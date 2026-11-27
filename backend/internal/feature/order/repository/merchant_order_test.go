//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"
	productmodel "pinshaoshao/internal/feature/product/model"
)

func TestOrderRepositoryListMerchantFiltersByShopAndStatus(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	firstShop, _, firstSKU := insertCatalogSKU(t, "shop01", "First shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	secondShop, _, secondSKU := insertCatalogSKU(t, "shop02", "Second shop", productmodel.ProductStatusOnShelf, true, 2999, 10)
	repository := newTestOrderRepository()

	firstInput := testDirectOrderInput(firstSKU.ID, 1, testRequestHash("a"))
	firstInput.CheckoutKey = testCheckoutKeyFor(1)
	first, err := repository.CreateDirect(ctx, user.ID, firstInput)
	if err != nil {
		t.Fatalf("create first order: %v", err)
	}
	secondInput := testDirectOrderInput(firstSKU.ID, 2, testRequestHash("b"))
	secondInput.CheckoutKey = testCheckoutKeyFor(2)
	second, err := repository.CreateDirect(ctx, user.ID, secondInput)
	if err != nil {
		t.Fatalf("create second order: %v", err)
	}
	otherInput := testDirectOrderInput(secondSKU.ID, 1, testRequestHash("c"))
	otherInput.CheckoutKey = testCheckoutKeyFor(3)
	other, err := repository.CreateDirect(ctx, user.ID, otherInput)
	if err != nil {
		t.Fatalf("create other shop order: %v", err)
	}

	firstSubOrderID := first.SubOrders[0].ID
	if _, err := testPool.Exec(ctx, `UPDATE sub_orders SET status = 'awaiting_shipment' WHERE id = $1`, firstSubOrderID); err != nil {
		t.Fatalf("mark sub-order awaiting shipment: %v", err)
	}
	if _, err := testPool.Exec(
		ctx,
		`UPDATE orders SET status = 'paid', paid_at = now() WHERE id = $1`,
		first.ID,
	); err != nil {
		t.Fatalf("mark order paid: %v", err)
	}

	awaiting, err := repository.ListMerchant(ctx, firstShop.ID, model.MerchantOrderListFilter{
		Status:   model.SubOrderStatusAwaitingShipment,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("list awaiting merchant orders: %v", err)
	}
	if awaiting.Total != 1 || len(awaiting.Orders) != 1 {
		t.Fatalf("awaiting orders = %+v, want one order", awaiting)
	}
	got := awaiting.Orders[0]
	if got.ID != firstSubOrderID || got.OrderID != first.ID || got.TotalAmountCents != 1999 || got.ItemCount != 1 {
		t.Fatalf("awaiting order = %+v", got)
	}
	if got.PaidAt == nil {
		t.Fatal("paid_at = nil, want a timestamp")
	}

	receiver, err := repository.GetMerchantReceiver(ctx, firstShop.ID, firstSubOrderID)
	if err != nil {
		t.Fatalf("get merchant receiver: %v", err)
	}
	if receiver.ReceiverName != "Alice" ||
		receiver.ReceiverPhone != "13800138000" ||
		receiver.ShippingAddress != "Shanghai" {
		t.Fatalf("receiver = %+v", receiver)
	}

	if _, err := repository.GetMerchantReceiver(
		ctx,
		secondShop.ID,
		firstSubOrderID,
	); !errors.Is(err, apperr.ErrOrderNotFound) {
		t.Fatalf("other shop receiver error = %v, want %v", err, apperr.ErrOrderNotFound)
	}
	if _, err := repository.GetMerchantReceiver(
		ctx,
		firstShop.ID,
		second.SubOrders[0].ID,
	); !errors.Is(err, apperr.ErrReceiverUnavailable) {
		t.Fatalf("pending receiver error = %v, want %v", err, apperr.ErrReceiverUnavailable)
	}

	firstPage, err := repository.ListMerchant(ctx, firstShop.ID, model.MerchantOrderListFilter{Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("list first merchant page: %v", err)
	}
	if firstPage.Total != 2 || len(firstPage.Orders) != 1 {
		t.Fatalf("first merchant page = %+v, want total 2 with 1 item", firstPage)
	}
	if firstPage.Orders[0].OrderID != second.ID {
		t.Fatalf("first merchant order = %d, want newest order %d", firstPage.Orders[0].OrderID, second.ID)
	}
	secondPage, err := repository.ListMerchant(ctx, firstShop.ID, model.MerchantOrderListFilter{Page: 2, PageSize: 1})
	if err != nil {
		t.Fatalf("list second merchant page: %v", err)
	}
	if secondPage.Total != 2 || len(secondPage.Orders) != 1 || secondPage.Orders[0].OrderID != first.ID {
		t.Fatalf("second merchant page = %+v, want order %d", secondPage, first.ID)
	}

	otherShopOrders, err := repository.ListMerchant(ctx, secondShop.ID, model.MerchantOrderListFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list other shop orders: %v", err)
	}
	if otherShopOrders.Total != 1 || len(otherShopOrders.Orders) != 1 || otherShopOrders.Orders[0].OrderID != other.ID {
		t.Fatalf("other shop orders = %+v, want order %d", otherShopOrders, other.ID)
	}
}

func TestOrderRepositoryGetMerchantIsScopedToShop(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	user := insertBuyer(t, "buyer01")
	firstShop, _, firstSKU := insertCatalogSKU(t, "shop01", "First shop", productmodel.ProductStatusOnShelf, true, 1999, 10)
	secondShop, _, _ := insertCatalogSKU(t, "shop02", "Second shop", productmodel.ProductStatusOnShelf, true, 2999, 10)
	repository := newTestOrderRepository()

	input := testDirectOrderInput(firstSKU.ID, 2, testRequestHash("a"))
	created, err := repository.CreateDirect(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	subOrderID := created.SubOrders[0].ID

	detail, err := repository.GetMerchant(ctx, firstShop.ID, subOrderID)
	if err != nil {
		t.Fatalf("get merchant order: %v", err)
	}
	if detail.ID != subOrderID ||
		detail.OrderID != created.ID ||
		detail.TotalAmountCents != 3998 ||
		detail.ItemCount != 2 {
		t.Fatalf("merchant order = %+v", detail)
	}
	if len(detail.Items) != 1 ||
		detail.Items[0].ProductID != firstSKU.ProductID ||
		detail.Items[0].SKUID != firstSKU.ID ||
		detail.Items[0].Quantity != 2 ||
		detail.Items[0].LineAmountCents != 3998 {
		t.Fatalf("merchant order items = %+v", detail.Items)
	}

	if _, err := repository.GetMerchant(ctx, secondShop.ID, subOrderID); !errors.Is(err, apperr.ErrOrderNotFound) {
		t.Fatalf("error = %v, want %v", err, apperr.ErrOrderNotFound)
	}
}
