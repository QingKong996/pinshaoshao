package repository

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/database"
	"pinshaoshao/internal/feature/inventory/apperr"
	"pinshaoshao/internal/feature/inventory/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const inventorySelect = `SELECT ps.id, ps.product_id, p.name, p.category_id, pc.name, ps.specs, ps.price_cents,
       ii.total_stock, ii.frozen_stock, ii.version, ps.is_enabled, ii.updated_at
FROM inventory_items ii
JOIN product_skus ps ON ps.id = ii.sku_id
JOIN products p ON p.id = ps.product_id
LEFT JOIN product_categories pc ON pc.id = p.category_id`

type InventoryRepository struct {
	pool *pgxpool.Pool
}

func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

// ListInventoryByShopID returns every inventory record of the shop ordered by available stock.
func (r *InventoryRepository) ListInventoryByShopID(ctx context.Context, shopID int) ([]model.InventoryItem, error) {
	query := inventorySelect + ` WHERE p.shop_id = $1 ORDER BY (ii.total_stock - ii.frozen_stock) ASC, p.id, ps.id`

	items, err := queryInventory(ctx, r.pool, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListInventoryByShopID: %w", err)
	}
	return items, nil
}

func (r *InventoryRepository) UpdateInventoryTotalStock(ctx context.Context, shopID int, skuID int, totalStock int, version int64) (*model.InventoryItem, error) {
	var item *model.InventoryItem

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		const lockQuery = `SELECT ii.frozen_stock, ii.version
FROM inventory_items ii
JOIN product_skus ps ON ps.id = ii.sku_id
JOIN products p ON p.id = ps.product_id
WHERE ii.sku_id = $1 AND p.shop_id = $2
FOR UPDATE OF ii`

		var frozenStock int
		var currentVersion int64
		if err := tx.QueryRow(ctx, lockQuery, skuID, shopID).Scan(&frozenStock, &currentVersion); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.ErrInventoryItemNotFound
			}
			return err
		}
		if currentVersion != version {
			return apperr.ErrInventoryVersion
		}
		if totalStock < frozenStock {
			return apperr.ErrInventoryBelowFrozen
		}

		const updateQuery = `UPDATE inventory_items
SET total_stock = $1, version = version + 1, updated_at = now()
WHERE sku_id = $2`
		if _, err := tx.Exec(ctx, updateQuery, totalStock, skuID); err != nil {
			return err
		}

		query := inventorySelect + ` WHERE ii.sku_id = $1 AND p.shop_id = $2`
		var err error
		item, err = queryInventoryItem(ctx, tx, query, skuID, shopID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("repository.UpdateInventoryTotalStock: %w", err)
	}
	return item, nil
}

func queryInventory(ctx context.Context, db database.DBTX, query string, args ...any) ([]model.InventoryItem, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.InventoryItem, 0)
	for rows.Next() {
		item, err := scanInventoryItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func queryInventoryItem(ctx context.Context, db database.DBTX, query string, args ...any) (*model.InventoryItem, error) {
	row := db.QueryRow(ctx, query, args...)
	item, err := scanInventoryItem(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrInventoryItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func scanInventoryItem(row pgx.Row) (*model.InventoryItem, error) {
	var item model.InventoryItem

	err := row.Scan(
		&item.SKUID,
		&item.ProductID,
		&item.ProductName,
		&item.CategoryID,
		&item.CategoryName,
		&item.Specs,
		&item.PriceCents,
		&item.TotalStock,
		&item.FrozenStock,
		&item.Version,
		&item.IsEnabled,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
