package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
)

type LoginInput struct {
	Username string
	Password string
}

type RegisterUserInput struct {
	Username string
	Password string
	Email    string
}

// UserRepository is the persistence the user service depends on.
type UserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUserWithSession(ctx context.Context, user *model.User, tokenHash []byte) error
	CreateSession(ctx context.Context, user *model.User, tokenHash []byte) error
	GetUserBySessionTokenHash(ctx context.Context, tokenHash []byte) (*model.User, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash []byte) error
}

type UserService struct {
	userRepository UserRepository
}

func NewUserService(userRepository UserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (string, error) {
	user, err := s.userRepository.GetUserByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			security.DummyVerify(input.Password)
			return "", fmt.Errorf("service.Login: %w", apperr.ErrInvalidCredentials)
		}
		return "", fmt.Errorf("service.Login: get user by username: %w", err)
	}

	res, err := security.VerifyPassword(input.Password, user.PasswordHash)
	if err != nil {
		return "", fmt.Errorf("service.Login: verify password: %w: %w", apperr.ErrInvalidCredentials, err)
	}
	if !res {
		return "", fmt.Errorf("service.Login: verify password: %w", apperr.ErrInvalidCredentials)
	}

	if user.Status != model.UserStatusActive {
		return "", fmt.Errorf("service.Login: %w", apperr.ErrUserBanned)
	}

	token, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("service.Login: generate session token: %w", err)
	}

	err = s.userRepository.CreateSession(ctx, user, tokenHash)
	if err != nil {
		return "", fmt.Errorf("service.Login: create session: %w", err)
	}

	return token, nil
}

func (s *UserService) Register(ctx context.Context, input RegisterUserInput) (string, error) {
	passwordHash, err := security.HashPassword(input.Password)
	if err != nil {
		return "", fmt.Errorf("service.Register: hash password: %w", err)
	}

	user := &model.User{
		Username:     input.Username,
		PasswordHash: passwordHash,
		Email:        input.Email,
	}

	token, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("service.Register: generate session token: %w", err)
	}

	if err := s.userRepository.CreateUserWithSession(ctx, user, tokenHash); err != nil {
		return "", fmt.Errorf("service.Register: create user with session: %w", err)
	}

	return token, nil
}

// Authenticate resolves a session token to an active user.
func (s *UserService) Authenticate(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, fmt.Errorf("service.Authenticate: %w", apperr.ErrUnauthorized)
	}

	user, err := s.userRepository.GetUserBySessionTokenHash(ctx, security.HashSessionToken(token))
	if err != nil {
		return nil, fmt.Errorf("service.Authenticate: %w", err)
	}

	if user.Status != model.UserStatusActive {
		return nil, fmt.Errorf("service.Authenticate: %w", apperr.ErrUserBanned)
	}

	return user, nil
}

// Logout revokes the session identified by token.
func (s *UserService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("service.Logout: %w", apperr.ErrUnauthorized)
	}

	err := s.userRepository.DeleteSessionByTokenHash(ctx, security.HashSessionToken(token))
	if err != nil {
		return fmt.Errorf("service.Logout: delete session: %w", err)
	}

	return nil
}
