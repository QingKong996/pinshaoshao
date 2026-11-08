package repository

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const shopColumns = `id, username, password_hash, email, status, created_at, updated_at`

const aliasedShopColumns = `sh.id, sh.username, sh.password_hash, sh.email, sh.status, sh.created_at, sh.updated_at`

type ShopRepository struct {
	pool *pgxpool.Pool
}

func NewShopRepository(pool *pgxpool.Pool) *ShopRepository {
	return &ShopRepository{pool: pool}
}

func (r *ShopRepository) GetShopByUsername(ctx context.Context, username string) (*model.Shop, error) {
	const query = `SELECT ` + shopColumns + ` FROM shops WHERE username = $1`

	shop, err := scanShop(r.pool.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(`repository.GetShopByUsername: %w`, apperr.ErrShopNotFound)
		}
		return nil, fmt.Errorf("repository.GetShopByUsername: %w", err)
	}
	return shop, nil
}

// GetShopBySessionTokenHash returns the owner of a non-expired shop session.
func (r *ShopRepository) GetShopBySessionTokenHash(ctx context.Context, tokenHash []byte) (*model.Shop, error) {
	const query = `SELECT ` + aliasedShopColumns + `
FROM shops sh
JOIN shop_sessions ss ON ss.shop_id = sh.id
WHERE ss.token_hash = $1 AND ss.expires_at > now()`

	shop, err := scanShop(r.pool.QueryRow(ctx, query, hex.EncodeToString(tokenHash)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.GetShopBySessionTokenHash: %w", apperr.ErrInvalidShopSession)
		}
		return nil, fmt.Errorf("repository.GetShopBySessionTokenHash: %w", err)
	}
	return shop, nil
}

// CreateShopWithSession creates the shop and its first session atomically.
func (r *ShopRepository) CreateShopWithSession(ctx context.Context, shop *model.Shop, tokenHash []byte) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if err := createShop(ctx, tx, shop); err != nil {
			return err
		}
		return createShopSession(ctx, tx, shop, tokenHash)
	})
	if err != nil {
		return fmt.Errorf("repository.CreateShopWithSession: %w", err)
	}
	return nil
}

func (r *ShopRepository) CreateShop(ctx context.Context, shop *model.Shop) error {
	if err := createShop(ctx, r.pool, shop); err != nil {
		return fmt.Errorf("repository.CreateShop: %w", err)
	}
	return nil
}

func (r *ShopRepository) CreateSession(ctx context.Context, shop *model.Shop, tokenHash []byte) error {
	if err := createShopSession(ctx, r.pool, shop, tokenHash); err != nil {
		return fmt.Errorf("repository.CreateSession: %w", err)
	}
	return nil
}

func createShop(ctx context.Context, db dbtx, shop *model.Shop) error {
	const query = `INSERT INTO shops (username, password_hash, email) VALUES ($1, $2, $3) RETURNING id`

	err := db.QueryRow(ctx, query, shop.Username, shop.PasswordHash, shop.Email).Scan(&shop.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			switch pgErr.ConstraintName {
			case "shops_username_key":
				return apperr.ErrShopUsernameTaken
			case "shops_email_key":
				return apperr.ErrShopEmailTaken
			}
		}
		return err
	}
	return nil
}

func createShopSession(ctx context.Context, db dbtx, shop *model.Shop, tokenHash []byte) error {
	const query = `INSERT INTO shop_sessions (token_hash, shop_id) VALUES ($1, $2)`

	_, err := db.Exec(ctx, query, hex.EncodeToString(tokenHash), shop.ID)
	return err
}

func (r *ShopRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash []byte) error {
	const query = `DELETE FROM shop_sessions WHERE token_hash = $1`

	_, err := r.pool.Exec(ctx, query, hex.EncodeToString(tokenHash))
	if err != nil {
		return fmt.Errorf("repository.DeleteSessionByTokenHash: %w", err)
	}
	return nil
}

func scanShop(row pgx.Row) (*model.Shop, error) {
	var shop model.Shop

	err := row.Scan(
		&shop.ID,
		&shop.Username,
		&shop.PasswordHash,
		&shop.Email,
		&shop.Status,
		&shop.CreatedAt,
		&shop.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &shop, nil
}
