package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pinshaoshao/internal/mockpayment/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const paymentColumns = `id, payment_no::text, merchant_order_no, request_hash,
status::text, amount_cents, currency, subject, notify_url, COALESCE(return_url, ''),
expires_at, paid_at, closed_at, created_at, updated_at`

type PaymentRepository struct {
	pool                    *pgxpool.Pool
	notifications           *NotificationRepository
	maxNotificationAttempts int
}

func NewPaymentRepository(
	pool *pgxpool.Pool,
	notifications *NotificationRepository,
	maxNotificationAttempts int,
) *PaymentRepository {
	if maxNotificationAttempts < 1 {
		maxNotificationAttempts = model.DefaultNotificationMaxAttempts
	}
	return &PaymentRepository{
		pool:                    pool,
		notifications:           notifications,
		maxNotificationAttempts: maxNotificationAttempts,
	}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *model.Payment) (*model.Payment, bool, error) {
	const query = `INSERT INTO mock_payment_orders
    (payment_no, merchant_order_no, request_hash, status, amount_cents, currency, subject,
     notify_url, return_url, expires_at, created_at, updated_at)
VALUES ($1::uuid, $2, $3, $4::mock_payment_status, $5, $6, $7, $8, $9, $10, $11, $11)
ON CONFLICT (merchant_order_no) DO NOTHING
RETURNING ` + paymentColumns

	created, err := scanPayment(r.pool.QueryRow(
		ctx,
		query,
		payment.PaymentNo,
		payment.MerchantOrderNo,
		payment.RequestHash,
		payment.Status,
		payment.AmountCents,
		payment.Currency,
		payment.Subject,
		payment.NotifyURL,
		nullableString(payment.ReturnURL),
		payment.ExpiresAt,
		payment.CreatedAt,
	))
	if err == nil {
		return created, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, fmt.Errorf("insert payment: %w", err)
	}

	existing, err := r.getByMerchantOrderNo(ctx, payment.MerchantOrderNo)
	if err != nil {
		return nil, false, err
	}
	if existing.RequestHash != payment.RequestHash {
		return nil, false, model.ErrPaymentConflict
	}
	return existing, false, nil
}

func (r *PaymentRepository) Get(ctx context.Context, paymentNo string) (*model.Payment, error) {
	const query = `SELECT ` + paymentColumns + ` FROM mock_payment_orders WHERE payment_no = $1::uuid`

	payment, err := scanPayment(r.pool.QueryRow(ctx, query, paymentNo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return payment, nil
}

func (r *PaymentRepository) Pay(ctx context.Context, paymentNo string, now time.Time) (*model.Payment, error) {
	var result *model.Payment
	expired := false

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		payment, err := getPaymentForUpdate(ctx, tx, paymentNo)
		if err != nil {
			return err
		}
		switch payment.Status {
		case model.PaymentStatusPaid:
			result = payment
			return nil
		case model.PaymentStatusClosed:
			return model.ErrPaymentClosed
		}

		if !now.UTC().Before(payment.ExpiresAt) {
			expired = true
			if err := payment.Apply(model.PaymentActionClose, now); err != nil {
				return err
			}
			if err := updatePayment(ctx, tx, payment); err != nil {
				return err
			}
			return nil
		}

		if err := payment.Apply(model.PaymentActionPay, now); err != nil {
			return err
		}
		if err := updatePayment(ctx, tx, payment); err != nil {
			return err
		}

		eventID := uuid.NewString()
		payload, err := model.PaymentPaidPayload(eventID, payment, now)
		if err != nil {
			return err
		}
		notification := model.NewPaidNotification(
			eventID,
			payment,
			payload,
			r.maxNotificationAttempts,
			now,
		)
		if err := r.notifications.Create(ctx, tx, notification); err != nil {
			return err
		}

		result = payment
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("pay payment: %w", err)
	}
	if expired {
		return nil, model.ErrPaymentExpired
	}
	return result, nil
}

func (r *PaymentRepository) Close(ctx context.Context, paymentNo string, now time.Time) (*model.Payment, error) {
	var result *model.Payment

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		payment, err := getPaymentForUpdate(ctx, tx, paymentNo)
		if err != nil {
			return err
		}
		switch payment.Status {
		case model.PaymentStatusClosed:
			result = payment
			return nil
		case model.PaymentStatusPaid:
			return model.ErrPaymentNotClosable
		}

		if err := payment.Apply(model.PaymentActionClose, now); err != nil {
			return err
		}
		if err := updatePayment(ctx, tx, payment); err != nil {
			return err
		}
		result = payment
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("close payment: %w", err)
	}
	return result, nil
}

func (r *PaymentRepository) CloseExpired(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]model.Payment, error) {
	if limit <= 0 {
		return nil, nil
	}

	closed := make([]model.Payment, 0, limit)
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		const query = `SELECT ` + paymentColumns + `
FROM mock_payment_orders
WHERE status = 'pending' AND expires_at <= $1
ORDER BY expires_at, id
LIMIT $2
FOR UPDATE SKIP LOCKED`

		rows, err := tx.Query(ctx, query, now, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		payments := make([]*model.Payment, 0, limit)
		for rows.Next() {
			payment, err := scanPayment(rows)
			if err != nil {
				return err
			}
			payments = append(payments, payment)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for _, payment := range payments {
			if err := payment.Apply(model.PaymentActionClose, now); err != nil {
				return err
			}
			if err := updatePayment(ctx, tx, payment); err != nil {
				return err
			}
			closed = append(closed, *payment)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("close expired payments: %w", err)
	}
	return closed, nil
}

func (r *PaymentRepository) getByMerchantOrderNo(ctx context.Context, merchantOrderNo string) (*model.Payment, error) {
	const query = `SELECT ` + paymentColumns + ` FROM mock_payment_orders WHERE merchant_order_no = $1`

	payment, err := scanPayment(r.pool.QueryRow(ctx, query, merchantOrderNo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrPaymentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get payment by merchant order number: %w", err)
	}
	return payment, nil
}

func getPaymentForUpdate(ctx context.Context, tx pgx.Tx, paymentNo string) (*model.Payment, error) {
	const query = `SELECT ` + paymentColumns + `
FROM mock_payment_orders
WHERE payment_no = $1::uuid
FOR UPDATE`

	payment, err := scanPayment(tx.QueryRow(ctx, query, paymentNo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return payment, nil
}

func updatePayment(ctx context.Context, tx pgx.Tx, payment *model.Payment) error {
	const query = `UPDATE mock_payment_orders
SET status = $2::mock_payment_status,
    paid_at = $3,
    closed_at = $4,
    updated_at = $5
WHERE id = $1 AND status = 'pending'`

	tag, err := tx.Exec(
		ctx,
		query,
		payment.ID,
		payment.Status,
		payment.PaidAt,
		payment.ClosedAt,
		payment.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return model.ErrInvalidPaymentTransition
	}
	return nil
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
		&payment.MerchantOrderNo,
		&payment.RequestHash,
		&status,
		&payment.AmountCents,
		&payment.Currency,
		&payment.Subject,
		&payment.NotifyURL,
		&payment.ReturnURL,
		&payment.ExpiresAt,
		&payment.PaidAt,
		&payment.ClosedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return nil, err
	}
	payment.Status = model.PaymentStatus(status)
	return &payment, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
