package repository

import (
	"context"
	"fmt"

	"pinshaoshao/internal/feature/feed/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const feedProductSelect = `SELECT p.id, p.shop_id, s.name, p.name, p.description, pc.name,
       MIN(ps.price_cents)::BIGINT, MAX(ps.price_cents)::BIGINT,
       SUM(ii.total_stock - ii.frozen_stock)::BIGINT
FROM products p
JOIN shops s ON s.id = p.shop_id
JOIN product_skus ps ON ps.product_id = p.id AND ps.is_enabled
JOIN inventory_items ii ON ii.sku_id = ps.id
LEFT JOIN product_categories pc ON pc.id = p.category_id
WHERE p.status = 'on_shelf'
  AND ii.total_stock - ii.frozen_stock > 0
GROUP BY p.id, s.name, pc.name`

const feedProductCount = `SELECT COUNT(*)::BIGINT
FROM products p
WHERE p.status = 'on_shelf'
  AND EXISTS (
      SELECT 1
      FROM product_skus ps
      JOIN inventory_items ii ON ii.sku_id = ps.id
      WHERE ps.product_id = p.id
        AND ps.is_enabled
        AND ii.total_stock - ii.frozen_stock > 0
  )`

type FeedRepository struct {
	pool *pgxpool.Pool
}

func NewFeedRepository(pool *pgxpool.Pool) *FeedRepository {
	return &FeedRepository{pool: pool}
}

func (r *FeedRepository) List(ctx context.Context, userID int, page int, pageSize int) (*model.FeedResult, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, feedProductCount).Scan(&total); err != nil {
		return nil, fmt.Errorf("repository.ListFeed: count: %w", err)
	}

	items := make([]model.FeedItem, 0)
	if total > 0 {
		// The buyer ID seeds a stable product order across paginated requests.
		query := feedProductSelect + `
ORDER BY hashtextextended(p.id::text, $1::BIGINT), p.id
LIMIT $2 OFFSET $3`
		rows, err := r.pool.Query(ctx, query, userID, pageSize, (page-1)*pageSize)
		if err != nil {
			return nil, fmt.Errorf("repository.ListFeed: query: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			item, err := scanFeedItem(rows)
			if err != nil {
				return nil, fmt.Errorf("repository.ListFeed: scan: %w", err)
			}
			items = append(items, *item)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("repository.ListFeed: rows: %w", err)
		}
	}

	return &model.FeedResult{
		Items:    items,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func scanFeedItem(row pgx.Row) (*model.FeedItem, error) {
	var item model.FeedItem

	err := row.Scan(
		&item.ID,
		&item.ShopID,
		&item.ShopName,
		&item.Name,
		&item.Description,
		&item.CategoryName,
		&item.MinPriceCents,
		&item.MaxPriceCents,
		&item.AvailableStock,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
