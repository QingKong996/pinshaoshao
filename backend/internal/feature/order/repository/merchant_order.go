package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/model"

	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) ListMerchant(ctx context.Context, shopID int, filter model.MerchantOrderListFilter) (*model.MerchantOrderListResult, error) {
	var result *model.MerchantOrderListResult
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}, func(tx pgx.Tx) error {
		where, args := merchantOrderFilter(shopID, filter)

		var total int64
		countQuery := `SELECT COUNT(*) FROM sub_orders so ` + where
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return fmt.Errorf("count: %w", err)
		}

		limitIndex := len(args) + 1
		offsetIndex := len(args) + 2
		query := fmt.Sprintf(`SELECT so.id, so.order_id, so.status::text, so.total_amount_cents,
       COALESCE(SUM(ois.quantity), 0),
       o.paid_at, so.created_at
FROM sub_orders so
JOIN orders o ON o.id = so.order_id
LEFT JOIN order_item_snapshots ois ON ois.sub_order_id = so.id
%s
GROUP BY so.id, o.id
ORDER BY so.created_at DESC, so.id DESC
LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex)
		listArgs := append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

		rows, err := tx.Query(ctx, query, listArgs...)
		if err != nil {
			return err
		}
		defer rows.Close()

		orders := make([]model.MerchantOrderSummary, 0)
		for rows.Next() {
			var order model.MerchantOrderSummary
			var status string
			if err := rows.Scan(
				&order.ID,
				&order.OrderID,
				&status,
				&order.TotalAmountCents,
				&order.ItemCount,
				&order.PaidAt,
				&order.CreatedAt,
			); err != nil {
				return err
			}
			order.Status = model.SubOrderStatus(status)
			orders = append(orders, order)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		result = &model.MerchantOrderListResult{
			Orders:   orders,
			Page:     filter.Page,
			PageSize: filter.PageSize,
			Total:    total,
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository.ListMerchantOrders: %w", err)
	}
	return result, nil
}

func (r *OrderRepository) GetMerchant(ctx context.Context, shopID int, subOrderID int) (*model.MerchantOrderDetail, error) {
	var order *model.MerchantOrderDetail
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}, func(tx pgx.Tx) error {
		var err error
		order, err = loadMerchantOrder(ctx, tx, shopID, subOrderID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("repository.GetMerchantOrder: %w", err)
	}
	return order, nil
}

func loadMerchantOrder(
	ctx context.Context,
	db database.DBTX,
	shopID int,
	subOrderID int,
) (*model.MerchantOrderDetail, error) {
	const orderQuery = `SELECT so.id, so.order_id, so.status::text, so.total_amount_cents,
       o.paid_at, so.created_at, so.updated_at
FROM sub_orders so
JOIN orders o ON o.id = so.order_id
WHERE so.id = $1 AND so.shop_id = $2`

	var order model.MerchantOrderDetail
	var status string
	if err := db.QueryRow(ctx, orderQuery, subOrderID, shopID).Scan(
		&order.ID,
		&order.OrderID,
		&status,
		&order.TotalAmountCents,
		&order.PaidAt,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrOrderNotFound
		}
		return nil, err
	}
	order.Status = model.SubOrderStatus(status)
	order.Items = make([]model.OrderItem, 0)

	const itemQuery = `SELECT id, product_id, sku_id, product_name, sku_specs,
       unit_price_cents, quantity, line_amount_cents
FROM order_item_snapshots
WHERE sub_order_id = $1
ORDER BY id`

	rows, err := db.Query(ctx, itemQuery, subOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		item.SubOrderID = subOrderID
		if err := rows.Scan(
			&item.ID,
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
		order.ItemCount += int64(item.Quantity)
		order.Items = append(order.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) GetMerchantReceiver(
	ctx context.Context,
	shopID int,
	subOrderID int,
) (*model.MerchantOrderReceiver, error) {
	const query = `SELECT so.status::text, o.receiver_name, o.receiver_phone, o.shipping_address
FROM sub_orders so
JOIN orders o ON o.id = so.order_id
WHERE so.id = $1 AND so.shop_id = $2`

	var status string
	var receiver model.MerchantOrderReceiver
	if err := r.pool.QueryRow(ctx, query, subOrderID, shopID).Scan(
		&status,
		&receiver.ReceiverName,
		&receiver.ReceiverPhone,
		&receiver.ShippingAddress,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrOrderNotFound
		}
		return nil, fmt.Errorf("repository.GetMerchantReceiver: %w", err)
	}

	switch model.SubOrderStatus(status) {
	case model.SubOrderStatusAwaitingShipment,
		model.SubOrderStatusShipped,
		model.SubOrderStatusCompleted:
		return &receiver, nil
	default:
		return nil, apperr.ErrReceiverUnavailable
	}
}

func merchantOrderFilter(shopID int, filter model.MerchantOrderListFilter) (string, []any) {
	conditions := []string{"so.shop_id = $1"}
	args := []any{shopID}

	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, fmt.Sprintf("so.status = $%d", len(args)))
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}
