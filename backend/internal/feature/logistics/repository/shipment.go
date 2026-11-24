package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/logistics/apperr"
	"pinshaoshao/internal/feature/logistics/model"
	"pinshaoshao/internal/feature/order/shippingport"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const shipmentColumns = `id, shipment_no::text, sub_order_id, order_id, user_id, shop_id,
provider, provider_shipment_no::text, tracking_no, carrier, status::text, item_summary,
receiver_name, receiver_phone, receiver_address, origin_city, destination_city,
current_location, shipped_at, delivered_at, created_at, updated_at`

const summaryColumns = `id, shipment_no::text, sub_order_id, order_id, status::text,
carrier, tracking_no, item_summary, receiver_name, receiver_phone,
current_location, shipped_at, delivered_at, created_at, updated_at`

type ShipmentRepository struct {
	pool   *pgxpool.Pool
	orders shippingport.Lifecycle
}

func NewShipmentRepository(pool *pgxpool.Pool, orders shippingport.Lifecycle) *ShipmentRepository {
	return &ShipmentRepository{pool: pool, orders: orders}
}

func (r *ShipmentRepository) GetBySubOrder(
	ctx context.Context,
	shopID int,
	subOrderID int,
) (*model.Shipment, bool, error) {
	const query = `SELECT ` + shipmentColumns + `
FROM shipments
WHERE shop_id = $1 AND sub_order_id = $2`

	shipment, err := scanShipment(r.pool.QueryRow(ctx, query, shopID, subOrderID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("repository.GetShipmentBySubOrder: %w", err)
	}
	events, err := loadEvents(ctx, r.pool, shipment.ID)
	if err != nil {
		return nil, false, fmt.Errorf("repository.GetShipmentBySubOrder: %w", err)
	}
	shipment.Events = events
	return shipment, true, nil
}

func (r *ShipmentRepository) Create(
	ctx context.Context,
	shipment *model.Shipment,
	target *shippingport.Target,
) (*model.Shipment, bool, error) {
	var result *model.Shipment
	created := false

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		lockedTarget, transitioned, err := r.orders.ShipSubOrder(
			ctx,
			tx,
			target.ShopID,
			target.SubOrderID,
			shipment.ShippedAt,
		)
		if err != nil {
			return mapOrderError(err)
		}
		shipment.UserID = lockedTarget.UserID
		shipment.OrderID = lockedTarget.OrderID
		if !transitioned {
			existing, err := getShipmentBySubOrder(
				ctx,
				tx,
				target.ShopID,
				target.SubOrderID,
			)
			if err != nil {
				return err
			}
			events, err := loadEvents(ctx, tx, existing.ID)
			if err != nil {
				return err
			}
			existing.Events = events
			result = existing
			return nil
		}

		const query = `INSERT INTO shipments
    (shipment_no, sub_order_id, order_id, user_id, shop_id, provider,
     provider_shipment_no, tracking_no, carrier, status, item_summary,
     receiver_name, receiver_phone, receiver_address, origin_city,
     destination_city, current_location, shipped_at, delivered_at)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::uuid, $8, $9,
        $10::shipment_status, $11, $12, $13, $14, $15, $16, $17, $18, $19)
ON CONFLICT DO NOTHING
RETURNING ` + shipmentColumns

		inserted, err := scanShipment(tx.QueryRow(
			ctx,
			query,
			shipment.ShipmentNo,
			shipment.SubOrderID,
			shipment.OrderID,
			shipment.UserID,
			shipment.ShopID,
			shipment.Provider,
			shipment.ProviderShipmentNo,
			shipment.ProviderTrackingNo,
			shipment.Carrier,
			shipment.Status,
			shipment.ItemSummary,
			shipment.ReceiverName,
			shipment.ReceiverPhone,
			shipment.ReceiverAddress,
			shipment.OriginCity,
			shipment.DestinationCity,
			shipment.CurrentLocation,
			shipment.ShippedAt,
			shipment.DeliveredAt,
		))
		if err == nil {
			inserted.Events = make([]model.Event, 0)
			result = inserted
			created = true
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("insert shipment: %w", err)
		}

		existing, err := getShipmentBySubOrder(ctx, tx, target.ShopID, target.SubOrderID)
		if err != nil {
			return err
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
		return nil, false, fmt.Errorf("repository.CreateShipment: %w", err)
	}
	return result, created, nil
}

func (r *ShipmentRepository) GetForMerchant(
	ctx context.Context,
	shopID int,
	shipmentNo string,
) (*model.Shipment, error) {
	shipment, err := getShipment(ctx, r.pool, shopID, 0, shipmentNo, false)
	if err != nil {
		return nil, fmt.Errorf("repository.GetShipmentForMerchant: %w", err)
	}
	events, err := loadEvents(ctx, r.pool, shipment.ID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetShipmentForMerchant: %w", err)
	}
	shipment.Events = events
	return shipment, nil
}

func (r *ShipmentRepository) ListForMerchant(
	ctx context.Context,
	shopID int,
	filter model.ListFilter,
) (*model.ListResult, error) {
	where, args := shipmentFilter(shopID, 0, filter)

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM shipments `+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("repository.ListShipmentsForMerchant: count: %w", err)
	}

	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	query := fmt.Sprintf(`SELECT `+summaryColumns+`
FROM shipments
%s
ORDER BY created_at DESC, id DESC
LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex)
	listArgs := append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, fmt.Errorf("repository.ListShipmentsForMerchant: %w", err)
	}
	defer rows.Close()

	shipments := make([]model.Summary, 0)
	for rows.Next() {
		summary, err := scanSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListShipmentsForMerchant: %w", err)
		}
		shipments = append(shipments, *summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListShipmentsForMerchant: %w", err)
	}
	return &model.ListResult{
		Shipments: shipments,
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Total:     total,
	}, nil
}

func (r *ShipmentRepository) GetForBuyer(
	ctx context.Context,
	userID int,
	shipmentNo string,
) (*model.Shipment, error) {
	shipment, err := getShipment(ctx, r.pool, 0, userID, shipmentNo, false)
	if err != nil {
		return nil, fmt.Errorf("repository.GetShipmentForBuyer: %w", err)
	}
	events, err := loadEvents(ctx, r.pool, shipment.ID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetShipmentForBuyer: %w", err)
	}
	shipment.Events = events
	return shipment, nil
}

func (r *ShipmentRepository) ListForBuyerOrder(
	ctx context.Context,
	userID int,
	orderID int,
) ([]model.Summary, error) {
	const query = `SELECT ` + summaryColumns + `
FROM shipments
WHERE user_id = $1 AND order_id = $2
ORDER BY created_at, id`

	rows, err := r.pool.Query(ctx, query, userID, orderID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListShipmentsForBuyerOrder: %w", err)
	}
	defer rows.Close()

	shipments := make([]model.Summary, 0)
	for rows.Next() {
		summary, err := scanSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListShipmentsForBuyerOrder: %w", err)
		}
		shipments = append(shipments, *summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListShipmentsForBuyerOrder: %w", err)
	}
	return shipments, nil
}

func (r *ShipmentRepository) ApplyEvent(
	ctx context.Context,
	callback model.Callback,
) (*model.Shipment, error) {
	var result *model.Shipment

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		identity, err := getShipment(ctx, tx, 0, 0, callback.ShipmentNo, false)
		if err != nil {
			return err
		}
		if identity.ProviderShipmentNo != callback.ProviderShipmentNo {
			return apperr.ErrShipmentConflict
		}
		if err := r.orders.LockOrder(ctx, tx, identity.OrderID); err != nil {
			return err
		}

		shipment, err := getShipment(ctx, tx, 0, 0, callback.ShipmentNo, true)
		if err != nil {
			return err
		}
		if shipment.ProviderShipmentNo != callback.ProviderShipmentNo {
			return apperr.ErrShipmentConflict
		}

		var existingShipmentID int64
		err = tx.QueryRow(
			ctx,
			`SELECT shipment_id FROM shipment_events WHERE event_id = $1::uuid`,
			callback.EventID,
		).Scan(&existingShipmentID)
		switch {
		case err == nil && existingShipmentID != shipment.ID:
			return apperr.ErrShipmentConflict
		case err == nil:
			events, loadErr := loadEvents(ctx, tx, shipment.ID)
			if loadErr != nil {
				return loadErr
			}
			shipment.Events = events
			result = shipment
			return nil
		case !errors.Is(err, pgx.ErrNoRows):
			return fmt.Errorf("check shipment event id: %w", err)
		}

		previousStatus := shipment.Status
		if callback.Status.Rank() >= shipment.Status.Rank() {
			shipment.Status = callback.Status
			shipment.CurrentLocation = callback.Location
		}
		if callback.ProviderTrackingNo != "" {
			shipment.ProviderTrackingNo = callback.ProviderTrackingNo
		}
		if callback.Status.IsDelivered() && shipment.DeliveredAt == nil {
			deliveredAt := callback.OccurredAt
			shipment.DeliveredAt = &deliveredAt
		}
		shipment.UpdatedAt = time.Now().UTC()

		tag, err := tx.Exec(ctx, `UPDATE shipments
SET tracking_no = $2,
    status = $3::shipment_status,
    current_location = $4,
    delivered_at = $5,
    updated_at = $6
WHERE id = $1`, shipment.ID, shipment.ProviderTrackingNo, shipment.Status,
			shipment.CurrentLocation, shipment.DeliveredAt, shipment.UpdatedAt)
		if err != nil {
			return fmt.Errorf("update shipment from callback: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return apperr.ErrShipmentConflict
		}

		eventTag, err := tx.Exec(ctx, `INSERT INTO shipment_events
    (shipment_id, event_id, sequence, status, event_type, description, location, occurred_at)
VALUES ($1, $2::uuid, $3, $4::shipment_status, $5, $6, $7, $8)
ON CONFLICT (shipment_id, sequence) DO NOTHING`, shipment.ID, callback.EventID, callback.Sequence,
			callback.Status, callback.EventType, callback.Description,
			callback.Location, callback.OccurredAt)
		if err != nil {
			return fmt.Errorf("insert shipment event: %w", err)
		}
		if eventTag.RowsAffected() != 1 {
			return apperr.ErrShipmentConflict
		}

		if callback.Status.IsDelivered() || previousStatus.IsDelivered() {
			if err := r.orders.CompleteSubOrder(ctx, tx, shipment.SubOrderID, callback.OccurredAt); err != nil {
				return mapOrderError(err)
			}
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
		return nil, fmt.Errorf("repository.ApplyShipmentEvent: %w", err)
	}
	return result, nil
}

func getShipment(
	ctx context.Context,
	db database.DBTX,
	shopID int,
	userID int,
	shipmentNo string,
	forUpdate bool,
) (*model.Shipment, error) {
	conditions := []string{"shipment_no = $1::uuid"}
	args := []any{shipmentNo}
	if shopID > 0 {
		args = append(args, shopID)
		conditions = append(conditions, fmt.Sprintf("shop_id = $%d", len(args)))
	}
	if userID > 0 {
		args = append(args, userID)
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)))
	}

	query := `SELECT ` + shipmentColumns + ` FROM shipments WHERE ` + strings.Join(conditions, " AND ")
	if forUpdate {
		query += ` FOR UPDATE`
	}

	shipment, err := scanShipment(db.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrShipmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

func getShipmentBySubOrder(
	ctx context.Context,
	db database.DBTX,
	shopID int,
	subOrderID int,
) (*model.Shipment, error) {
	const query = `SELECT ` + shipmentColumns + `
FROM shipments
WHERE shop_id = $1 AND sub_order_id = $2`

	shipment, err := scanShipment(db.QueryRow(ctx, query, shopID, subOrderID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrShipmentNotFound
	}
	if err != nil {
		return nil, err
	}
	return shipment, nil
}

func loadEvents(ctx context.Context, db database.DBTX, shipmentID int64) ([]model.Event, error) {
	const query = `SELECT id, shipment_id, event_id::text, sequence, status::text,
event_type, description, location, occurred_at, created_at
FROM shipment_events
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
			&event.EventID,
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

func shipmentFilter(shopID int, userID int, filter model.ListFilter) (string, []any) {
	conditions := []string{"shop_id = $1"}
	args := []any{shopID}
	if userID > 0 {
		conditions[0] = "user_id = $1"
		args[0] = userID
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, fmt.Sprintf("status = $%d::shipment_status", len(args)))
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
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
		&shipment.SubOrderID,
		&shipment.OrderID,
		&shipment.UserID,
		&shipment.ShopID,
		&shipment.Provider,
		&shipment.ProviderShipmentNo,
		&shipment.ProviderTrackingNo,
		&shipment.Carrier,
		&status,
		&shipment.ItemSummary,
		&shipment.ReceiverName,
		&shipment.ReceiverPhone,
		&shipment.ReceiverAddress,
		&shipment.OriginCity,
		&shipment.DestinationCity,
		&shipment.CurrentLocation,
		&shipment.ShippedAt,
		&shipment.DeliveredAt,
		&shipment.CreatedAt,
		&shipment.UpdatedAt,
	); err != nil {
		return nil, err
	}
	shipment.Status = model.Status(status)
	return &shipment, nil
}

func scanSummary(row rowScanner) (*model.Summary, error) {
	var summary model.Summary
	var status string
	if err := row.Scan(
		&summary.ID,
		&summary.ShipmentNo,
		&summary.SubOrderID,
		&summary.OrderID,
		&status,
		&summary.Carrier,
		&summary.ProviderTrackingNo,
		&summary.ItemSummary,
		&summary.ReceiverName,
		&summary.ReceiverPhone,
		&summary.CurrentLocation,
		&summary.ShippedAt,
		&summary.DeliveredAt,
		&summary.CreatedAt,
		&summary.UpdatedAt,
	); err != nil {
		return nil, err
	}
	summary.Status = model.Status(status)
	return &summary, nil
}

func mapOrderError(err error) error {
	switch {
	case errors.Is(err, shippingport.ErrSubOrderNotFound):
		return apperr.ErrShipmentNotFound
	case errors.Is(err, shippingport.ErrSubOrderNotShippable):
		return apperr.ErrOrderNotShippable
	default:
		return err
	}
}
