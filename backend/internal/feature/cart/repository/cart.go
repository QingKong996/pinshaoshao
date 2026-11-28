package repository

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/cart/apperr"
	"pinshaoshao/internal/feature/cart/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cartItemSelect = `SELECT ci.id, ci.user_id, ci.product_id, ci.sku_id, ci.product_name, ci.sku_specs,
       ci.unit_price_cents, ci.quantity, ci.created_at, ci.updated_at,
       p.status, ps.is_enabled, ii.total_stock - ii.frozen_stock
FROM cart_items ci
LEFT JOIN products p ON p.id = ci.product_id
LEFT JOIN product_skus ps ON ps.id = ci.sku_id
LEFT JOIN inventory_items ii ON ii.sku_id = ci.sku_id`

type CartRepository struct {
	pool *pgxpool.Pool
}

func NewCartRepository(pool *pgxpool.Pool) *CartRepository {
	return &CartRepository{pool: pool}
}

func (r *CartRepository) ListByUserID(ctx context.Context, userID int) ([]model.CartItem, error) {
	query := cartItemSelect + ` WHERE ci.user_id = $1 ORDER BY ci.created_at DESC, ci.id DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListByUserID: %w", err)
	}
	defer rows.Close()

	items := make([]model.CartItem, 0)
	for rows.Next() {
		item, err := scanCartItem(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListByUserID: %w", err)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListByUserID: %w", err)
	}
	return items, nil
}

// AddItem snapshots the current catalog data and atomically accumulates quantity
// when the user already has the same SKU.
func (r *CartRepository) AddItem(ctx context.Context, userID int, skuID int, quantity int) (*model.CartItem, error) {
	const query = `
WITH candidate AS (
    SELECT p.id AS product_id, ps.id AS sku_id, p.name AS product_name,
           ps.specs, ps.price_cents
    FROM product_skus ps
    JOIN products p ON p.id = ps.product_id
    JOIN inventory_items ii ON ii.sku_id = ps.id
    WHERE ps.id = $2
      AND p.status = 'on_shelf'
      AND ps.is_enabled
      AND ii.total_stock - ii.frozen_stock > 0
), upserted AS (
    INSERT INTO cart_items (user_id, product_id, sku_id, product_name, sku_specs, unit_price_cents, quantity)
    SELECT $1, product_id, sku_id, product_name, specs, price_cents, $3
    FROM candidate
    ON CONFLICT (user_id, sku_id) DO UPDATE
    SET quantity = cart_items.quantity + EXCLUDED.quantity,
        updated_at = now()
    RETURNING id, user_id, product_id, sku_id, product_name, sku_specs,
              unit_price_cents, quantity, created_at, updated_at
)
SELECT u.id, u.user_id, u.product_id, u.sku_id, u.product_name, u.sku_specs,
       u.unit_price_cents, u.quantity, u.created_at, u.updated_at,
       p.status, ps.is_enabled, ii.total_stock - ii.frozen_stock
FROM upserted u
LEFT JOIN products p ON p.id = u.product_id
LEFT JOIN product_skus ps ON ps.id = u.sku_id
LEFT JOIN inventory_items ii ON ii.sku_id = u.sku_id`

	item, err := scanCartItem(r.pool.QueryRow(ctx, query, userID, skuID, quantity))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, r.addItemError(ctx, skuID)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == database.NumericValueOutOfRangeCode {
			return nil, fmt.Errorf("repository.AddItem: %w", apperr.ErrInvalidCartQuantity)
		}
		return nil, fmt.Errorf("repository.AddItem: %w", err)
	}
	return item, nil
}

func (r *CartRepository) addItemError(ctx context.Context, skuID int) error {
	var id int
	if err := r.pool.QueryRow(ctx, `SELECT id FROM product_skus WHERE id = $1`, skuID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("repository.AddItem: %w", apperr.ErrProductNotFound)
		}
		return fmt.Errorf("repository.AddItem: %w", err)
	}
	return fmt.Errorf("repository.AddItem: %w", apperr.ErrProductUnavailable)
}

func (r *CartRepository) UpdateQuantity(ctx context.Context, userID int, itemID int, quantity int) (*model.CartItem, error) {
	const query = `
WITH updated AS (
    UPDATE cart_items
    SET quantity = $1, updated_at = now()
    WHERE id = $2 AND user_id = $3
    RETURNING id, user_id, product_id, sku_id, product_name, sku_specs,
              unit_price_cents, quantity, created_at, updated_at
)
SELECT u.id, u.user_id, u.product_id, u.sku_id, u.product_name, u.sku_specs,
       u.unit_price_cents, u.quantity, u.created_at, u.updated_at,
       p.status, ps.is_enabled, ii.total_stock - ii.frozen_stock
FROM updated u
LEFT JOIN products p ON p.id = u.product_id
LEFT JOIN product_skus ps ON ps.id = u.sku_id
LEFT JOIN inventory_items ii ON ii.sku_id = u.sku_id`

	item, err := scanCartItem(r.pool.QueryRow(ctx, query, quantity, itemID, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.UpdateQuantity: %w", apperr.ErrCartItemNotFound)
		}
		return nil, fmt.Errorf("repository.UpdateQuantity: %w", err)
	}
	return item, nil
}

func (r *CartRepository) DeleteItem(ctx context.Context, userID int, itemID int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM cart_items WHERE id = $1 AND user_id = $2`, itemID, userID)
	if err != nil {
		return fmt.Errorf("repository.DeleteItem: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository.DeleteItem: %w", apperr.ErrCartItemNotFound)
	}
	return nil
}

func scanCartItem(row pgx.Row) (*model.CartItem, error) {
	var item model.CartItem

	err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.SKUID,
		&item.ProductName,
		&item.Specs,
		&item.UnitPriceCents,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.ProductStatus,
		&item.SKUEnabled,
		&item.AvailableStock,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
