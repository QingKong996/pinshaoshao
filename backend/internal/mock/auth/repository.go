package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	ID           int64
	Username     string
	PasswordHash string
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) UpsertAccount(
	ctx context.Context,
	username string,
	passwordHash string,
) (int64, error) {
	const query = `INSERT INTO mock_auth_accounts (username, password_hash)
VALUES ($1, $2)
ON CONFLICT (username) DO UPDATE
SET password_hash = EXCLUDED.password_hash, updated_at = now()
RETURNING id`

	var id int64
	if err := r.pool.QueryRow(ctx, query, username, passwordHash).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) GetAccountByUsername(
	ctx context.Context,
	username string,
) (*Account, error) {
	const query = `SELECT id, username, password_hash
FROM mock_auth_accounts
WHERE username = $1`

	var account Account
	if err := r.pool.QueryRow(ctx, query, username).Scan(
		&account.ID,
		&account.Username,
		&account.PasswordHash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *Repository) CreateSession(
	ctx context.Context,
	accountID int64,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	const query = `INSERT INTO mock_auth_sessions (account_id, token_hash, expires_at)
VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, accountID, tokenHash, expiresAt)
	return err
}

func (r *Repository) GetAccountBySessionTokenHash(
	ctx context.Context,
	tokenHash []byte,
) (*Account, error) {
	const query = `SELECT a.id, a.username, a.password_hash
FROM mock_auth_sessions s
JOIN mock_auth_accounts a ON a.id = s.account_id
WHERE s.token_hash = $1 AND s.expires_at > now()`

	var account Account
	if err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&account.ID,
		&account.Username,
		&account.PasswordHash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}
	return &account, nil
}

func (r *Repository) DeleteSessionByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) error {
	const query = `DELETE FROM mock_auth_sessions WHERE token_hash = $1`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	return err
}

func (r *Repository) DeleteSessionsByAccountID(
	ctx context.Context,
	accountID int64,
) error {
	_, err := r.pool.Exec(
		ctx,
		`DELETE FROM mock_auth_sessions WHERE account_id = $1`,
		accountID,
	)
	return err
}

func (r *Repository) DeleteOtherAccounts(
	ctx context.Context,
	username string,
) error {
	_, err := r.pool.Exec(
		ctx,
		`DELETE FROM mock_auth_accounts WHERE username <> $1`,
		username,
	)
	return err
}

func (r *Repository) DeleteExpiredSessions(ctx context.Context) error {
	_, err := r.pool.Exec(
		ctx,
		`DELETE FROM mock_auth_sessions WHERE expires_at <= now()`,
	)
	return err
}
