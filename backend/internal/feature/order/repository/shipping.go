package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pinshaoshao/internal/database"
	inventoryapperr "pinshaoshao/internal/feature/inventory/apperr"
	inventorymodel "pinshaoshao/internal/feature/inventory/model"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/shippingport"

	"github.com/jackc/pgx/v5"
)

const maxShipmentItemSummaryLen = 200

func (r *OrderRepository) GetShipmentTarget(
	ctx context.Context,
	shopID int,
	subOrderID int,
) (*shippingport.Target, error) {
	target, err := loadShippingTarget(ctx, r.pool, shopID, subOrderID, false)
	if err != nil {
		return nil, fmt.Errorf("repository.GetShipmentTarget: %w", err)
	}
	return target, nil
}

func (r *OrderRepository) LockOrder(ctx context.Context, tx pgx.Tx, orderID int) error {
	if err := lockOrder(ctx, tx, orderID); err != nil {
		return fmt.Errorf("repository.LockOrder: %w", err)
	}
	return nil
}

func (r *OrderRepository) ShipSubOrder(
	ctx context.Context,
	tx pgx.Tx,
	shopID int,
	subOrderID int,
	shippedAt time.Time,
) (*shippingport.Target, bool, error) {
	target, err := loadShippingTarget(ctx, tx, shopID, subOrderID, false)
	if err != nil {
		return nil, false, fmt.Errorf("repository.ShipSubOrder: %w", err)
	}
	if err := lockOrder(ctx, tx, target.OrderID); err != nil {
		return nil, false, fmt.Errorf("repository.ShipSubOrder: %w", err)
	}
	target, err = loadShippingTarget(ctx, tx, shopID, subOrderID, true)
	if err != nil {
		return nil, false, fmt.Errorf("repository.ShipSubOrder: %w", err)
	}

	switch target.Status {
	case "shipped", "completed":
		return target, false, nil
	case "awaiting_shipment":
	default:
		return nil, false, shippingport.ErrSubOrderNotShippable
	}

	items, err := subOrderReservationItems(ctx, tx, subOrderID)
	if err != nil {
		return nil, false, fmt.Errorf("repository.ShipSubOrder: %w", err)
	}
	if err := r.reservation.Commit(ctx, tx, items); err != nil {
		return nil, false, fmt.Errorf("repository.ShipSubOrder: %w", mapCommitError(err))
	}

	tag, err := tx.Exec(ctx, `UPDATE sub_orders
SET status = 'shipped', updated_at = $2
WHERE id = $1 AND status = 'awaiting_shipment'`, subOrderID, shippedAt.UTC())
	if err != nil {
		return nil, false, fmt.Errorf("repository.ShipSubOrder: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return nil, false, shippingport.ErrSubOrderNotShippable
	}
	if err := recalculateOrderStatus(ctx, tx, target.OrderID); err != nil {
		return nil, false, fmt.Errorf("repository.ShipSubOrder: %w", err)
	}

	target.Status = "shipped"
	return target, true, nil
}

func (r *OrderRepository) CompleteSubOrder(
	ctx context.Context,
	tx pgx.Tx,
	subOrderID int,
	deliveredAt time.Time,
) error {
	target, err := loadShippingTarget(ctx, tx, 0, subOrderID, false)
	if err != nil {
		return fmt.Errorf("repository.CompleteSubOrder: %w", err)
	}
	if err := lockOrder(ctx, tx, target.OrderID); err != nil {
		return fmt.Errorf("repository.CompleteSubOrder: %w", err)
	}
	target, err = loadShippingTarget(ctx, tx, 0, subOrderID, true)
	if err != nil {
		return fmt.Errorf("repository.CompleteSubOrder: %w", err)
	}

	switch target.Status {
	case "completed":
		return nil
	case "shipped":
	default:
		return shippingport.ErrSubOrderNotShippable
	}

	tag, err := tx.Exec(ctx, `UPDATE sub_orders
SET status = 'completed', updated_at = $2
WHERE id = $1 AND status = 'shipped'`, subOrderID, deliveredAt.UTC())
	if err != nil {
		return fmt.Errorf("repository.CompleteSubOrder: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return shippingport.ErrSubOrderNotShippable
	}
	if err := recalculateOrderStatus(ctx, tx, target.OrderID); err != nil {
		return fmt.Errorf("repository.CompleteSubOrder: %w", err)
	}
	return nil
}

func loadShippingTarget(
	ctx context.Context,
	db database.DBTX,
	shopID int,
	subOrderID int,
	forUpdate bool,
) (*shippingport.Target, error) {
	conditions := []string{"so.id = $1"}
	args := []any{subOrderID}
	if shopID > 0 {
		args = append(args, shopID)
		conditions = append(conditions, fmt.Sprintf("so.shop_id = $%d", len(args)))
	}

	query := `SELECT so.id, so.order_id, o.user_id, so.shop_id, so.shop_name,
       so.status::text, o.receiver_name, o.receiver_phone, o.shipping_address
FROM sub_orders so
JOIN orders o ON o.id = so.order_id
WHERE ` + strings.Join(conditions, " AND ")
	if forUpdate {
		query += `
FOR UPDATE OF so`
	}

	var target shippingport.Target
	if err := db.QueryRow(ctx, query, args...).Scan(
		&target.SubOrderID,
		&target.OrderID,
		&target.UserID,
		&target.ShopID,
		&target.ShopName,
		&target.Status,
		&target.ReceiverName,
		&target.ReceiverPhone,
		&target.ShippingAddress,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, shippingport.ErrSubOrderNotFound
		}
		return nil, err
	}

	items, err := loadSubOrderItemSummary(ctx, db, subOrderID)
	if err != nil {
		return nil, err
	}
	target.ItemSummary = items
	return &target, nil
}

func loadSubOrderItemSummary(
	ctx context.Context,
	db database.DBTX,
	subOrderID int,
) (string, error) {
	const query = `SELECT product_name, quantity
FROM order_item_snapshots
WHERE sub_order_id = $1
ORDER BY id`

	rows, err := db.Query(ctx, query, subOrderID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	parts := make([]string, 0)
	for rows.Next() {
		var name string
		var quantity int
		if err := rows.Scan(&name, &quantity); err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("%s x%d", name, quantity))
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return "", apperr.ErrOrderStateInvalid
	}
	return truncateSummary(strings.Join(parts, ", "), maxShipmentItemSummaryLen), nil
}

func lockOrder(ctx context.Context, tx pgx.Tx, orderID int) error {
	var id int
	if err := tx.QueryRow(ctx, `SELECT id FROM orders WHERE id = $1 FOR UPDATE`, orderID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.ErrOrderNotFound
		}
		return err
	}
	return nil
}

func subOrderReservationItems(
	ctx context.Context,
	db database.DBTX,
	subOrderID int,
) ([]inventorymodel.ReservationItem, error) {
	const query = `SELECT sku_id, SUM(quantity)::int
FROM order_item_snapshots
WHERE sub_order_id = $1
GROUP BY sku_id
ORDER BY sku_id`

	rows, err := db.Query(ctx, query, subOrderID)
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

func recalculateOrderStatus(ctx context.Context, tx pgx.Tx, orderID int) error {
	const query = `SELECT COUNT(*),
       COUNT(*) FILTER (WHERE status = 'pending_payment'),
       COUNT(*) FILTER (WHERE status = 'awaiting_shipment'),
       COUNT(*) FILTER (WHERE status = 'shipped'),
       COUNT(*) FILTER (WHERE status = 'completed'),
       COUNT(*) FILTER (WHERE status = 'cancelled')
FROM sub_orders
WHERE order_id = $1`

	var total, pending, awaiting, shipped, completed, cancelled int
	if err := tx.QueryRow(ctx, query, orderID).Scan(
		&total,
		&pending,
		&awaiting,
		&shipped,
		&completed,
		&cancelled,
	); err != nil {
		return err
	}
	if total == 0 {
		return apperr.ErrOrderStateInvalid
	}

	status := ""
	switch {
	case completed == total:
		status = "completed"
	case shipped+completed == total:
		status = "shipped"
	case shipped+completed > 0:
		status = "partially_shipped"
	case awaiting == total:
		status = "paid"
	case cancelled == total:
		status = "cancelled"
	case pending == total:
		status = "pending_payment"
	default:
		return apperr.ErrOrderStateInvalid
	}

	if _, err := tx.Exec(ctx, `UPDATE orders SET status = $2, updated_at = now() WHERE id = $1`, orderID, status); err != nil {
		return err
	}
	return nil
}

func mapCommitError(err error) error {
	if errors.Is(err, inventoryapperr.ErrReservationInvalid) ||
		errors.Is(err, inventoryapperr.ErrInventoryItemNotFound) {
		return apperr.ErrOrderStateInvalid
	}
	return err
}

func truncateSummary(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return strings.TrimSpace(string(runes[:limit]))
}

var _ shippingport.Lifecycle = (*OrderRepository)(nil)
