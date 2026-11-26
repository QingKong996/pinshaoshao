package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pinshaoshao/internal/mock/payment/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const refundColumns = `id, refund_no::text, provider_refund_no::text, payment_id,
amount_cents, reason, status::text, refunded_at, created_at, updated_at`

type RefundRepository struct {
	pool *pgxpool.Pool
}

func NewRefundRepository(pool *pgxpool.Pool) *RefundRepository {
	return &RefundRepository{pool: pool}
}

func (r *RefundRepository) Refund(
	ctx context.Context,
	paymentNo string,
	input model.RefundInput,
	now time.Time,
) (*model.Refund, error) {
	var result *model.Refund
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		payment, err := getPaymentForUpdate(ctx, tx, paymentNo)
		if err != nil {
			return err
		}
		if payment.Status != model.PaymentStatusPaid {
			return model.ErrPaymentNotRefundable
		}
		if payment.AmountCents != input.AmountCents {
			return model.ErrRefundConflict
		}

		existing, err := getRefundByPaymentForUpdate(ctx, tx, payment.ID)
		if err == nil {
			if existing.RefundNo != input.RefundNo ||
				existing.AmountCents != input.AmountCents ||
				existing.Reason != input.Reason {
				return model.ErrRefundConflict
			}
			existing.PaymentNo = payment.PaymentNo
			result = existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		existing, err = getRefundByRefundNo(ctx, tx, input.RefundNo)
		if err == nil {
			if existing.PaymentID != payment.ID ||
				existing.AmountCents != input.AmountCents ||
				existing.Reason != input.Reason {
				return model.ErrRefundConflict
			}
			existing.PaymentNo = payment.PaymentNo
			result = existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		const query = `INSERT INTO mock_payment_refunds
    (refund_no, provider_refund_no, payment_id, amount_cents, reason, status, refunded_at)
VALUES ($1::uuid, $2::uuid, $3, $4, $5, 'succeeded', $6)
RETURNING ` + refundColumns
		refund, err := scanRefund(tx.QueryRow(
			ctx,
			query,
			input.RefundNo,
			uuid.NewString(),
			payment.ID,
			input.AmountCents,
			input.Reason,
			now.UTC(),
		))
		if err != nil {
			return err
		}
		refund.PaymentNo = payment.PaymentNo
		result = refund
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("refund mock payment: %w", err)
	}
	return result, nil
}

func getRefundByPaymentForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	paymentID int64,
) (*model.Refund, error) {
	const query = `SELECT ` + refundColumns + `
FROM mock_payment_refunds
WHERE payment_id = $1
FOR UPDATE`
	return scanRefund(tx.QueryRow(ctx, query, paymentID))
}

func getRefundByRefundNo(
	ctx context.Context,
	tx pgx.Tx,
	refundNo string,
) (*model.Refund, error) {
	const query = `SELECT ` + refundColumns + `
FROM mock_payment_refunds
WHERE refund_no = $1::uuid
FOR UPDATE`
	return scanRefund(tx.QueryRow(ctx, query, refundNo))
}

func scanRefund(row rowScanner) (*model.Refund, error) {
	var refund model.Refund
	var status string
	if err := row.Scan(
		&refund.ID,
		&refund.RefundNo,
		&refund.ProviderRefundNo,
		&refund.PaymentID,
		&refund.AmountCents,
		&refund.Reason,
		&status,
		&refund.RefundedAt,
		&refund.CreatedAt,
		&refund.UpdatedAt,
	); err != nil {
		return nil, err
	}
	refund.Status = model.RefundStatus(status)
	return &refund, nil
}
