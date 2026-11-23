package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/inventory"
	inventoryapperr "pinshaoshao/internal/feature/inventory/apperr"
	inventorymodel "pinshaoshao/internal/feature/inventory/model"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const dueOrderCleanupLimit = 100

type OrderRepository struct {
	pool        *pgxpool.Pool
	reservation inventory.Reservation
}

// NewOrderRepository builds the order persistence around the stock reservation port.
func NewOrderRepository(pool *pgxpool.Pool, reservation inventory.Reservation) *OrderRepository {
	return &OrderRepository{pool: pool, reservation: reservation}
}

func (r *OrderRepository) Create(ctx context.Context, userID int, input model.CreateOrderInput) (*model.Order, error) {
	order, err := r.create(ctx, userID, checkoutDetails{
		CheckoutKey:     input.CheckoutKey,
		RequestHash:     input.RequestHash,
		ReceiverName:    input.ReceiverName,
		ReceiverPhone:   input.ReceiverPhone,
		ShippingAddress: input.ShippingAddress,
	}, func(ctx context.Context, tx pgx.Tx, orderID int) error {
		return r.fillCartOrder(ctx, tx, userID, orderID, input.Items)
	})
	if err != nil {
		return nil, fmt.Errorf("repository.CreateOrder: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) CreateDirect(ctx context.Context, userID int, input model.CreateDirectOrderInput) (*model.Order, error) {
	order, err := r.create(ctx, userID, checkoutDetails{
		CheckoutKey:     input.CheckoutKey,
		RequestHash:     input.RequestHash,
		ReceiverName:    input.ReceiverName,
		ReceiverPhone:   input.ReceiverPhone,
		ShippingAddress: input.ShippingAddress,
	}, func(ctx context.Context, tx pgx.Tx, orderID int) error {
		return r.fillDirectOrder(ctx, tx, orderID, input.SKUID, input.Quantity)
	})
	if err != nil {
		return nil, fmt.Errorf("repository.CreateDirectOrder: %w", err)
	}
	return order, nil
}

type checkoutDetails struct {
	CheckoutKey     string
	RequestHash     string
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
}

func (r *OrderRepository) create(
	ctx context.Context,
	userID int,
	checkout checkoutDetails,
	fill func(context.Context, pgx.Tx, int) error,
) (*model.Order, error) {
	var result *model.Order
	var replayOrderID int

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		orderID, created, err := claimCheckoutKey(ctx, tx, userID, checkout)
		if err != nil {
			return err
		}
		if !created {
			replayOrderID = orderID
			return nil
		}

		if err := fill(ctx, tx, orderID); err != nil {
			return err
		}
		result, err = loadOrder(ctx, tx, userID, orderID)
		return err
	})
	if err != nil {
		return nil, err
	}
	if replayOrderID > 0 {
		replay, err := r.Get(ctx, userID, replayOrderID)
		if err != nil {
			return nil, err
		}
		return replay, nil
	}
	return result, nil
}

func (r *OrderRepository) List(ctx context.Context, userID int, filter model.OrderListFilter) (*model.OrderListResult, error) {
	r.cancelDueOrdersForUser(ctx, userID)

	var result *model.OrderListResult

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}, func(tx pgx.Tx) error {
		const countQuery = `SELECT COUNT(*) FROM orders WHERE user_id = $1`

		var total int64
		if err := tx.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
			return fmt.Errorf("count: %w", err)
		}

		const query = `SELECT o.id, o.status::text, o.total_amount_cents,
       COALESCE(SUM(ois.quantity), 0)::int,
       COUNT(DISTINCT so.id)::int,
       o.created_at
FROM orders o
LEFT JOIN sub_orders so ON so.order_id = o.id
LEFT JOIN order_item_snapshots ois ON ois.sub_order_id = so.id
WHERE o.user_id = $1
GROUP BY o.id
ORDER BY o.created_at DESC, o.id DESC
LIMIT $2 OFFSET $3`

		rows, err := tx.Query(ctx, query, userID, filter.PageSize, (filter.Page-1)*filter.PageSize)
		if err != nil {
			return err
		}
		defer rows.Close()

		orders := make([]model.OrderSummary, 0)
		for rows.Next() {
			var order model.OrderSummary
			var status string
			if err := rows.Scan(
				&order.ID,
				&status,
				&order.TotalAmountCents,
				&order.ItemCount,
				&order.ShopCount,
				&order.CreatedAt,
			); err != nil {
				return err
			}
			order.Status = model.OrderStatus(status)
			orders = append(orders, order)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		result = &model.OrderListResult{
			Orders:   orders,
			Page:     filter.Page,
			PageSize: filter.PageSize,
			Total:    total,
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository.ListOrders: %w", err)
	}
	return result, nil
}

func (r *OrderRepository) Get(ctx context.Context, userID int, orderID int) (*model.Order, error) {
	if err := r.cancelDueOrder(ctx, userID, orderID); err != nil {
		slog.ErrorContext(ctx, "cancel due order", "user_id", userID, "order_id", orderID, "error", err)
	}

	var order *model.Order
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}, func(tx pgx.Tx) error {
		var err error
		order, err = loadOrder(ctx, tx, userID, orderID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("repository.GetOrder: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) Cancel(ctx context.Context, userID int, orderID int) (*model.Order, error) {
	var result *model.Order

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		const lockQuery = `SELECT status::text FROM orders WHERE id = $1 AND user_id = $2 FOR UPDATE`

		var status string
		if err := tx.QueryRow(ctx, lockQuery, orderID, userID).Scan(&status); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.ErrOrderNotFound
			}
			return err
		}
		if model.OrderStatus(status) == model.OrderStatusCancelled {
			var err error
			result, err = loadOrder(ctx, tx, userID, orderID)
			return err
		}
		if model.OrderStatus(status) != model.OrderStatusPendingPayment {
			return apperr.ErrOrderCannotCancel
		}

		if err := r.cancelPendingOrder(ctx, tx, orderID); err != nil {
			return err
		}

		var err error
		result, err = loadOrder(ctx, tx, userID, orderID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("repository.CancelOrder: %w", err)
	}
	return result, nil
}

func (r *OrderRepository) cancelDueOrdersForUser(ctx context.Context, userID int) {
	const query = `SELECT id
FROM orders
WHERE user_id = $1
  AND status = 'pending_payment'
  AND created_at <= now() - make_interval(secs => $3)
ORDER BY created_at, id
LIMIT $2`

	rows, err := r.pool.Query(ctx, query, userID, dueOrderCleanupLimit, model.PendingPaymentTTL.Seconds())
	if err != nil {
		slog.ErrorContext(ctx, "find due orders", "user_id", userID, "error", err)
		return
	}

	orderIDs := make([]int, 0, dueOrderCleanupLimit)
	for rows.Next() {
		var orderID int
		if err := rows.Scan(&orderID); err != nil {
			slog.ErrorContext(ctx, "scan due order", "user_id", userID, "error", err)
			rows.Close()
			return
		}
		orderIDs = append(orderIDs, orderID)
	}
	if err := rows.Err(); err != nil {
		slog.ErrorContext(ctx, "read due orders", "user_id", userID, "error", err)
		rows.Close()
		return
	}
	rows.Close()

	for _, orderID := range orderIDs {
		if err := r.cancelDueOrder(ctx, userID, orderID); err != nil {
			slog.ErrorContext(ctx, "cancel due order", "user_id", userID, "order_id", orderID, "error", err)
		}
	}
}

func (r *OrderRepository) cancelDueOrder(ctx context.Context, userID int, orderID int) error {
	return pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		const query = `SELECT id
FROM orders
WHERE id = $1
  AND user_id = $2
  AND status = 'pending_payment'
  AND created_at <= now() - make_interval(secs => $3)
FOR UPDATE`

		var id int
		if err := tx.QueryRow(ctx, query, orderID, userID, model.PendingPaymentTTL.Seconds()).Scan(&id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}
		return r.cancelPendingOrder(ctx, tx, orderID)
	})
}

func (r *OrderRepository) cancelPendingOrder(ctx context.Context, tx pgx.Tx, orderID int) error {
	items, err := orderReservationItems(ctx, tx, orderID)
	if err != nil {
		return err
	}
	if err := r.reservation.Release(ctx, tx, items); err != nil {
		return mapReleaseError(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE sub_orders SET status = 'cancelled', updated_at = now() WHERE order_id = $1`, orderID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE orders SET status = 'cancelled', updated_at = now() WHERE id = $1`, orderID); err != nil {
		return err
	}
	return nil
}

func (r *OrderRepository) fillCartOrder(
	ctx context.Context,
	tx pgx.Tx,
	userID int,
	orderID int,
	items []model.CreateOrderItem,
) error {
	cartItems, err := lockCartItems(ctx, tx, userID, items)
	if err != nil {
		return err
	}

	skuIDs := make([]int64, 0, len(cartItems))
	for _, item := range cartItems {
		skuIDs = append(skuIDs, int64(item.SKUID))
	}
	catalog, err := lockCatalogSKUs(ctx, tx, skuIDs)
	if err != nil {
		return err
	}

	reservations := make(map[int]int, len(cartItems))
	for _, item := range cartItems {
		sku, ok := catalog[item.SKUID]
		if !ok || sku.Status != "on_shelf" || !sku.Enabled {
			return apperr.ErrProductUnavailable
		}
		reservations[item.SKUID] += item.Quantity
	}

	reservationItems := make([]inventorymodel.ReservationItem, 0, len(reservations))
	for skuID, quantity := range reservations {
		reservationItems = append(reservationItems, inventorymodel.ReservationItem{
			SKUID:    skuID,
			Quantity: quantity,
		})
	}
	sort.Slice(reservationItems, func(i, j int) bool {
		return reservationItems[i].SKUID < reservationItems[j].SKUID
	})
	if err := r.reservation.Freeze(ctx, tx, reservationItems); err != nil {
		return mapFreezeError(err)
	}

	groups := make(map[int]*shopGroup, len(cartItems))
	for _, item := range cartItems {
		sku := catalog[item.SKUID]
		group, ok := groups[sku.ShopID]
		if !ok {
			group = &shopGroup{
				shopID:   sku.ShopID,
				shopName: sku.ShopName,
				items:    make([]snapshotItem, 0),
			}
			groups[sku.ShopID] = group
		}
		lineAmount := item.UnitPriceCents * int64(item.Quantity)
		group.totalAmountCents += lineAmount
		group.items = append(group.items, snapshotItem{
			productID:       item.ProductID,
			skuID:           item.SKUID,
			productName:     item.ProductName,
			specs:           item.Specs,
			unitPriceCents:  item.UnitPriceCents,
			quantity:        item.Quantity,
			lineAmountCents: lineAmount,
		})
	}

	if err := persistOrder(ctx, tx, orderID, groups); err != nil {
		return err
	}

	itemIDs := make([]int64, 0, len(items))
	for _, item := range items {
		itemIDs = append(itemIDs, int64(item.CartItemID))
	}
	tag, err := tx.Exec(ctx, `DELETE FROM cart_items WHERE user_id = $1 AND id = ANY($2::bigint[])`, userID, itemIDs)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != int64(len(itemIDs)) {
		return apperr.ErrCartItemsChanged
	}
	return nil
}

func (r *OrderRepository) fillDirectOrder(
	ctx context.Context,
	tx pgx.Tx,
	orderID int,
	skuID int,
	quantity int,
) error {
	catalog, err := lockCatalogSKUs(ctx, tx, []int64{int64(skuID)})
	if err != nil {
		return err
	}
	sku, ok := catalog[skuID]
	if !ok || sku.Status != "on_shelf" || !sku.Enabled {
		return apperr.ErrProductUnavailable
	}

	if err := r.reservation.Freeze(ctx, tx, []inventorymodel.ReservationItem{{
		SKUID:    skuID,
		Quantity: quantity,
	}}); err != nil {
		return mapFreezeError(err)
	}

	lineAmount := sku.PriceCents * int64(quantity)
	groups := map[int]*shopGroup{
		sku.ShopID: {
			shopID:           sku.ShopID,
			shopName:         sku.ShopName,
			totalAmountCents: lineAmount,
			items: []snapshotItem{{
				productID:       sku.ProductID,
				skuID:           sku.SKUID,
				productName:     sku.ProductName,
				specs:           sku.Specs,
				unitPriceCents:  sku.PriceCents,
				quantity:        quantity,
				lineAmountCents: lineAmount,
			}},
		},
	}
	return persistOrder(ctx, tx, orderID, groups)
}

func persistOrder(ctx context.Context, tx pgx.Tx, orderID int, groups map[int]*shopGroup) error {
	shopIDs := make([]int, 0, len(groups))
	for shopID := range groups {
		shopIDs = append(shopIDs, shopID)
	}
	sort.Ints(shopIDs)

	var totalAmount int64
	for _, shopID := range shopIDs {
		group := groups[shopID]
		var subOrderID int
		const insertSubOrder = `INSERT INTO sub_orders (order_id, shop_id, shop_name, total_amount_cents)
VALUES ($1, $2, $3, $4)
RETURNING id`
		if err := tx.QueryRow(ctx, insertSubOrder, orderID, group.shopID, group.shopName, group.totalAmountCents).Scan(&subOrderID); err != nil {
			return err
		}

		for _, item := range group.items {
			const insertItem = `INSERT INTO order_item_snapshots
    (sub_order_id, product_id, sku_id, product_name, sku_specs, unit_price_cents, quantity)
VALUES ($1, $2, $3, $4, $5, $6, $7)`
			if _, err := tx.Exec(
				ctx,
				insertItem,
				subOrderID,
				item.productID,
				item.skuID,
				item.productName,
				item.specs,
				item.unitPriceCents,
				item.quantity,
			); err != nil {
				return err
			}
		}
		totalAmount += group.totalAmountCents
	}

	if _, err := tx.Exec(ctx, `UPDATE orders SET total_amount_cents = $1, updated_at = now() WHERE id = $2`, totalAmount, orderID); err != nil {
		return err
	}
	return nil
}

type lockedCartItem struct {
	ID             int
	ProductID      int
	SKUID          int
	ProductName    string
	Specs          map[string]string
	UnitPriceCents int64
	Quantity       int
}

func lockCartItems(ctx context.Context, db database.DBTX, userID int, requested []model.CreateOrderItem) ([]lockedCartItem, error) {
	itemIDs := make([]int64, 0, len(requested))
	quantities := make(map[int]int, len(requested))
	for _, item := range requested {
		itemIDs = append(itemIDs, int64(item.CartItemID))
		quantities[item.CartItemID] = item.Quantity
	}

	const query = `SELECT id, product_id, sku_id, product_name, sku_specs, unit_price_cents, quantity
FROM cart_items
WHERE user_id = $1 AND id = ANY($2::bigint[])
ORDER BY id
FOR UPDATE`

	rows, err := db.Query(ctx, query, userID, itemIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]lockedCartItem, 0, len(requested))
	for rows.Next() {
		var item lockedCartItem
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.SKUID,
			&item.ProductName,
			&item.Specs,
			&item.UnitPriceCents,
			&item.Quantity,
		); err != nil {
			return nil, err
		}
		if quantities[item.ID] != item.Quantity {
			return nil, apperr.ErrCartItemsChanged
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(items) != len(requested) {
		return nil, apperr.ErrCartItemsChanged
	}
	return items, nil
}

type catalogSKU struct {
	SKUID       int
	ProductID   int
	ProductName string
	Specs       map[string]string
	PriceCents  int64
	ShopID      int
	ShopName    string
	Status      string
	Enabled     bool
}

func lockCatalogSKUs(ctx context.Context, db database.DBTX, skuIDs []int64) (map[int]catalogSKU, error) {
	const query = `SELECT ps.id, p.id, p.name, ps.specs, ps.price_cents,
       p.shop_id, s.name, p.status::text, ps.is_enabled
FROM product_skus ps
JOIN products p ON p.id = ps.product_id
JOIN shops s ON s.id = p.shop_id
WHERE ps.id = ANY($1::bigint[])
ORDER BY p.id, ps.id
FOR SHARE OF p, ps`

	rows, err := db.Query(ctx, query, skuIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int]catalogSKU, len(skuIDs))
	for rows.Next() {
		var sku catalogSKU
		if err := rows.Scan(
			&sku.SKUID,
			&sku.ProductID,
			&sku.ProductName,
			&sku.Specs,
			&sku.PriceCents,
			&sku.ShopID,
			&sku.ShopName,
			&sku.Status,
			&sku.Enabled,
		); err != nil {
			return nil, err
		}
		result[sku.SKUID] = sku
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

type shopGroup struct {
	shopID           int
	shopName         string
	totalAmountCents int64
	items            []snapshotItem
}

type snapshotItem struct {
	productID       int
	skuID           int
	productName     string
	specs           map[string]string
	unitPriceCents  int64
	quantity        int
	lineAmountCents int64
}

func claimCheckoutKey(ctx context.Context, db database.DBTX, userID int, checkout checkoutDetails) (int, bool, error) {
	const insertQuery = `INSERT INTO orders
    (user_id, checkout_key, request_hash, receiver_name, receiver_phone, shipping_address)
VALUES ($1, $2::uuid, $3, $4, $5, $6)
ON CONFLICT (user_id, checkout_key) DO NOTHING
RETURNING id`

	var orderID int
	err := db.QueryRow(
		ctx,
		insertQuery,
		userID,
		checkout.CheckoutKey,
		checkout.RequestHash,
		checkout.ReceiverName,
		checkout.ReceiverPhone,
		checkout.ShippingAddress,
	).Scan(&orderID)
	if err == nil {
		return orderID, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, err
	}

	const existingQuery = `SELECT id, request_hash FROM orders WHERE user_id = $1 AND checkout_key = $2::uuid`
	var existingID int
	var existingHash string
	if err := db.QueryRow(ctx, existingQuery, userID, checkout.CheckoutKey).Scan(&existingID, &existingHash); err != nil {
		return 0, false, err
	}
	if existingHash != checkout.RequestHash {
		return 0, false, apperr.ErrCheckoutConflict
	}
	return existingID, false, nil
}

func orderReservationItems(ctx context.Context, db database.DBTX, orderID int) ([]inventorymodel.ReservationItem, error) {
	const query = `SELECT ois.sku_id, SUM(ois.quantity)::int
FROM order_item_snapshots ois
JOIN sub_orders so ON so.id = ois.sub_order_id
WHERE so.order_id = $1
GROUP BY ois.sku_id
ORDER BY ois.sku_id`

	rows, err := db.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]inventorymodel.ReservationItem, 0)
	for rows.Next() {
		var item inventorymodel.ReservationItem
		if err := rows.Scan(&item.SKUID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func mapFreezeError(err error) error {
	switch {
	case errors.Is(err, inventoryapperr.ErrInsufficientStock):
		return apperr.ErrInsufficientStock
	case errors.Is(err, inventoryapperr.ErrInventoryItemNotFound):
		return apperr.ErrProductUnavailable
	default:
		return err
	}
}

func mapReleaseError(err error) error {
	if errors.Is(err, inventoryapperr.ErrReservationInvalid) || errors.Is(err, inventoryapperr.ErrInventoryItemNotFound) {
		return fmt.Errorf("%w: %v", apperr.ErrOrderStateInvalid, err)
	}
	return err
}

func loadOrder(ctx context.Context, db database.DBTX, userID int, orderID int) (*model.Order, error) {
	const orderQuery = `SELECT id, user_id, checkout_key::text, status::text, total_amount_cents,
       receiver_name, receiver_phone, shipping_address, paid_at, created_at, updated_at
FROM orders
WHERE id = $1 AND user_id = $2`

	var order model.Order
	var status string
	if err := db.QueryRow(ctx, orderQuery, orderID, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.CheckoutKey,
		&status,
		&order.TotalAmountCents,
		&order.ReceiverName,
		&order.ReceiverPhone,
		&order.ShippingAddress,
		&order.PaidAt,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrOrderNotFound
		}
		return nil, err
	}
	order.Status = model.OrderStatus(status)
	order.SubOrders = make([]model.SubOrder, 0)

	const subOrderQuery = `SELECT id, shop_id, shop_name, status::text, total_amount_cents, created_at, updated_at
FROM sub_orders
WHERE order_id = $1
ORDER BY id`

	rows, err := db.Query(ctx, subOrderQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subOrderIndex := make(map[int]int)
	for rows.Next() {
		var subOrder model.SubOrder
		var subStatus string
		if err := rows.Scan(
			&subOrder.ID,
			&subOrder.ShopID,
			&subOrder.ShopName,
			&subStatus,
			&subOrder.TotalAmountCents,
			&subOrder.CreatedAt,
			&subOrder.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subOrder.OrderID = orderID
		subOrder.Status = model.SubOrderStatus(subStatus)
		subOrder.Items = make([]model.OrderItem, 0)
		subOrderIndex[subOrder.ID] = len(order.SubOrders)
		order.SubOrders = append(order.SubOrders, subOrder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	const itemQuery = `SELECT ois.id, ois.sub_order_id, ois.product_id, ois.sku_id,
       ois.product_name, ois.sku_specs, ois.unit_price_cents, ois.quantity, ois.line_amount_cents
FROM order_item_snapshots ois
JOIN sub_orders so ON so.id = ois.sub_order_id
WHERE so.order_id = $1
ORDER BY ois.sub_order_id, ois.id`

	itemRows, err := db.Query(ctx, itemQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item model.OrderItem
		if err := itemRows.Scan(
			&item.ID,
			&item.SubOrderID,
			&item.ProductID,
			&item.SKUID,
			&item.ProductName,
			&item.Specs,
			&item.UnitPriceCents,
			&item.Quantity,
			&item.LineAmountCents,
		); err != nil {
			return nil, err
		}
		index, ok := subOrderIndex[item.SubOrderID]
		if !ok {
			continue
		}
		order.SubOrders[index].Items = append(order.SubOrders[index].Items, item)
	}
	if err := itemRows.Err(); err != nil {
		return nil, err
	}
	return &order, nil
}
