package repository

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const productColumns = `id, shop_id, name, description, status, category_id, created_at, updated_at`

const productSummarySelect = `SELECT p.id, p.shop_id, p.name, p.description, p.status, p.category_id,
       p.created_at, p.updated_at, pc.name,
       COUNT(ps.id)::BIGINT,
       COALESCE(SUM(ps.stock), 0)::BIGINT,
       MIN(ps.price_cents)::BIGINT,
       MAX(ps.price_cents)::BIGINT
FROM products p
LEFT JOIN product_categories pc ON pc.id = p.category_id
LEFT JOIN product_skus ps ON ps.product_id = p.id`

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, product *model.Product) error {
	const query = `INSERT INTO products (shop_id, name, description, status, category_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.pool.QueryRow(ctx, query, product.ShopID, product.Name, product.Description, product.Status, product.CategoryID).Scan(&product.ID)
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

// GetProductSummaryByID returns a product of the shop with its category and SKU aggregates.
func (r *ProductRepository) GetProductSummaryByID(ctx context.Context, shopID int, productID int) (*model.ProductSummary, error) {
	query := productSummarySelect + ` WHERE p.shop_id = $1 AND p.id = $2 GROUP BY p.id, pc.name`

	summary, err := scanProductSummary(r.pool.QueryRow(ctx, query, shopID, productID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.GetProductSummaryByID: %w", apperr.ErrProductNotFound)
		}
		return nil, fmt.Errorf("repository.GetProductSummaryByID: %w", err)
	}
	return summary, nil
}

// ListProductSummaries returns a filtered and sorted page of products with their aggregates.
func (r *ProductRepository) ListProductSummaries(ctx context.Context, shopID int, filter model.ProductListFilter) (*model.ProductListResult, error) {
	where, args := productSummaryFilter(shopID, filter)

	var total int64
	countQuery := `SELECT COUNT(*)::BIGINT FROM products p ` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("repository.ListProductSummaries: count: %w", err)
	}

	orderBy, err := productSummaryOrderBy(filter.Sort, filter.Order)
	if err != nil {
		return nil, fmt.Errorf("repository.ListProductSummaries: %w", err)
	}

	limitIndex := len(args) + 1
	offsetIndex := len(args) + 2
	query := fmt.Sprintf(
		"%s %s GROUP BY p.id, pc.name ORDER BY %s LIMIT $%d OFFSET $%d",
		productSummarySelect,
		where,
		orderBy,
		limitIndex,
		offsetIndex,
	)
	listArgs := append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, fmt.Errorf("repository.ListProductSummaries: %w", err)
	}
	defer rows.Close()

	summaries := make([]model.ProductSummary, 0)
	for rows.Next() {
		summary, err := scanProductSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListProductSummaries: %w", err)
		}
		summaries = append(summaries, *summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListProductSummaries: %w", err)
	}
	return &model.ProductListResult{Products: summaries, Total: total}, nil
}

func productSummaryFilter(shopID int, filter model.ProductListFilter) (string, []any) {
	conditions := []string{"p.shop_id = $1"}
	args := []any{shopID}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		conditions = append(conditions, fmt.Sprintf(
			"(p.name ILIKE $%d OR p.description ILIKE $%d)",
			len(args),
			len(args),
		))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		conditions = append(conditions, fmt.Sprintf("p.status = $%d", len(args)))
	}
	if filter.CategoryID != nil {
		args = append(args, *filter.CategoryID)
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", len(args)))
	}
	if filter.Uncategorized {
		conditions = append(conditions, "p.category_id IS NULL")
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func productSummaryOrderBy(sort model.ProductSort, order model.ProductSortOrder) (string, error) {
	var expression string
	switch sort {
	case model.ProductSortCreatedAt:
		expression = "p.created_at"
	case model.ProductSortUpdatedAt:
		expression = "p.updated_at"
	case model.ProductSortName:
		expression = "LOWER(p.name)"
	case model.ProductSortCategory:
		expression = "LOWER(pc.name)"
	case model.ProductSortStatus:
		expression = "p.status"
	case model.ProductSortSKUCount:
		expression = "COUNT(ps.id)"
	case model.ProductSortTotalStock:
		expression = "COALESCE(SUM(ps.stock), 0)"
	case model.ProductSortMinPrice:
		expression = "MIN(ps.price_cents)"
	default:
		return "", apperr.ErrInvalidProductListQuery
	}

	var direction string
	switch order {
	case model.ProductSortOrderAsc:
		direction = "ASC"
	case model.ProductSortOrderDesc:
		direction = "DESC"
	default:
		return "", apperr.ErrInvalidProductListQuery
	}

	return fmt.Sprintf("%s %s NULLS LAST, p.id %s", expression, direction, direction), nil
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
	const query = `UPDATE products SET name = $1, description = $2, status = $3, category_id = $4, updated_at = now() WHERE id = $5 AND shop_id = $6`

	tag, err := r.pool.Exec(ctx, query, product.Name, product.Description, product.Status, product.CategoryID, product.ID, product.ShopID)
	if err != nil {
		return fmt.Errorf("repository.UpdateProduct: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository.UpdateProduct: %w", apperr.ErrProductNotFound)
	}
	return nil
}

// DeleteProduct deletes a product of the shop; its SKUs are removed by the schema cascade.
func (r *ProductRepository) DeleteProduct(ctx context.Context, shopID int, productID int) error {
	const query = `DELETE FROM products WHERE id = $1 AND shop_id = $2`

	tag, err := r.pool.Exec(ctx, query, productID, shopID)
	if err != nil {
		return fmt.Errorf("repository.DeleteProduct: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository.DeleteProduct: %w", apperr.ErrProductNotFound)
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
		&product.CategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func scanProductSummary(row pgx.Row) (*model.ProductSummary, error) {
	var summary model.ProductSummary

	err := row.Scan(
		&summary.ID,
		&summary.ShopID,
		&summary.Name,
		&summary.Description,
		&summary.Status,
		&summary.CategoryID,
		&summary.CreatedAt,
		&summary.UpdatedAt,
		&summary.CategoryName,
		&summary.SKUCount,
		&summary.TotalStock,
		&summary.MinPriceCents,
		&summary.MaxPriceCents,
	)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}
