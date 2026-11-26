package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/payment/apperr"
	"pinshaoshao/internal/feature/payment/model"
	"pinshaoshao/internal/outbox"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const refundColumns = `id, refund_no::text, payment_id, amount_cents, reason,
status::text, provider_refund_no, refunded_at, failed_at, last_error, created_at, updated_at`

type RefundOutboxWriter interface {
	Enqueue(ctx context.Context, db database.DBTX, event outbox.Event) error
}

type RefundRepository struct {
	pool   *pgxpool.Pool
	outbox RefundOutboxWriter
}

func NewRefundRepository(pool *pgxpool.Pool, writer RefundOutboxWriter) *RefundRepository {
	return &RefundRepository{pool: pool, outbox: writer}
}

func (r *RefundRepository) EnsureForCancelledPayment(
	ctx context.Context,
	db database.DBTX,
	payment *model.Payment,
) (*model.Refund, error) {
	const query = `INSERT INTO refunds (refund_no, payment_id, amount_cents, reason)
VALUES ($1::uuid, $2, $3, $4)
ON CONFLICT (payment_id) DO NOTHING
RETURNING ` + refundColumns

	refund, err := scanRefund(db.QueryRow(
		ctx,
		query,
		uuid.NewString(),
		payment.ID,
		payment.AmountCents,
		model.RefundReasonOrderCancelled,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		refund, err = r.getByPaymentID(ctx, db, payment.ID)
	}
	if err != nil {
		return nil, fmt.Errorf("create refund for cancelled payment: %w", err)
	}
	if refund == nil {
		return nil, apperr.ErrRefundNotFound
	}
	refund.PaymentNo = payment.PaymentNo
	refund.ProviderTransactionID = payment.ProviderTransactionID

	if r.outbox == nil {
		return nil, apperr.ErrInternalServer
	}
	payload, err := json.Marshal(struct {
		RefundNo string `json:"refund_no"`
	}{RefundNo: refund.RefundNo})
	if err != nil {
		return nil, fmt.Errorf("encode refund event: %w", err)
	}
	if err := r.outbox.Enqueue(ctx, db, outbox.Event{
		Key:           "payment.refund:" + payment.PaymentNo,
		Topic:         outbox.TopicPaymentRefundProvider,
		AggregateType: "refund",
		AggregateID:   refund.ID,
		Payload:       payload,
	}); err != nil {
		return nil, err
	}
	return refund, nil
}

func (r *RefundRepository) GetByPayment(
	ctx context.Context,
	db database.DBTX,
	paymentID int64,
) (*model.Refund, error) {
	refund, err := r.getByPaymentID(ctx, db, paymentID)
	if err != nil {
		return nil, fmt.Errorf("get refund by payment: %w", err)
	}
	return refund, nil
}

func (r *RefundRepository) GetForProcessing(
	ctx context.Context,
	refundNo string,
) (*model.Refund, error) {
	const query = `SELECT r.id, r.refund_no::text, r.payment_id, r.amount_cents, r.reason,
       r.status::text, r.provider_refund_no, r.refunded_at, r.failed_at, r.last_error,
       r.created_at, r.updated_at, p.payment_no::text, p.provider_transaction_id
FROM refunds r
JOIN payments p ON p.id = r.payment_id
WHERE r.refund_no = $1::uuid`

	var refund model.Refund
	var status string
	if err := r.pool.QueryRow(ctx, query, refundNo).Scan(
		&refund.ID,
		&refund.RefundNo,
		&refund.PaymentID,
		&refund.AmountCents,
		&refund.Reason,
		&status,
		&refund.ProviderRefundNo,
		&refund.RefundedAt,
		&refund.FailedAt,
		&refund.LastError,
		&refund.CreatedAt,
		&refund.UpdatedAt,
		&refund.PaymentNo,
		&refund.ProviderTransactionID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrRefundNotFound
		}
		return nil, fmt.Errorf("get refund for processing: %w", err)
	}
	refund.Status = model.RefundStatus(status)
	return &refund, nil
}

func (r *RefundRepository) MarkSucceeded(
	ctx context.Context,
	refundNo string,
	providerRefundNo string,
	refundedAt time.Time,
) error {
	const query = `UPDATE refunds
SET status = 'succeeded',
    provider_refund_no = $2,
    refunded_at = $3,
    failed_at = NULL,
    last_error = NULL,
    updated_at = now()
WHERE refund_no = $1::uuid
  AND status IN ('pending', 'succeeded')`
	tag, err := r.pool.Exec(ctx, query, refundNo, providerRefundNo, refundedAt.UTC())
	if err != nil {
		return fmt.Errorf("mark refund succeeded: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	var status string
	var storedProviderRefundNo *string
	if err := r.pool.QueryRow(
		ctx,
		`SELECT status::text, provider_refund_no FROM refunds WHERE refund_no = $1::uuid`,
		refundNo,
	).Scan(&status, &storedProviderRefundNo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.ErrRefundNotFound
		}
		return fmt.Errorf("check refund after success: %w", err)
	}
	if model.RefundStatus(status) == model.RefundStatusSucceeded &&
		storedProviderRefundNo != nil &&
		*storedProviderRefundNo == providerRefundNo {
		return nil
	}
	return apperr.ErrPaymentConflict
}

func (r *RefundRepository) MarkFailed(
	ctx context.Context,
	refundNo string,
	cause string,
) error {
	const query = `UPDATE refunds
SET status = 'failed',
    failed_at = now(),
    last_error = $2,
    updated_at = now()
WHERE refund_no = $1::uuid AND status = 'pending'`
	tag, err := r.pool.Exec(ctx, query, refundNo, cause)
	if err != nil {
		return fmt.Errorf("mark refund failed: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	var status string
	if err := r.pool.QueryRow(
		ctx,
		`SELECT status::text FROM refunds WHERE refund_no = $1::uuid`,
		refundNo,
	).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.ErrRefundNotFound
		}
		return fmt.Errorf("check refund after failure: %w", err)
	}
	if model.RefundStatus(status) == model.RefundStatusFailed ||
		model.RefundStatus(status) == model.RefundStatusSucceeded {
		return nil
	}
	return apperr.ErrPaymentConflict
}

func (r *RefundRepository) getByPaymentID(
	ctx context.Context,
	db database.DBTX,
	paymentID int64,
) (*model.Refund, error) {
	const query = `SELECT ` + refundColumns + ` FROM refunds WHERE payment_id = $1`
	refund, err := scanRefund(db.QueryRow(ctx, query, paymentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return refund, nil
}

func scanRefund(row rowScanner) (*model.Refund, error) {
	var refund model.Refund
	var status string
	if err := row.Scan(
		&refund.ID,
		&refund.RefundNo,
		&refund.PaymentID,
		&refund.AmountCents,
		&refund.Reason,
		&status,
		&refund.ProviderRefundNo,
		&refund.RefundedAt,
		&refund.FailedAt,
		&refund.LastError,
		&refund.CreatedAt,
		&refund.UpdatedAt,
	); err != nil {
		return nil, err
	}
	refund.Status = model.RefundStatus(status)
	return &refund, nil
}
