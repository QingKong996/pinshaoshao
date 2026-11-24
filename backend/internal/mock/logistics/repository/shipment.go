package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/mock/logistics/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const shipmentColumns = `id, shipment_no::text, merchant_order_no, request_hash, tracking_no,
carrier, status::text, item_summary, sender_name, sender_phone, sender_address,
receiver_name, receiver_phone, receiver_address, origin_city, destination_city,
notify_url, current_location, dispatched_at, picked_up_at, delivered_at, created_at, updated_at`

type ShipmentRepository struct {
	pool                    *pgxpool.Pool
	notifications           *NotificationRepository
	maxNotificationAttempts int
}

func NewShipmentRepository(
	pool *pgxpool.Pool,
	notifications *NotificationRepository,
	maxNotificationAttempts int,
) *ShipmentRepository {
	if maxNotificationAttempts < 1 {
		maxNotificationAttempts = model.DefaultNotificationMaxAttempts
	}
	return &ShipmentRepository{
		pool:                    pool,
		notifications:           notifications,
		maxNotificationAttempts: maxNotificationAttempts,
	}
}

func (r *ShipmentRepository) Create(
	ctx context.Context,
	shipment *model.Shipment,
) (*model.Shipment, bool, error) {
	var result *model.Shipment
	created := false

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		const query = `INSERT INTO mock_shipments
    (shipment_no, merchant_order_no, request_hash, tracking_no, carrier, status,
     item_summary, sender_name, sender_phone, sender_address, receiver_name,
     receiver_phone, receiver_address, origin_city, destination_city, notify_url,
     current_location, created_at, updated_at)
VALUES ($1::uuid, $2, $3, $4, $5, $6::mock_shipment_status, $7, $8, $9, $10,
        $11, $12, $13, $14, $15, $16, $17, $18, $18)
ON CONFLICT (merchant_order_no) DO NOTHING
RETURNING ` + shipmentColumns

		inserted, err := scanShipment(tx.QueryRow(
			ctx,
			query,
			shipment.ShipmentNo,
			shipment.MerchantOrderNo,
			shipment.RequestHash,
			shipment.TrackingNo,
			shipment.Carrier,
			shipment.Status,
			shipment.ItemSummary,
			shipment.Sender.Name,
			shipment.Sender.Phone,
			shipment.Sender.Address,
			shipment.Receiver.Name,
			shipment.Receiver.Phone,
			shipment.Receiver.Address,
			shipment.OriginCity,
			shipment.DestinationCity,
			shipment.NotifyURL,
			shipment.CurrentLocation,
			shipment.CreatedAt,
		))
		if err == nil {
			event, err := insertInitialEvent(ctx, tx, inserted, shipment.CreatedAt)
			if err != nil {
				return err
			}
			inserted.Events = []model.Event{*event}
			result = inserted
			created = true
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("insert shipment: %w", err)
		}

		existing, err := getShipmentByMerchantOrderNo(ctx, tx, shipment.MerchantOrderNo)
		if err != nil {
			return err
		}
		if existing.RequestHash != shipment.RequestHash {
			return model.ErrShipmentConflict
		}
		events, err := loadEvents(ctx, tx, existing.ID)
		if err != nil {
			return err
		}
		existing.Events = events
		result = existing
		return nil
	})
	if err != nil {
		return nil, false, fmt.Errorf("create shipment: %w", err)
	}
	return result, created, nil
}

func (r *ShipmentRepository) Get(ctx context.Context, shipmentNo string) (*model.Shipment, error) {
	shipment, err := getShipment(ctx, r.pool, shipmentNo, false)
	if err != nil {
		return nil, err
	}
	events, err := loadEvents(ctx, r.pool, shipment.ID)
	if err != nil {
		return nil, err
	}
	shipment.Events = events
	return shipment, nil
}

func (r *ShipmentRepository) Transition(
	ctx context.Context,
	shipmentNo string,
	action model.Action,
	now time.Time,
	details model.EventDetails,
) (*model.Shipment, error) {
	var result *model.Shipment

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		shipment, err := getShipment(ctx, tx, shipmentNo, true)
		if err != nil {
			return err
		}
		if action == model.ActionOutbound && shipment.IsDispatched() {
			events, err := loadEvents(ctx, tx, shipment.ID)
			if err != nil {
				return err
			}
			shipment.Events = events
			result = shipment
			return nil
		}
		if action == model.ActionComplete && shipment.Status.IsTerminal() {
			events, err := loadEvents(ctx, tx, shipment.ID)
			if err != nil {
				return err
			}
			shipment.Events = events
			result = shipment
			return nil
		}

		transition, err := shipment.Apply(action, now)
		if err != nil {
			return err
		}
		shipment.CurrentLocation = details.Location
		if err := updateShipment(ctx, tx, shipment); err != nil {
			return err
		}

		event, err := insertEvent(ctx, tx, shipment, transition, details, now)
		if err != nil {
			return err
		}
		eventID := uuid.NewString()
		payload, err := model.NewEventPayload(eventID, shipment, event)
		if err != nil {
			return err
		}
		notification := model.NewNotification(
			eventID,
			shipment,
			event,
			payload,
			r.maxNotificationAttempts,
			now,
		)
		if err := r.notifications.Create(ctx, tx, notification); err != nil {
			return err
		}

		events, err := loadEvents(ctx, tx, shipment.ID)
		if err != nil {
			return err
		}
		shipment.Events = events
		result = shipment
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transition shipment: %w", err)
	}
	return result, nil
}

func getShipment(
	ctx context.Context,
	db database.DBTX,
	shipmentNo string,
	forUpdate bool,
) (*model.Shipment, error) {
	query := `SELECT ` + shipmentColumns + `
FROM mock_shipments
WHERE shipment_no = $1::uuid`
	if forUpdate {
		query += `
FOR UPDATE`
	}

	shipment, err := scanShipment(db.QueryRow(ctx, query, shipmentNo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrShipmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get shipment: %w", err)
	}
	return shipment, nil
}

func getShipmentByMerchantOrderNo(
	ctx context.Context,
	db database.DBTX,
	merchantOrderNo string,
) (*model.Shipment, error) {
	const query = `SELECT ` + shipmentColumns + `
FROM mock_shipments
WHERE merchant_order_no = $1`

	shipment, err := scanShipment(db.QueryRow(ctx, query, merchantOrderNo))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrShipmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get shipment by merchant order number: %w", err)
	}
	return shipment, nil
}

func updateShipment(ctx context.Context, tx pgx.Tx, shipment *model.Shipment) error {
	const query = `UPDATE mock_shipments
SET status = $2::mock_shipment_status,
    current_location = $3,
    dispatched_at = $4,
    picked_up_at = $5,
    delivered_at = $6,
    updated_at = $7
WHERE id = $1`

	tag, err := tx.Exec(
		ctx,
		query,
		shipment.ID,
		shipment.Status,
		shipment.CurrentLocation,
		shipment.DispatchedAt,
		shipment.PickedUpAt,
		shipment.DeliveredAt,
		shipment.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return model.ErrInvalidShipmentTransition
	}
	return nil
}

func insertInitialEvent(
	ctx context.Context,
	tx pgx.Tx,
	shipment *model.Shipment,
	now time.Time,
) (*model.Event, error) {
	event := &model.Event{
		ShipmentID:  shipment.ID,
		Sequence:    1,
		Status:      model.StatusPendingOutbound,
		EventType:   model.EventTypeShipmentCreated,
		Description: "等待商家出库",
		Location:    shipment.CurrentLocation,
		OccurredAt:  now.UTC(),
		CreatedAt:   now.UTC(),
	}
	const query = `INSERT INTO mock_logistics_events
    (shipment_id, sequence, status, event_type, description, location, occurred_at, created_at)
VALUES ($1, $2, $3::mock_shipment_status, $4, $5, $6, $7, $7)
RETURNING id`
	if err := tx.QueryRow(
		ctx,
		query,
		event.ShipmentID,
		event.Sequence,
		event.Status,
		event.EventType,
		event.Description,
		event.Location,
		event.OccurredAt,
	).Scan(&event.ID); err != nil {
		return nil, fmt.Errorf("insert initial shipment event: %w", err)
	}
	return event, nil
}

func insertEvent(
	ctx context.Context,
	tx pgx.Tx,
	shipment *model.Shipment,
	transition model.Transition,
	details model.EventDetails,
	now time.Time,
) (*model.Event, error) {
	var sequence int
	if err := tx.QueryRow(
		ctx,
		`SELECT COALESCE(MAX(sequence), 0) + 1 FROM mock_logistics_events WHERE shipment_id = $1`,
		shipment.ID,
	).Scan(&sequence); err != nil {
		return nil, fmt.Errorf("next shipment event sequence: %w", err)
	}

	event := &model.Event{
		ShipmentID:  shipment.ID,
		Sequence:    sequence,
		Status:      transition.Status,
		EventType:   transition.EventType,
		Description: details.Description,
		Location:    details.Location,
		OccurredAt:  now.UTC(),
		CreatedAt:   now.UTC(),
	}
	const query = `INSERT INTO mock_logistics_events
    (shipment_id, sequence, status, event_type, description, location, occurred_at, created_at)
VALUES ($1, $2, $3::mock_shipment_status, $4, $5, $6, $7, $7)
RETURNING id`
	if err := tx.QueryRow(
		ctx,
		query,
		event.ShipmentID,
		event.Sequence,
		event.Status,
		event.EventType,
		event.Description,
		event.Location,
		event.OccurredAt,
	).Scan(&event.ID); err != nil {
		return nil, fmt.Errorf("insert shipment event: %w", err)
	}
	return event, nil
}

func loadEvents(ctx context.Context, db database.DBTX, shipmentID int64) ([]model.Event, error) {
	const query = `SELECT id, shipment_id, sequence, status::text, event_type,
description, location, occurred_at, created_at
FROM mock_logistics_events
WHERE shipment_id = $1
ORDER BY sequence`

	rows, err := db.Query(ctx, query, shipmentID)
	if err != nil {
		return nil, fmt.Errorf("list shipment events: %w", err)
	}
	defer rows.Close()

	events := make([]model.Event, 0)
	for rows.Next() {
		var event model.Event
		var status string
		if err := rows.Scan(
			&event.ID,
			&event.ShipmentID,
			&event.Sequence,
			&status,
			&event.EventType,
			&event.Description,
			&event.Location,
			&event.OccurredAt,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		event.Status = model.Status(status)
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanShipment(row rowScanner) (*model.Shipment, error) {
	var shipment model.Shipment
	var status string
	if err := row.Scan(
		&shipment.ID,
		&shipment.ShipmentNo,
		&shipment.MerchantOrderNo,
		&shipment.RequestHash,
		&shipment.TrackingNo,
		&shipment.Carrier,
		&status,
		&shipment.ItemSummary,
		&shipment.Sender.Name,
		&shipment.Sender.Phone,
		&shipment.Sender.Address,
		&shipment.Receiver.Name,
		&shipment.Receiver.Phone,
		&shipment.Receiver.Address,
		&shipment.OriginCity,
		&shipment.DestinationCity,
		&shipment.NotifyURL,
		&shipment.CurrentLocation,
		&shipment.DispatchedAt,
		&shipment.PickedUpAt,
		&shipment.DeliveredAt,
		&shipment.CreatedAt,
		&shipment.UpdatedAt,
	); err != nil {
		return nil, err
	}
	shipment.Status = model.Status(status)
	return &shipment, nil
}
