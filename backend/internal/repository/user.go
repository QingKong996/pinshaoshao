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

const uniqueViolationCode = "23505"

const userColumns = `id, username, password_hash, COALESCE(email, ''), status, created_at, updated_at`

const aliasedUserColumns = `u.id, u.username, u.password_hash, COALESCE(u.email, ''), u.status, u.created_at, u.updated_at`

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	const query = `SELECT ` + userColumns + ` FROM users WHERE username = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(`repository.GetUserByUsername: %w`, apperr.ErrUserNotFound)
		}
		return nil, fmt.Errorf("repository.GetUserByUsername: %w", err)
	}
	return user, nil
}

// GetUserBySessionTokenHash returns the owner of a non-expired session.
func (r *UserRepository) GetUserBySessionTokenHash(ctx context.Context, tokenHash []byte) (*model.User, error) {
	const query = `SELECT ` + aliasedUserColumns + `
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token_hash = $1 AND s.expires_at > now()`

	user, err := scanUser(r.pool.QueryRow(ctx, query, hex.EncodeToString(tokenHash)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("repository.GetUserBySessionTokenHash: %w", apperr.ErrInvalidSession)
		}
		return nil, fmt.Errorf("repository.GetUserBySessionTokenHash: %w", err)
	}
	return user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	const query = `INSERT INTO users (username, password_hash, email) VALUES ($1, $2, NULLIF($3, '')) RETURNING id`

	err := r.pool.QueryRow(ctx, query, user.Username, user.PasswordHash, user.Email).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			switch pgErr.ConstraintName {
			case "users_username_key":
				return fmt.Errorf("repository.CreateUser: %w", apperr.ErrUsernameTaken)
			case "users_email_key":
				return fmt.Errorf("repository.CreateUser: %w", apperr.ErrEmailTaken)
			}
		}
		return fmt.Errorf("repository.CreateUser: %w", err)
	}
	return nil
}

func (r *UserRepository) CreateSession(ctx context.Context, user *model.User, tokenHash []byte) error {
	const query = `INSERT INTO sessions (token_hash, user_id) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, hex.EncodeToString(tokenHash), user.ID)
	if err != nil {
		return fmt.Errorf("repository.CreateSession: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash []byte) error {
	const query = `DELETE FROM sessions WHERE token_hash = $1`

	_, err := r.pool.Exec(ctx, query, hex.EncodeToString(tokenHash))
	if err != nil {
		return fmt.Errorf("repository.DeleteSessionByTokenHash: %w", err)
	}
	return nil
}

func scanUser(row pgx.Row) (*model.User, error) {
	var user model.User

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
