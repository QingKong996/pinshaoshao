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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	const query = `SELECT id, username, password_hash, email, status, created_at, updated_at FROM users WHERE username = $1`

	var user model.User

	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(`repository.GetUserByUsername: %w`, apperr.ErrUserNotFound)
		}
		return nil, fmt.Errorf("repository.GetUserByUsername: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) CreateSession(ctx context.Context, user *model.User, tokenHash []byte) error {
	const query = `INSERT INTO sessions (token_hash, user_id) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, tokenHash, user.ID)
	if err != nil {
		return fmt.Errorf("repository.CreateSession: %w", err)
	}
	return nil
}
