package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sessionTokenBytes = 32
	maxUsernameBytes  = 100
	minPasswordBytes  = 12
	maxPasswordBytes  = 1024
)

type sessionRepository interface {
	UpsertAccount(ctx context.Context, username, passwordHash string) (int64, error)
	GetAccountByUsername(ctx context.Context, username string) (*Account, error)
	CreateSession(ctx context.Context, accountID int64, tokenHash []byte, expiresAt time.Time) error
	GetAccountBySessionTokenHash(ctx context.Context, tokenHash []byte) (*Account, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash []byte) error
	DeleteSessionsByAccountID(ctx context.Context, accountID int64) error
	DeleteOtherAccounts(ctx context.Context, username string) error
	DeleteExpiredSessions(ctx context.Context) error
}

type Service struct {
	repository   sessionRepository
	config       Config
	passwordHash string
	dummyHash    string
	now          func() time.Time
	randomRead   func([]byte) (int, error)
}

func NewService(pool *pgxpool.Pool, config Config) (*Service, error) {
	config = config.withDefaults()
	if err := config.validate(); err != nil {
		return nil, err
	}
	return NewServiceWithRepository(NewRepository(pool), config)
}

func NewServiceWithRepository(
	repository sessionRepository,
	config Config,
) (*Service, error) {
	config = config.withDefaults()
	if err := config.validate(); err != nil {
		return nil, err
	}
	passwordHash, err := argon2id.CreateHash(config.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("hash mock auth password: %w", err)
	}
	return &Service{
		repository:   repository,
		config:       config,
		passwordHash: passwordHash,
		dummyHash:    passwordHash,
		now:          time.Now,
		randomRead:   rand.Read,
	}, nil
}

func (s *Service) Bootstrap(ctx context.Context) error {
	rotated := true
	account, err := s.repository.GetAccountByUsername(ctx, s.config.Username)
	if err == nil {
		matches, compareErr := argon2id.ComparePasswordAndHash(
			s.config.Password,
			account.PasswordHash,
		)
		rotated = compareErr != nil || !matches
	} else if !errors.Is(err, errAccountNotFound) {
		return fmt.Errorf("read mock auth account: %w", err)
	}

	var accountID int64
	if rotated {
		accountID, err = s.repository.UpsertAccount(
			ctx,
			s.config.Username,
			s.passwordHash,
		)
		if err != nil {
			return fmt.Errorf("bootstrap mock auth account: %w", err)
		}
	} else {
		accountID = account.ID
	}
	if err := s.repository.DeleteOtherAccounts(ctx, s.config.Username); err != nil {
		return fmt.Errorf("delete obsolete mock auth accounts: %w", err)
	}
	if rotated {
		if err := s.repository.DeleteSessionsByAccountID(ctx, accountID); err != nil {
			return fmt.Errorf("revoke mock auth sessions after credential rotation: %w", err)
		}
	}
	if err := s.repository.DeleteExpiredSessions(ctx); err != nil {
		return fmt.Errorf("delete expired mock auth sessions: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, username, password string) (string, error) {
	account, err := s.repository.GetAccountByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, errAccountNotFound) {
			_, _ = argon2id.ComparePasswordAndHash(password, s.dummyHash)
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("get mock auth account: %w", err)
	}

	matches, err := argon2id.ComparePasswordAndHash(password, account.PasswordHash)
	if err != nil || !matches {
		return "", ErrInvalidCredentials
	}

	token, tokenHash, err := s.generateSessionToken()
	if err != nil {
		return "", fmt.Errorf("generate mock auth session: %w", err)
	}
	expiresAt := s.now().Add(s.config.SessionTTL)
	if err := s.repository.CreateSession(
		ctx,
		account.ID,
		tokenHash,
		expiresAt,
	); err != nil {
		return "", fmt.Errorf("create mock auth session: %w", err)
	}
	return token, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*Account, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrUnauthorized
	}
	account, err := s.repository.GetAccountBySessionTokenHash(
		ctx,
		hashSessionToken(token),
	)
	if err != nil {
		if errors.Is(err, ErrInvalidSession) {
			return nil, err
		}
		return nil, fmt.Errorf("authenticate mock session: %w", err)
	}
	return account, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrUnauthorized
	}
	if err := s.repository.DeleteSessionByTokenHash(
		ctx,
		hashSessionToken(token),
	); err != nil {
		return fmt.Errorf("delete mock auth session: %w", err)
	}
	return nil
}

func (s *Service) generateSessionToken() (string, []byte, error) {
	raw := make([]byte, sessionTokenBytes)
	if n, err := s.randomRead(raw); err != nil || n != len(raw) {
		if err == nil {
			err = errors.New("short random read")
		}
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, hashSessionToken(token), nil
}

func hashSessionToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
