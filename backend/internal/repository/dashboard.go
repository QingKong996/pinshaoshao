package repository

import (
	"context"
	"fmt"
	"pinshaoshao/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

const dashboardStatsQuery = `SELECT
    (SELECT COUNT(*) FROM products WHERE shop_id = $1),
    (SELECT COUNT(*) FROM products WHERE shop_id = $1 AND status = 'on_shelf'),
    (SELECT COUNT(*) FROM product_skus ps JOIN products p ON p.id = ps.product_id WHERE p.shop_id = $1),
    (SELECT COUNT(*) FROM product_categories WHERE shop_id = $1),
    (SELECT COALESCE(SUM(ps.stock), 0)::BIGINT FROM product_skus ps JOIN products p ON p.id = ps.product_id WHERE p.shop_id = $1),
    (SELECT COALESCE(SUM(ps.price_cents * ps.stock), 0)::BIGINT FROM product_skus ps JOIN products p ON p.id = ps.product_id WHERE p.shop_id = $1),
    (SELECT COUNT(*) FROM product_skus ps JOIN products p ON p.id = ps.product_id WHERE p.shop_id = $1 AND ps.is_enabled AND ps.stock = 0)`

type DashboardRepository struct {
	pool *pgxpool.Pool
}

func NewDashboardRepository(pool *pgxpool.Pool) *DashboardRepository {
	return &DashboardRepository{pool: pool}
}

// GetDashboardStats returns the shop counters.
func (r *DashboardRepository) GetDashboardStats(ctx context.Context, shopID int) (*model.DashboardStats, error) {
	var stats model.DashboardStats

	err := r.pool.QueryRow(ctx, dashboardStatsQuery, shopID).Scan(
		&stats.ProductCount,
		&stats.OnShelfProductCount,
		&stats.SKUCount,
		&stats.CategoryCount,
		&stats.TotalStock,
		&stats.InventoryValueCents,
		&stats.OutOfStockSKUCount,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.GetDashboardStats: %w", err)
	}
	return &stats, nil
}
