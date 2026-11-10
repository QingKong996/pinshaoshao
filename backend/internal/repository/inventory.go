package repository

import (
	"context"
	"fmt"
	"pinshaoshao/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const inventorySelect = `SELECT ps.id, ps.product_id, p.name, p.category_id, pc.name, ps.specs, ps.price_cents, ps.stock, ps.is_enabled, ps.updated_at
FROM product_skus ps
JOIN products p ON p.id = ps.product_id
LEFT JOIN product_categories pc ON pc.id = p.category_id`

type InventoryRepository struct {
	pool *pgxpool.Pool
}

func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

// ListInventoryByShopID returns every SKU of the shop ordered by stock.
func (r *InventoryRepository) ListInventoryByShopID(ctx context.Context, shopID int) ([]model.InventoryItem, error) {
	query := inventorySelect + ` WHERE p.shop_id = $1 ORDER BY ps.stock ASC, p.id, ps.id`

	items, err := queryInventory(ctx, r.pool, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListInventoryByShopID: %w", err)
	}
	return items, nil
}

func queryInventory(ctx context.Context, db dbtx, query string, args ...any) ([]model.InventoryItem, error) {
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
		&item.Stock,
		&item.IsEnabled,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
