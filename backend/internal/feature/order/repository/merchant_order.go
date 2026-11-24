package repository

import (
	"context"
	"fmt"
	"strings"

	"pinshaoshao/internal/feature/order/model"
)

func (r *OrderRepository) ListMerchant(ctx context.Context, shopID int, filter model.MerchantOrderListFilter) (*model.MerchantOrderListResult, error) {
	where, args := merchantOrderFilter(shopID, filter)

	var total int64
	countQuery := `SELECT COUNT(*) FROM sub_orders so ` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("repository.ListMerchantOrders: count: %w", err)
	}

	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	query := fmt.Sprintf(`SELECT so.id, so.order_id, so.status::text, so.total_amount_cents,
       COALESCE(SUM(ois.quantity), 0)::int,
       o.receiver_name, o.receiver_phone, o.shipping_address, o.paid_at, so.created_at
FROM sub_orders so
JOIN orders o ON o.id = so.order_id
LEFT JOIN order_item_snapshots ois ON ois.sub_order_id = so.id
%s
GROUP BY so.id, o.id
ORDER BY so.created_at DESC, so.id DESC
LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex)
	listArgs := append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, fmt.Errorf("repository.ListMerchantOrders: %w", err)
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
			&order.ReceiverName,
			&order.ReceiverPhone,
			&order.ShippingAddress,
			&order.PaidAt,
			&order.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository.ListMerchantOrders: %w", err)
		}
		order.Status = model.SubOrderStatus(status)
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListMerchantOrders: %w", err)
	}
	return &model.MerchantOrderListResult{
		Orders:   orders,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Total:    total,
	}, nil
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
