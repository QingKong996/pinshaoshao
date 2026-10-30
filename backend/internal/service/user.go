package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
)

// UserRepository is the persistence the user service depends on.
type UserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
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

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (string, error) {
	user, err := s.userRepository.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, apperr.ErrUserNotFound) {
			security.DummyVerify(req.Password)
			return "", fmt.Errorf("service.Login: %w", apperr.ErrInvalidCredentials)
		}
		return "", fmt.Errorf("service.Login: get user by username: %w", err)
	}

	res, err := security.VerifyPassword(req.Password, user.PasswordHash)
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

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (string, error) {
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return "", fmt.Errorf("service.Register: hash password: %w", err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
	}

	err = s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return "", fmt.Errorf("service.Register: create user: %w", err)
	}

	token, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("service.Register: generate session token: %w", err)
	}

	err = s.userRepository.CreateSession(ctx, user, tokenHash)
	if err != nil {
		return "", fmt.Errorf("service.Register: create session: %w", err)
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
