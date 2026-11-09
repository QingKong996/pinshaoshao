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

// CreateProductSKU inserts a SKU while holding the product row lock, so another
// concurrent SKU creation cannot change the spec key set in between.
func (r *ProductSKURepository) CreateProductSKU(ctx context.Context, sku *model.ProductSKU) error {
	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		if err := lockProduct(ctx, tx, sku.ProductID); err != nil {
			return err
		}

		skus, err := listProductSKUs(ctx, tx, sku.ProductID)
		if err != nil {
			return fmt.Errorf("list product skus: %w", err)
		}
		for i := range skus {
			if !hasSameSpecKeys(skus[i].Specs, sku.Specs) {
				return apperr.ErrProductSKUSpecsMismatch
			}
		}

		return insertProductSKU(ctx, tx, sku)
	})
	if err != nil {
		return fmt.Errorf("repository.CreateProductSKU: %w", err)
	}
	return nil
}

// UpdateProductSKU applies a partial update while holding the product row lock, so
// the spec key set cannot change between validation and write. Only the provided
// fields are written, so concurrent updates of other fields are not overwritten.
// Updating the only SKU of a product is allowed to change the spec keys.
func (r *ProductSKURepository) UpdateProductSKU(ctx context.Context, productID int, skuID int, update *model.ProductSKUUpdate) (*model.ProductSKU, error) {
	var sku *model.ProductSKU

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		if err := lockProduct(ctx, tx, productID); err != nil {
			return err
		}

		skus, err := listProductSKUs(ctx, tx, productID)
		if err != nil {
			return fmt.Errorf("list product skus: %w", err)
		}

		found := false
		for i := range skus {
			if skus[i].ID == skuID {
				found = true
				continue
			}
			if update.Specs != nil && !hasSameSpecKeys(skus[i].Specs, update.Specs) {
				return apperr.ErrProductSKUSpecsMismatch
			}
		}
		if !found {
			return apperr.ErrProductSKUNotFound
		}

		sku, err = updateProductSKU(ctx, tx, productID, skuID, update)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("repository.UpdateProductSKU: %w", err)
	}
	return sku, nil
}

// DeleteProductSKU deletes a SKU of the product.
func (r *ProductSKURepository) DeleteProductSKU(ctx context.Context, productID int, skuID int) error {
	const query = `DELETE FROM product_skus WHERE product_id = $1 AND id = $2`

	tag, err := r.pool.Exec(ctx, query, productID, skuID)
	if err != nil {
		return fmt.Errorf("repository.DeleteProductSKU: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository.DeleteProductSKU: %w", apperr.ErrProductSKUNotFound)
	}
	return nil
}

func lockProduct(ctx context.Context, db dbtx, productID int) error {
	const query = `SELECT id FROM products WHERE id = $1 FOR UPDATE`

	var id int
	if err := db.QueryRow(ctx, query, productID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.ErrProductNotFound
		}
		return fmt.Errorf("lock product: %w", err)
	}
	return nil
}

func (r *ProductSKURepository) ListProductSKUsByProductID(ctx context.Context, productID int) ([]model.ProductSKU, error) {
	skus, err := listProductSKUs(ctx, r.pool, productID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListProductSKUsByProductID: %w", err)
	}
	return skus, nil
}

func listProductSKUs(ctx context.Context, db dbtx, productID int) ([]model.ProductSKU, error) {
	const query = `SELECT ` + productSKUColumns + ` FROM product_skus WHERE product_id = $1 ORDER BY id`

	rows, err := db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skus := make([]model.ProductSKU, 0)
	for rows.Next() {
		sku, err := scanProductSKU(rows)
		if err != nil {
			return nil, err
		}
		skus = append(skus, *sku)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return skus, nil
}

func insertProductSKU(ctx context.Context, db dbtx, sku *model.ProductSKU) error {
	const query = `INSERT INTO product_skus (product_id, specs, price_cents, stock, is_enabled) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`

	err := db.QueryRow(ctx, query, sku.ProductID, sku.Specs, sku.PriceCents, sku.Stock, sku.IsEnabled).Scan(
		&sku.ID,
		&sku.CreatedAt,
		&sku.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == "product_skus_product_id_specs_key":
				return apperr.ErrProductSKUAlreadyExists
			case pgErr.Code == foreignKeyViolationCode:
				return apperr.ErrProductNotFound
			}
		}
		return err
	}
	return nil
}

func updateProductSKU(ctx context.Context, db dbtx, productID int, skuID int, update *model.ProductSKUUpdate) (*model.ProductSKU, error) {
	const query = `UPDATE product_skus SET
    specs = COALESCE($1::jsonb, specs),
    price_cents = COALESCE($2, price_cents),
    stock = COALESCE($3, stock),
    is_enabled = COALESCE($4, is_enabled),
    updated_at = now()
WHERE id = $5 AND product_id = $6
RETURNING ` + productSKUColumns

	var specs any
	if update.Specs != nil {
		specs = update.Specs
	}

	sku, err := scanProductSKU(db.QueryRow(ctx, query, specs, update.PriceCents, update.Stock, update.IsEnabled, skuID, productID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrProductSKUNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == "product_skus_product_id_specs_key" {
			return nil, apperr.ErrProductSKUAlreadyExists
		}
		return nil, err
	}
	return sku, nil
}

func hasSameSpecKeys(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for name := range a {
		if _, ok := b[name]; !ok {
			return false
		}
	}
	return true
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
