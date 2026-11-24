package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pinshaoshao/internal/feature/order/apperr"
	ordermodel "pinshaoshao/internal/feature/order/model"
	"pinshaoshao/internal/feature/order/paymentport"

	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) LockForPayment(
	ctx context.Context,
	tx pgx.Tx,
	userID int,
	orderID int,
) (*paymentport.Order, error) {
	const query = `SELECT id, user_id, status::text, total_amount_cents, created_at,
       created_at <= now() - make_interval(secs => $3) AS expired
FROM orders
WHERE id = $1 AND user_id = $2
FOR UPDATE`

	var result paymentport.Order
	var status string
	var expired bool
	var createdAt time.Time
	if err := tx.QueryRow(ctx, query, orderID, userID, ordermodel.PendingPaymentTTL.Seconds()).Scan(
		&result.ID,
		&result.UserID,
		&status,
		&result.TotalAmountCents,
		&createdAt,
		&expired,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrOrderNotFound
		}
		return nil, fmt.Errorf("lock order for payment: %w", err)
	}
	result.Status = paymentport.OrderStatus(status)
	result.ExpiresAt = createdAt.Add(ordermodel.PendingPaymentTTL)

	if result.Status == paymentport.OrderStatusPendingPayment && expired {
		if err := r.cancelPendingOrder(ctx, tx, orderID); err != nil {
			return nil, fmt.Errorf("cancel expired order before payment: %w", err)
		}
		result.Status = paymentport.OrderStatusCancelled
	}
	return &result, nil
}

func (r *OrderRepository) SettlePayment(
	ctx context.Context,
	tx pgx.Tx,
	orderID int,
	paidAt time.Time,
) (paymentport.OrderStatus, error) {
	const lockQuery = `SELECT status::text,
       created_at <= now() - make_interval(secs => $2) AS expired
FROM orders
WHERE id = $1
FOR UPDATE`

	var status string
	var expired bool
	if err := tx.QueryRow(
		ctx,
		lockQuery,
		orderID,
		ordermodel.PendingPaymentTTL.Seconds(),
	).Scan(&status, &expired); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperr.ErrOrderNotFound
		}
		return "", fmt.Errorf("lock order for settlement: %w", err)
	}
	current := paymentport.OrderStatus(status)
	if current == paymentport.OrderStatusPendingPayment && expired {
		if err := r.cancelPendingOrder(ctx, tx, orderID); err != nil {
			return "", fmt.Errorf("cancel expired order during settlement: %w", err)
		}
		return paymentport.OrderStatusCancelled, nil
	}
	if current != paymentport.OrderStatusPendingPayment {
		return current, nil
	}

	paidAt = paidAt.UTC()
	if _, err := tx.Exec(ctx, `
UPDATE orders
SET status = 'paid', paid_at = $2, updated_at = now()
WHERE id = $1`, orderID, paidAt); err != nil {
		return "", fmt.Errorf("mark order paid: %w", err)
	}
	if _, err := tx.Exec(ctx, `
UPDATE sub_orders
SET status = 'awaiting_shipment', updated_at = now()
WHERE order_id = $1 AND status = 'pending_payment'`, orderID); err != nil {
		return "", fmt.Errorf("mark sub-orders awaiting shipment: %w", err)
	}
	return paymentport.OrderStatusPaid, nil
}
