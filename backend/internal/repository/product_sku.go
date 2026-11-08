package repository

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const foreignKeyViolationCode = "23503"

const productSKUColumns = `id, product_id, specs, price_cents, stock, is_enabled, created_at, updated_at`

type ProductSKURepository struct {
	pool *pgxpool.Pool
}

func NewProductSKURepository(pool *pgxpool.Pool) *ProductSKURepository {
	return &ProductSKURepository{pool: pool}
}

func (r *ProductSKURepository) CreateProductSKU(ctx context.Context, sku *model.ProductSKU) error {
	const query = `INSERT INTO product_skus (product_id, specs, price_cents, stock, is_enabled) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, sku.ProductID, sku.Specs, sku.PriceCents, sku.Stock, sku.IsEnabled).Scan(
		&sku.ID,
		&sku.CreatedAt,
		&sku.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == "product_skus_product_id_specs_key":
				return fmt.Errorf("repository.CreateProductSKU: %w", apperr.ErrProductSKUAlreadyExists)
			case pgErr.Code == foreignKeyViolationCode:
				return fmt.Errorf("repository.CreateProductSKU: %w", apperr.ErrProductNotFound)
			}
		}
		return fmt.Errorf("repository.CreateProductSKU: %w", err)
	}
	return nil
}

func (r *ProductSKURepository) ListProductSKUsByProductID(ctx context.Context, productID int) ([]model.ProductSKU, error) {
	const query = `SELECT ` + productSKUColumns + ` FROM product_skus WHERE product_id = $1 ORDER BY id`

	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListProductSKUsByProductID: %w", err)
	}
	defer rows.Close()

	skus := make([]model.ProductSKU, 0)
	for rows.Next() {
		sku, err := scanProductSKU(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListProductSKUsByProductID: %w", err)
		}
		skus = append(skus, *sku)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListProductSKUsByProductID: %w", err)
	}
	return skus, nil
}

func scanProductSKU(row pgx.Row) (*model.ProductSKU, error) {
	var sku model.ProductSKU

	err := row.Scan(
		&sku.ID,
		&sku.ProductID,
		&sku.Specs,
		&sku.PriceCents,
		&sku.Stock,
		&sku.IsEnabled,
		&sku.CreatedAt,
		&sku.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sku, nil
}
