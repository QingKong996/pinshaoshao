package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pinshaoshao/internal/database"
	orderapperr "pinshaoshao/internal/feature/order/apperr"
	"pinshaoshao/internal/feature/order/paymentport"
	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const paymentColumns = `id, payment_no::text, order_id, provider, provider_transaction_id,
status::text, amount_cents, paid_at, created_at, updated_at`

type PaymentRepository struct {
	pool    *pgxpool.Pool
	orders  paymentport.Lifecycle
	refunds *RefundRepository
}

func NewPaymentRepository(
	pool *pgxpool.Pool,
	orders paymentport.Lifecycle,
	refunds *RefundRepository,
) *PaymentRepository {
	return &PaymentRepository{pool: pool, orders: orders, refunds: refunds}
}

func (r *PaymentRepository) CreateForOrder(
	ctx context.Context,
	userID int,
	orderID int,
	paymentNo string,
) (*model.Payment, bool, error) {
	var result *model.Payment
	created := false

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		target, err := r.orders.LockForPayment(ctx, tx, userID, orderID)
		if err != nil {
			return mapOrderError(err)
		}
		if target.Expired || target.Status != paymentport.OrderStatusPendingPayment {
			return apperr.ErrOrderNotPayable
		}

		existing, err := getByOrderForUpdate(ctx, tx, orderID)
		if err == nil {
			existing.OrderExpiresAt = target.ExpiresAt
			result = existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		const query = `INSERT INTO payments
    (payment_no, order_id, provider, status, amount_cents)
VALUES ($1::uuid, $2, $3, 'pending', $4)
RETURNING ` + paymentColumns
		createdPayment, err := scanPayment(tx.QueryRow(
			ctx,
			query,
			paymentNo,
			orderID,
			model.ProviderMockPay,
			target.TotalAmountCents,
		))
		if err != nil {
			return fmt.Errorf("insert payment: %w", err)
		}
		createdPayment.OrderExpiresAt = target.ExpiresAt
		result = createdPayment
		created = true
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("repository.CreatePayment: %w", err)
	}
	return result, created, nil
}

func (r *PaymentRepository) GetForUser(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*model.Payment, error) {
	orderID, err := paymentOrderID(ctx, r.pool, paymentNo)
	if err != nil {
		return nil, err
	}

	var result *model.Payment
	err = pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		target, err := r.orders.LockForPayment(ctx, tx, userID, orderID)
		if err != nil {
			return mapOrderError(err)
		}
		payment, err := getByPaymentNo(ctx, tx, paymentNo, false)
		if err != nil {
			return err
		}
		if payment.OrderID != orderID {
			return apperr.ErrPaymentConflict
		}
		payment.OrderExpiresAt = target.ExpiresAt
		if err := r.attachRefund(ctx, tx, payment); err != nil {
			return err
		}
		result = payment
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository.GetPayment: %w", err)
	}
	return result, nil
}

func (r *PaymentRepository) GetForPayment(
	ctx context.Context,
	userID int,
	paymentNo string,
) (*model.Payment, error) {
	var result *model.Payment

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		payment, err := getByPaymentNo(ctx, tx, paymentNo, false)
		if err != nil {
			return err
		}
		target, err := r.orders.LockForPayment(ctx, tx, userID, payment.OrderID)
		if err != nil {
			return mapOrderError(err)
		}
		payment, err = getByPaymentNo(ctx, tx, paymentNo, true)
		if err != nil {
			return err
		}

		switch payment.Status {
		case model.PaymentStatusSucceeded:
			if err := r.attachRefund(ctx, tx, payment); err != nil {
				return err
			}
			result = payment
			return nil
		case model.PaymentStatusClosed:
			return apperr.ErrPaymentClosed
		}
		switch target.Status {
		case paymentport.OrderStatusPendingPayment:
			payment.OrderExpiresAt = target.ExpiresAt
			result = payment
			return nil
		case paymentport.OrderStatusPaid:
			return apperr.ErrPaymentConflict
		default:
			return apperr.ErrOrderNotPayable
		}
	})
	if err != nil {
		return nil, fmt.Errorf("repository.GetPaymentForPayment: %w", err)
	}
	return result, nil
}

func (r *PaymentRepository) SetProviderTransaction(
	ctx context.Context,
	paymentNo string,
	providerTransactionID string,
) error {
	const query = `UPDATE payments
SET provider_transaction_id = $2, updated_at = now()
WHERE payment_no = $1::uuid
  AND status IN ('pending', 'succeeded')
  AND (provider_transaction_id IS NULL OR provider_transaction_id = $2)`

	tag, err := r.pool.Exec(ctx, query, paymentNo, providerTransactionID)
	if err != nil {
		return fmt.Errorf("set provider transaction: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	var status string
	if err := r.pool.QueryRow(
		ctx,
		`SELECT status::text FROM payments WHERE payment_no = $1::uuid`,
		paymentNo,
	).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.ErrPaymentNotFound
		}
		return fmt.Errorf("check payment after provider transaction: %w", err)
	}
	if model.PaymentStatus(status) == model.PaymentStatusClosed {
		return apperr.ErrPaymentClosed
	}
	return apperr.ErrPaymentConflict
}

func (r *PaymentRepository) ApplyPaidCallback(
	ctx context.Context,
	callback model.PaidCallback,
) (*model.Payment, error) {
	orderID, err := paymentOrderID(ctx, r.pool, callback.PaymentNo)
	if err != nil {
		return nil, err
	}

	var result *model.Payment
	err = pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		orderStatus, err := r.orders.SettlePayment(ctx, tx, orderID, callback.PaidAt)
		if err != nil {
			return mapOrderError(err)
		}

		payment, err := getByPaymentNo(ctx, tx, callback.PaymentNo, true)
		if err != nil {
			return err
		}
		if payment.OrderID != orderID {
			return apperr.ErrPaymentConflict
		}
		if payment.AmountCents != callback.AmountCents {
			return apperr.ErrPaymentConflict
		}
		if payment.Status == model.PaymentStatusSucceeded {
			if payment.ProviderTransactionID == nil ||
				*payment.ProviderTransactionID != callback.ProviderTransactionID {
				return apperr.ErrPaymentConflict
			}
		} else {
			if err := payment.MarkSucceeded(callback.ProviderTransactionID, callback.PaidAt); err != nil {
				return err
			}

			const query = `UPDATE payments
SET provider_transaction_id = $2, status = 'succeeded', paid_at = $3, updated_at = now()
WHERE id = $1 AND status IN ('pending', 'closed')`
			tag, err := tx.Exec(
				ctx,
				query,
				payment.ID,
				payment.ProviderTransactionID,
				payment.PaidAt,
			)
			if err != nil {
				return fmt.Errorf("apply paid callback: %w", err)
			}
			if tag.RowsAffected() != 1 {
				return apperr.ErrPaymentConflict
			}
		}

		if orderStatus == paymentport.OrderStatusCancelled {
			if r.refunds == nil {
				return apperr.ErrInternalServer
			}
			refund, err := r.refunds.EnsureForCancelledPayment(ctx, tx, payment)
			if err != nil {
				return err
			}
			payment.Refund = refund
		}
		result = payment
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository.ApplyPaidCallback: %w", err)
	}
	return result, nil
}

func (r *PaymentRepository) attachRefund(
	ctx context.Context,
	db database.DBTX,
	payment *model.Payment,
) error {
	if r.refunds == nil {
		return nil
	}
	refund, err := r.refunds.GetByPayment(ctx, db, payment.ID)
	if err != nil {
		return err
	}
	payment.Refund = refund
	return nil
}

func (r *PaymentRepository) CloseForOrder(
	ctx context.Context,
	db database.DBTX,
	orderID int,
) (*paymentport.CloseResult, error) {
	payment, err := getByOrderForUpdate(ctx, db, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("close payment for order: %w", err)
	}

	result := &paymentport.CloseResult{
		PaymentID:             payment.ID,
		PaymentNo:             payment.PaymentNo,
		ProviderTransactionID: payment.ProviderTransactionID,
	}
	switch payment.Status {
	case model.PaymentStatusClosed:
		result.Closed = true
		return result, nil
	case model.PaymentStatusSucceeded:
		result.AlreadySettled = true
		return result, nil
	}

	if err := payment.MarkClosed(time.Now()); err != nil {
		return nil, err
	}
	const query = `UPDATE payments
SET status = 'closed', updated_at = $2
WHERE id = $1 AND status = 'pending'`
	tag, err := db.Exec(ctx, query, payment.ID, payment.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("close payment for order: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return nil, apperr.ErrPaymentConflict
	}
	result.Closed = true
	return result, nil
}

func paymentOrderID(ctx context.Context, db database.DBTX, paymentNo string) (int, error) {
	var orderID int
	if err := db.QueryRow(
		ctx,
		`SELECT order_id FROM payments WHERE payment_no = $1::uuid`,
		paymentNo,
	).Scan(&orderID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, apperr.ErrPaymentNotFound
		}
		return 0, fmt.Errorf("read payment order: %w", err)
	}
	return orderID, nil
}

func getByOrderForUpdate(
	ctx context.Context,
	db database.DBTX,
	orderID int,
) (*model.Payment, error) {
	const query = `SELECT ` + paymentColumns + ` FROM payments WHERE order_id = $1 FOR UPDATE`
	return scanPayment(db.QueryRow(ctx, query, orderID))
}

func getByPaymentNo(
	ctx context.Context,
	db database.DBTX,
	paymentNo string,
	forUpdate bool,
) (*model.Payment, error) {
	query := `SELECT ` + paymentColumns + ` FROM payments WHERE payment_no = $1::uuid`
	if forUpdate {
		query += ` FOR UPDATE`
	}

	payment, err := scanPayment(db.QueryRow(ctx, query, paymentNo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return payment, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPayment(row rowScanner) (*model.Payment, error) {
	var payment model.Payment
	var status string
	if err := row.Scan(
		&payment.ID,
		&payment.PaymentNo,
		&payment.OrderID,
		&payment.Provider,
		&payment.ProviderTransactionID,
		&status,
		&payment.AmountCents,
		&payment.PaidAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return nil, err
	}
	payment.Status = model.PaymentStatus(status)
	return &payment, nil
}

func mapOrderError(err error) error {
	switch {
	case errors.Is(err, orderapperr.ErrOrderNotFound):
		return apperr.ErrOrderNotFound
	default:
		return err
	}
}
