package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/mock/logistics/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const notificationColumns = `id, event_id::text, shipment_id, event_type, target_url, payload,
status::text, attempt_count, max_attempts, next_attempt_at, locked_until,
last_attempt_at, delivered_at, last_error, created_at, updated_at`

const qualifiedNotificationColumns = `notification.id, notification.event_id::text,
notification.shipment_id, notification.event_type, notification.target_url, notification.payload,
notification.status::text, notification.attempt_count, notification.max_attempts,
notification.next_attempt_at, notification.locked_until, notification.last_attempt_at,
notification.delivered_at, notification.last_error, notification.created_at, notification.updated_at`

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(
	ctx context.Context,
	db database.DBTX,
	notification *model.Notification,
) error {
	const query = `INSERT INTO mock_logistics_notifications
    (event_id, shipment_id, event_type, target_url, payload, status, attempt_count,
     max_attempts, next_attempt_at, created_at, updated_at)
VALUES ($1::uuid, $2, $3, $4, $5::jsonb, $6::mock_logistics_notification_status, $7, $8, $9, $10, $10)
RETURNING id`

	if err := db.QueryRow(
		ctx,
		query,
		notification.EventID,
		notification.ShipmentID,
		notification.EventType,
		notification.TargetURL,
		string(notification.Payload),
		notification.Status,
		notification.AttemptCount,
		notification.MaxAttempts,
		notification.NextAttemptAt,
		notification.CreatedAt,
	).Scan(&notification.ID); err != nil {
		return fmt.Errorf("insert shipment notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ClaimDue(
	ctx context.Context,
	now time.Time,
	lease time.Duration,
	limit int,
) ([]*model.Notification, error) {
	if limit <= 0 {
		return nil, nil
	}

	notifications := make([]*model.Notification, 0, limit)
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		const recoverQuery = `UPDATE mock_logistics_notifications
SET status = 'exhausted',
    locked_until = NULL,
    last_error = COALESCE(last_error, 'delivery lease expired after the final attempt'),
    updated_at = $1
WHERE attempt_count >= max_attempts
  AND (
      status = 'pending'
      OR (status = 'delivering' AND locked_until <= $1)
  )`
		if _, err := tx.Exec(ctx, recoverQuery, now); err != nil {
			return err
		}

		const query = `WITH due AS (
    SELECT id
    FROM mock_logistics_notifications
    WHERE (
        (status = 'pending' AND next_attempt_at <= $1)
        OR (status = 'delivering' AND locked_until <= $1)
    )
      AND attempt_count < max_attempts
    ORDER BY next_attempt_at, id
    LIMIT $3
    FOR UPDATE SKIP LOCKED
)
UPDATE mock_logistics_notifications AS notification
SET status = 'delivering',
    attempt_count = attempt_count + 1,
    locked_until = $1 + make_interval(secs => $2),
    last_attempt_at = $1,
    updated_at = $1
FROM due
WHERE notification.id = due.id
RETURNING ` + qualifiedNotificationColumns

		rows, err := tx.Query(ctx, query, now, lease.Seconds(), limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			notification, err := scanNotification(rows)
			if err != nil {
				return err
			}
			notifications = append(notifications, notification)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("claim shipment notifications: %w", err)
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkSucceeded(ctx context.Context, id int64, now time.Time) error {
	const query = `UPDATE mock_logistics_notifications
SET status = 'succeeded',
    delivered_at = $2,
    locked_until = NULL,
    last_error = NULL,
    updated_at = $2
WHERE id = $1 AND status = 'delivering'`

	if _, err := r.pool.Exec(ctx, query, id, now); err != nil {
		return fmt.Errorf("mark shipment notification succeeded: %w", err)
	}
	return nil
}

func (r *NotificationRepository) MarkFailed(
	ctx context.Context,
	id int64,
	now time.Time,
	nextAttemptAt time.Time,
	lastError string,
) (model.NotificationStatus, error) {
	const query = `UPDATE mock_logistics_notifications
SET status = CASE
        WHEN attempt_count >= max_attempts THEN 'exhausted'::mock_logistics_notification_status
        ELSE 'pending'::mock_logistics_notification_status
    END,
    next_attempt_at = $3,
    locked_until = NULL,
    last_error = $4,
    updated_at = $2
WHERE id = $1 AND status = 'delivering'
RETURNING status::text`

	var status string
	if err := r.pool.QueryRow(ctx, query, id, now, nextAttemptAt, lastError).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrNotificationNotFound
		}
		return "", fmt.Errorf("mark shipment notification failed: %w", err)
	}
	return model.NotificationStatus(status), nil
}

func (r *NotificationRepository) ListByShipmentNo(
	ctx context.Context,
	shipmentNo string,
) ([]*model.Notification, error) {
	const query = `SELECT ` + qualifiedNotificationColumns + `
FROM mock_logistics_notifications AS notification
JOIN mock_shipments AS shipment ON shipment.id = notification.shipment_id
WHERE shipment.shipment_no = $1::uuid
ORDER BY notification.id`

	rows, err := r.pool.Query(ctx, query, shipmentNo)
	if err != nil {
		return nil, fmt.Errorf("list shipment notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*model.Notification, 0)
	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepository) Retry(
	ctx context.Context,
	shipmentNo string,
	notificationID int64,
	now time.Time,
) (*model.Notification, error) {
	var result *model.Notification

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		const lockQuery = `SELECT ` + qualifiedNotificationColumns + `
FROM mock_logistics_notifications AS notification
JOIN mock_shipments AS shipment ON shipment.id = notification.shipment_id
WHERE shipment.shipment_no = $1::uuid AND notification.id = $2
FOR UPDATE OF notification`

		notification, err := scanNotification(tx.QueryRow(ctx, lockQuery, shipmentNo, notificationID))
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrNotificationNotFound
		}
		if err != nil {
			return err
		}
		if notification.Status != model.NotificationStatusExhausted {
			return model.ErrNotificationNotRetryable
		}

		const updateQuery = `UPDATE mock_logistics_notifications
SET status = 'pending',
    attempt_count = 0,
    next_attempt_at = $2,
    locked_until = NULL,
    last_attempt_at = NULL,
    delivered_at = NULL,
    last_error = NULL,
    updated_at = $2
WHERE id = $1
RETURNING ` + notificationColumns

		notification, err = scanNotification(tx.QueryRow(ctx, updateQuery, notification.ID, now))
		if err != nil {
			return err
		}
		result = notification
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("retry shipment notification: %w", err)
	}
	return result, nil
}

func scanNotification(row rowScanner) (*model.Notification, error) {
	var notification model.Notification
	var status string
	if err := row.Scan(
		&notification.ID,
		&notification.EventID,
		&notification.ShipmentID,
		&notification.EventType,
		&notification.TargetURL,
		&notification.Payload,
		&status,
		&notification.AttemptCount,
		&notification.MaxAttempts,
		&notification.NextAttemptAt,
		&notification.LockedUntil,
		&notification.LastAttemptAt,
		&notification.DeliveredAt,
		&notification.LastError,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	); err != nil {
		return nil, err
	}
	notification.Status = model.NotificationStatus(status)
	return &notification, nil
}
