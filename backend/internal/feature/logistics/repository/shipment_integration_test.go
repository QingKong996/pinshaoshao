//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	inventoryrepository "pinshaoshao/internal/feature/inventory/repository"
	"pinshaoshao/internal/feature/logistics/apperr"
	"pinshaoshao/internal/feature/logistics/model"
	logisticsrepository "pinshaoshao/internal/feature/logistics/repository"
	orderrepository "pinshaoshao/internal/feature/order/repository"

	"github.com/google/uuid"
)

func TestShipmentRepositoryShippingAndDeliveryLifecycle(t *testing.T) {
	ctx := context.Background()
	resetDatabase(t)
	seed := insertShippableSubOrder(t)

	orders := orderrepository.NewShippingLifecycle(
		testPool,
		inventoryrepository.NewInventoryRepository(testPool),
	)
	repository := logisticsrepository.NewShipmentRepository(testPool, orders)
	target, err := orders.GetShipmentTarget(ctx, seed.shopID, seed.subOrderID)
	if err != nil {
		t.Fatalf("get shipment target: %v", err)
	}

	shippedAt := time.Now().UTC().Truncate(time.Microsecond)
	created, isNew, err := repository.Create(ctx, &model.Shipment{
		ShipmentNo:         uuid.NewString(),
		SubOrderID:         target.SubOrderID,
		OrderID:            target.OrderID,
		UserID:             target.UserID,
		ShopID:             target.ShopID,
		Provider:           model.ProviderMockLogistics,
		ProviderShipmentNo: uuid.NewString(),
		ProviderTrackingNo: "SF123456789",
		Carrier:            "顺丰速运",
		Status:             model.StatusPendingPickup,
		ItemSummary:        target.ItemSummary,
		ReceiverName:       target.ReceiverName,
		ReceiverPhone:      target.ReceiverPhone,
		ReceiverAddress:    target.ShippingAddress,
		OriginCity:         "Shanghai",
		DestinationCity:    "Hangzhou",
		CurrentLocation:    "Shanghai merchant warehouse",
		ShippedAt:          shippedAt,
	}, target)
	if err != nil {
		t.Fatalf("create shipment: %v", err)
	}
	if !isNew || created.Carrier != "顺丰速运" {
		t.Fatalf("created shipment = %+v, isNew = %v", created, isNew)
	}
	requireInventory(t, seed.skuID, 3, 0)
	requireSubOrderStatus(t, seed.subOrderID, "shipped")
	requireOrderStatus(t, seed.orderID, "shipped")

	replayed, replayedIsNew, err := repository.Create(ctx, &model.Shipment{
		ShipmentNo:         uuid.NewString(),
		SubOrderID:         target.SubOrderID,
		OrderID:            target.OrderID,
		UserID:             target.UserID,
		ShopID:             target.ShopID,
		Provider:           model.ProviderMockLogistics,
		ProviderShipmentNo: uuid.NewString(),
		ProviderTrackingNo: "OTHER123",
		Carrier:            "Other carrier",
		Status:             model.StatusPendingPickup,
		ItemSummary:        target.ItemSummary,
		ReceiverName:       target.ReceiverName,
		ReceiverPhone:      target.ReceiverPhone,
		ReceiverAddress:    target.ShippingAddress,
		ShippedAt:          shippedAt,
	}, target)
	if err != nil {
		t.Fatalf("replay shipment: %v", err)
	}
	if replayedIsNew || replayed.ShipmentNo != created.ShipmentNo {
		t.Fatalf("replayed shipment = %+v, isNew = %v", replayed, replayedIsNew)
	}
	requireInventory(t, seed.skuID, 3, 0)

	deliveredAt := shippedAt.Add(24 * time.Hour)
	callback := model.Callback{
		EventID:            uuid.NewString(),
		EventType:          model.EventTypeShipmentDelivered,
		ShipmentNo:         created.ShipmentNo,
		ProviderShipmentNo: created.ProviderShipmentNo,
		ProviderTrackingNo: created.ProviderTrackingNo,
		Status:             model.StatusDelivered,
		Sequence:           7,
		Description:        "Delivered",
		Location:           "Hangzhou",
		OccurredAt:         deliveredAt,
	}
	delivered, err := repository.ApplyEvent(ctx, callback)
	if err != nil {
		t.Fatalf("apply delivered callback: %v", err)
	}
	if delivered.Status != model.StatusDelivered ||
		delivered.DeliveredAt == nil ||
		len(delivered.Events) != 1 {
		t.Fatalf("delivered shipment = %+v", delivered)
	}
	requireSubOrderStatus(t, seed.subOrderID, "completed")
	requireOrderStatus(t, seed.orderID, "completed")

	replayedDelivered, err := repository.ApplyEvent(ctx, callback)
	if err != nil {
		t.Fatalf("replay delivered callback: %v", err)
	}
	if len(replayedDelivered.Events) != 1 {
		t.Fatalf("events after replay = %d, want 1", len(replayedDelivered.Events))
	}

	outOfOrder, err := repository.ApplyEvent(ctx, model.Callback{
		EventID:            uuid.NewString(),
		EventType:          model.EventTypeShipmentInTransit,
		ShipmentNo:         created.ShipmentNo,
		ProviderShipmentNo: created.ProviderShipmentNo,
		ProviderTrackingNo: created.ProviderTrackingNo,
		Status:             model.StatusInTransit,
		Sequence:           4,
		Description:        "In transit",
		Location:           "Shanghai",
		OccurredAt:         deliveredAt.Add(-time.Hour),
	})
	if err != nil {
		t.Fatalf("apply out-of-order callback: %v", err)
	}
	if outOfOrder.Status != model.StatusDelivered ||
		len(outOfOrder.Events) != 2 {
		t.Fatalf("out-of-order shipment = %+v", outOfOrder)
	}
	requireSubOrderStatus(t, seed.subOrderID, "completed")
	requireOrderStatus(t, seed.orderID, "completed")

	_, err = repository.ApplyEvent(ctx, model.Callback{
		EventID:            uuid.NewString(),
		EventType:          model.EventTypeShipmentDelivered,
		ShipmentNo:         created.ShipmentNo,
		ProviderShipmentNo: created.ProviderShipmentNo,
		ProviderTrackingNo: created.ProviderTrackingNo,
		Status:             model.StatusDelivered,
		Sequence:           7,
		Description:        "Conflicting event",
		Location:           "Hangzhou",
		OccurredAt:         deliveredAt,
	})
	if !errors.Is(err, apperr.ErrShipmentConflict) {
		t.Fatalf("event sequence conflict = %v, want conflict", err)
	}

	buyerShipments, err := repository.ListForBuyerOrder(
		ctx,
		seed.userID,
		seed.orderID,
	)
	if err != nil {
		t.Fatalf("list buyer shipments: %v", err)
	}
	if len(buyerShipments) != 1 ||
		buyerShipments[0].Status != model.StatusDelivered {
		t.Fatalf("buyer shipments = %+v", buyerShipments)
	}
}

type shippableSubOrder struct {
	userID     int
	shopID     int
	orderID    int
	subOrderID int
	skuID      int
}

func insertShippableSubOrder(t *testing.T) shippableSubOrder {
	t.Helper()
	ctx := context.Background()

	var seed shippableSubOrder
	if err := testPool.QueryRow(ctx, `INSERT INTO users (username, password_hash, email)
VALUES ('logistics-buyer', 'hash', 'logistics-buyer@example.com')
RETURNING id`).Scan(&seed.userID); err != nil {
		t.Fatalf("insert buyer: %v", err)
	}
	if err := testPool.QueryRow(ctx, `INSERT INTO shops (username, password_hash, email, name)
VALUES ('logistics-shop', 'hash', 'logistics-shop@example.com', 'Logistics Shop')
RETURNING id`).Scan(&seed.shopID); err != nil {
		t.Fatalf("insert shop: %v", err)
	}

	var productID int
	if err := testPool.QueryRow(ctx, `INSERT INTO products (shop_id, name, status)
VALUES ($1, 'Keyboard', 'on_shelf')
RETURNING id`, seed.shopID).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	if err := testPool.QueryRow(ctx, `INSERT INTO product_skus
    (product_id, specs, price_cents, is_enabled)
VALUES ($1, '{"Color":"Black"}'::jsonb, 1999, true)
RETURNING id`, productID).Scan(&seed.skuID); err != nil {
		t.Fatalf("insert SKU: %v", err)
	}
	if _, err := testPool.Exec(ctx, `INSERT INTO inventory_items
    (sku_id, total_stock, frozen_stock)
VALUES ($1, 5, 2)`, seed.skuID); err != nil {
		t.Fatalf("insert inventory: %v", err)
	}

	if err := testPool.QueryRow(ctx, `INSERT INTO orders
    (user_id, checkout_key, request_hash, status, total_amount_cents,
     receiver_name, receiver_phone, shipping_address, paid_at)
VALUES ($1, $2::uuid, $3, 'paid', 3998, 'Receiver', '1234567890', 'Hangzhou', now())
RETURNING id`,
		seed.userID,
		uuid.NewString(),
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	).Scan(&seed.orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	if err := testPool.QueryRow(ctx, `INSERT INTO sub_orders
    (order_id, shop_id, shop_name, status, total_amount_cents)
VALUES ($1, $2, 'Logistics Shop', 'awaiting_shipment', 3998)
RETURNING id`, seed.orderID, seed.shopID).Scan(&seed.subOrderID); err != nil {
		t.Fatalf("insert sub-order: %v", err)
	}
	if _, err := testPool.Exec(ctx, `INSERT INTO order_item_snapshots
    (sub_order_id, product_id, sku_id, product_name, sku_specs,
     unit_price_cents, quantity)
VALUES ($1, $2, $3, 'Keyboard', '{"Color":"Black"}'::jsonb, 1999, 2)`,
		seed.subOrderID, productID, seed.skuID); err != nil {
		t.Fatalf("insert item snapshot: %v", err)
	}
	return seed
}

func requireInventory(t *testing.T, skuID int, total int, frozen int) {
	t.Helper()

	var gotTotal, gotFrozen int
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT total_stock, frozen_stock FROM inventory_items WHERE sku_id = $1`,
		skuID,
	).Scan(&gotTotal, &gotFrozen); err != nil {
		t.Fatalf("read inventory: %v", err)
	}
	if gotTotal != total || gotFrozen != frozen {
		t.Fatalf(
			"inventory = %d/%d, want %d/%d",
			gotTotal,
			gotFrozen,
			total,
			frozen,
		)
	}
}

func requireSubOrderStatus(t *testing.T, subOrderID int, want string) {
	t.Helper()

	var got string
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT status::text FROM sub_orders WHERE id = $1`,
		subOrderID,
	).Scan(&got); err != nil {
		t.Fatalf("read sub-order status: %v", err)
	}
	if got != want {
		t.Fatalf("sub-order status = %q, want %q", got, want)
	}
}

func requireOrderStatus(t *testing.T, orderID int, want string) {
	t.Helper()

	var got string
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT status::text FROM orders WHERE id = $1`,
		orderID,
	).Scan(&got); err != nil {
		t.Fatalf("read order status: %v", err)
	}
	if got != want {
		t.Fatalf("order status = %q, want %q", got, want)
	}
}
