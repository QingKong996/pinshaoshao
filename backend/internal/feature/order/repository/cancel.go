package repository

import (
	"context"
	"errors"
	"fmt"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"

	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) LockForCancel(
	ctx context.Context,
	db database.DBTX,
	userID int,
	orderID int,
) (*model.CancelTarget, error) {
	const query = `SELECT id, user_id, status::text
FROM orders
WHERE id = $1 AND user_id = $2
FOR UPDATE`
	target, err := lockCancelTarget(ctx, db, query, orderID, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrOrderNotFound
	}
	return target, err
}

func (r *OrderRepository) LockDueForCancel(
	ctx context.Context,
	db database.DBTX,
	orderID int,
) (*model.CancelTarget, error) {
	const query = `SELECT id, user_id, status::text
FROM orders
WHERE id = $1
  AND status = 'pending_payment'
  AND created_at <= now() - make_interval(secs => $2)
FOR UPDATE`
	target, err := lockCancelTarget(ctx, db, query, orderID, model.PendingPaymentTTL.Seconds())
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return target, err
}

func lockCancelTarget(
	ctx context.Context,
	db database.DBTX,
	query string,
	args ...any,
) (*model.CancelTarget, error) {
	var target model.CancelTarget
	var status string
	if err := db.QueryRow(ctx, query, args...).Scan(
		&target.OrderID,
		&target.UserID,
		&status,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("lock order for cancellation: %w", err)
	}
	target.Status = model.OrderStatus(status)

	items, err := cancellationItems(ctx, db, target.OrderID)
	if err != nil {
		return nil, err
	}
	target.Items = items
	return &target, nil
}

func cancellationItems(
	ctx context.Context,
	db database.DBTX,
	orderID int,
) ([]model.ReservationItem, error) {
	const query = `SELECT ois.sku_id, SUM(ois.quantity)::int
FROM order_item_snapshots ois
JOIN sub_orders so ON so.id = ois.sub_order_id
WHERE so.order_id = $1
GROUP BY ois.sku_id
ORDER BY ois.sku_id`

	rows, err := db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("load order reservation items: %w", err)
	}
	defer rows.Close()

	items := make([]model.ReservationItem, 0)
	for rows.Next() {
		var item model.ReservationItem
		if err := rows.Scan(&item.SKUID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("scan order reservation item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read order reservation items: %w", err)
	}
	return items, nil
}

func (r *OrderRepository) MarkCancelled(ctx context.Context, db database.DBTX, orderID int) error {
	if _, err := db.Exec(
		ctx,
		`UPDATE sub_orders SET status = 'cancelled', updated_at = now()
WHERE order_id = $1 AND status = 'pending_payment'`,
		orderID,
	); err != nil {
		return fmt.Errorf("cancel sub-orders: %w", err)
	}
	tag, err := db.Exec(
		ctx,
		`UPDATE orders SET status = 'cancelled', updated_at = now()
WHERE id = $1 AND status = 'pending_payment'`,
		orderID,
	)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return apperr.ErrOrderCannotCancel
	}
	return nil
}

func (r *OrderRepository) Load(
	ctx context.Context,
	db database.DBTX,
	userID int,
	orderID int,
) (*model.Order, error) {
	return loadOrder(ctx, db, userID, orderID)
}

func (r *OrderRepository) ListDueOrderIDs(ctx context.Context, limit int) ([]int, error) {
	if limit <= 0 {
		return nil, nil
	}

	const query = `SELECT id
FROM orders
WHERE status = 'pending_payment'
  AND created_at <= now() - make_interval(secs => $2)
ORDER BY created_at, id
LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit, model.PendingPaymentTTL.Seconds())
	if err != nil {
		return nil, fmt.Errorf("list due orders: %w", err)
	}
	defer rows.Close()

	ids := make([]int, 0, limit)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan due order: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read due orders: %w", err)
	}
	return ids, nil
}
