package repository

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const productColumns = `id, shop_id, name, description, status, created_at, updated_at`

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	const query = `INSERT INTO products (shop_id, name, description) VALUES ($1, $2, $3) RETURNING id`

	err := r.pool.QueryRow(ctx, query, product.ShopID, product.Name, product.Description).Scan(&product.ID)
	if err != nil {
		return fmt.Errorf("repository.CreateProduct: %w", err)
	}
	return nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id int) (*model.Product, error) {
	const query = `SELECT ` + productColumns + ` FROM products WHERE id = $1`

	product, err := scanProduct(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.GetProductByID: %w", apperr.ErrProductNotFound)
		}
		return nil, fmt.Errorf("repository.GetProductByID: %w", err)
	}
	return product, nil
}

func (r *ProductRepository) ListProductsByShopID(ctx context.Context, shopID int) ([]model.Product, error) {
	const query = `SELECT ` + productColumns + ` FROM products WHERE shop_id = $1 ORDER BY id`

	rows, err := r.pool.Query(ctx, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListProductsByShopID: %w", err)
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListProductsByShopID: %w", err)
		}
		products = append(products, *product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListProductsByShopID: %w", err)
	}
	return products, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *model.Product) error {
	const query = `UPDATE products SET name = $1, description = $2, status = $3, updated_at = now() WHERE id = $4`

	tag, err := r.pool.Exec(ctx, query, product.Name, product.Description, product.Status, product.ID)
	if err != nil {
		return fmt.Errorf("repository.UpdateProduct: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository.UpdateProduct: %w", apperr.ErrProductNotFound)
	}
	return nil
}

func scanProduct(row pgx.Row) (*model.Product, error) {
	var product model.Product

	err := row.Scan(
		&product.ID,
		&product.ShopID,
		&product.Name,
		&product.Description,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}
