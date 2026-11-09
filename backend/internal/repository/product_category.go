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

const productCategorySelect = `SELECT pc.id, pc.shop_id, pc.name, pc.created_at, pc.updated_at, COUNT(p.id)::BIGINT
FROM product_categories pc
LEFT JOIN products p ON p.category_id = pc.id`

type ProductCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewProductCategoryRepository(pool *pgxpool.Pool) *ProductCategoryRepository {
	return &ProductCategoryRepository{pool: pool}
}

// ListProductCategoriesByShopID returns the categories of the shop ordered by name.
func (r *ProductCategoryRepository) ListProductCategoriesByShopID(ctx context.Context, shopID int) ([]model.ProductCategory, error) {
	query := productCategorySelect + ` WHERE pc.shop_id = $1 GROUP BY pc.id ORDER BY pc.name`

	rows, err := r.pool.Query(ctx, query, shopID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListProductCategoriesByShopID: %w", err)
	}
	defer rows.Close()

	categories := make([]model.ProductCategory, 0)
	for rows.Next() {
		category, err := scanProductCategoryWithCount(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListProductCategoriesByShopID: %w", err)
		}
		categories = append(categories, *category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListProductCategoriesByShopID: %w", err)
	}
	return categories, nil
}

// GetProductCategoryByID returns a category of the shop with its product count.
func (r *ProductCategoryRepository) GetProductCategoryByID(ctx context.Context, shopID int, categoryID int) (*model.ProductCategory, error) {
	query := productCategorySelect + ` WHERE pc.shop_id = $1 AND pc.id = $2 GROUP BY pc.id`

	category, err := scanProductCategoryWithCount(r.pool.QueryRow(ctx, query, shopID, categoryID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.GetProductCategoryByID: %w", apperr.ErrCategoryNotFound)
		}
		return nil, fmt.Errorf("repository.GetProductCategoryByID: %w", err)
	}
	return category, nil
}

func (r *ProductCategoryRepository) CreateProductCategory(ctx context.Context, category *model.ProductCategory) error {
	const query = `INSERT INTO product_categories (shop_id, name) VALUES ($1, $2) RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query, category.ShopID, category.Name).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == "product_categories_shop_id_name_key" {
			return apperr.ErrCategoryNameTaken
		}
		return fmt.Errorf("repository.CreateProductCategory: %w", err)
	}
	return nil
}

func (r *ProductCategoryRepository) UpdateProductCategory(ctx context.Context, category *model.ProductCategory) error {
	const query = `UPDATE product_categories SET name = $1, updated_at = now() WHERE id = $2 AND shop_id = $3`

	tag, err := r.pool.Exec(ctx, query, category.Name, category.ID, category.ShopID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == "product_categories_shop_id_name_key" {
			return apperr.ErrCategoryNameTaken
		}
		return fmt.Errorf("repository.UpdateProductCategory: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository.UpdateProductCategory: %w", apperr.ErrCategoryNotFound)
	}
	return nil
}

// DeleteProductCategory deletes a category of the shop; its products become uncategorized.
func (r *ProductCategoryRepository) DeleteProductCategory(ctx context.Context, shopID int, categoryID int) error {
	const query = `DELETE FROM product_categories WHERE id = $1 AND shop_id = $2`

	tag, err := r.pool.Exec(ctx, query, categoryID, shopID)
	if err != nil {
		return fmt.Errorf("repository.DeleteProductCategory: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository.DeleteProductCategory: %w", apperr.ErrCategoryNotFound)
	}
	return nil
}

func scanProductCategoryWithCount(row pgx.Row) (*model.ProductCategory, error) {
	var category model.ProductCategory

	err := row.Scan(
		&category.ID,
		&category.ShopID,
		&category.Name,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.ProductCount,
	)
	if err != nil {
		return nil, err
	}
	return &category, nil
}
