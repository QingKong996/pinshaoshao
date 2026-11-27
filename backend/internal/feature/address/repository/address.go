package repository

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/feature/address/apperr"
	"pinshaoshao/internal/feature/address/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const addressColumns = `id, user_id, receiver_name, receiver_phone, shipping_address,
is_default, created_at, updated_at`

type AddressRepository struct {
	pool *pgxpool.Pool
}

func NewAddressRepository(pool *pgxpool.Pool) *AddressRepository {
	return &AddressRepository{pool: pool}
}

func (r *AddressRepository) ListByUserID(ctx context.Context, userID int) ([]model.Address, error) {
	const query = `SELECT ` + addressColumns + `
FROM user_addresses
WHERE user_id = $1
ORDER BY is_default DESC, updated_at DESC, id DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository.ListByUserID: %w", err)
	}
	defer rows.Close()

	addresses := make([]model.Address, 0)
	for rows.Next() {
		address, err := scanAddress(rows)
		if err != nil {
			return nil, fmt.Errorf("repository.ListByUserID: %w", err)
		}
		addresses = append(addresses, *address)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.ListByUserID: %w", err)
	}
	return addresses, nil
}

func (r *AddressRepository) Create(ctx context.Context, userID int, address model.Address) (*model.Address, error) {
	var created *model.Address

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if err := lockUserAddresses(ctx, tx, userID); err != nil {
			return err
		}

		var hasAddress bool
		if err := tx.QueryRow(
			ctx,
			`SELECT EXISTS (SELECT 1 FROM user_addresses WHERE user_id = $1)`,
			userID,
		).Scan(&hasAddress); err != nil {
			return err
		}

		isDefault := address.IsDefault || !hasAddress
		if isDefault {
			if _, err := tx.Exec(
				ctx,
				`UPDATE user_addresses
SET is_default = false, updated_at = now()
WHERE user_id = $1 AND is_default`,
				userID,
			); err != nil {
				return err
			}
		}

		const query = `INSERT INTO user_addresses (
    user_id, receiver_name, receiver_phone, shipping_address, is_default
) VALUES ($1, $2, $3, $4, $5)
RETURNING ` + addressColumns

		var err error
		created, err = scanAddress(tx.QueryRow(
			ctx,
			query,
			userID,
			address.ReceiverName,
			address.ReceiverPhone,
			address.ShippingAddress,
			isDefault,
		))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("repository.Create: %w", err)
	}
	return created, nil
}

func (r *AddressRepository) Update(ctx context.Context, userID int, addressID int, address model.Address) (*model.Address, error) {
	var updated *model.Address

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if err := lockUserAddresses(ctx, tx, userID); err != nil {
			return err
		}

		var exists bool
		if err := tx.QueryRow(
			ctx,
			`SELECT EXISTS (
    SELECT 1 FROM user_addresses WHERE id = $1 AND user_id = $2
)`,
			addressID,
			userID,
		).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return apperr.ErrAddressNotFound
		}

		if address.IsDefault {
			if _, err := tx.Exec(
				ctx,
				`UPDATE user_addresses
SET is_default = false, updated_at = now()
WHERE user_id = $1 AND id <> $2 AND is_default`,
				userID,
				addressID,
			); err != nil {
				return err
			}
		}

		const query = `UPDATE user_addresses
SET receiver_name = $3,
    receiver_phone = $4,
    shipping_address = $5,
    is_default = $6,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING ` + addressColumns

		var err error
		updated, err = scanAddress(tx.QueryRow(
			ctx,
			query,
			addressID,
			userID,
			address.ReceiverName,
			address.ReceiverPhone,
			address.ShippingAddress,
			address.IsDefault,
		))
		return err
	})
	if err != nil {
		if errors.Is(err, apperr.ErrAddressNotFound) {
			return nil, fmt.Errorf("repository.Update: %w", apperr.ErrAddressNotFound)
		}
		return nil, fmt.Errorf("repository.Update: %w", err)
	}
	return updated, nil
}

func (r *AddressRepository) Delete(ctx context.Context, userID int, addressID int) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if err := lockUserAddresses(ctx, tx, userID); err != nil {
			return err
		}

		var wasDefault bool
		err := tx.QueryRow(
			ctx,
			`SELECT is_default
FROM user_addresses
WHERE id = $1 AND user_id = $2
FOR UPDATE`,
			addressID,
			userID,
		).Scan(&wasDefault)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.ErrAddressNotFound
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			`DELETE FROM user_addresses WHERE id = $1 AND user_id = $2`,
			addressID,
			userID,
		); err != nil {
			return err
		}

		if !wasDefault {
			return nil
		}

		_, err = tx.Exec(
			ctx,
			`UPDATE user_addresses
SET is_default = true, updated_at = now()
WHERE id = (
    SELECT id
    FROM user_addresses
    WHERE user_id = $1
    ORDER BY created_at DESC, id DESC
    LIMIT 1
)`,
			userID,
		)
		return err
	})
	if err != nil {
		if errors.Is(err, apperr.ErrAddressNotFound) {
			return fmt.Errorf("repository.Delete: %w", apperr.ErrAddressNotFound)
		}
		return fmt.Errorf("repository.Delete: %w", err)
	}
	return nil
}

func (r *AddressRepository) SetDefault(ctx context.Context, userID int, addressID int) (*model.Address, error) {
	var address *model.Address

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if err := lockUserAddresses(ctx, tx, userID); err != nil {
			return err
		}

		var exists bool
		if err := tx.QueryRow(
			ctx,
			`SELECT EXISTS (
    SELECT 1 FROM user_addresses WHERE id = $1 AND user_id = $2
)`,
			addressID,
			userID,
		).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return apperr.ErrAddressNotFound
		}

		if _, err := tx.Exec(
			ctx,
			`UPDATE user_addresses
SET is_default = false, updated_at = now()
WHERE user_id = $1 AND is_default AND id <> $2`,
			userID,
			addressID,
		); err != nil {
			return err
		}

		const query = `UPDATE user_addresses
SET is_default = true, updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING ` + addressColumns

		var err error
		address, err = scanAddress(tx.QueryRow(ctx, query, addressID, userID))
		return err
	})
	if err != nil {
		if errors.Is(err, apperr.ErrAddressNotFound) {
			return nil, fmt.Errorf("repository.SetDefault: %w", apperr.ErrAddressNotFound)
		}
		return nil, fmt.Errorf("repository.SetDefault: %w", err)
	}
	return address, nil
}

func lockUserAddresses(ctx context.Context, tx pgx.Tx, userID int) error {
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, userID)
	return err
}

func scanAddress(row pgx.Row) (*model.Address, error) {
	var address model.Address

	err := row.Scan(
		&address.ID,
		&address.UserID,
		&address.ReceiverName,
		&address.ReceiverPhone,
		&address.ShippingAddress,
		&address.IsDefault,
		&address.CreatedAt,
		&address.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &address, nil
}
