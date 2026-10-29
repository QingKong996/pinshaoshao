package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/repository"
	"pinshaoshao/internal/security"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(userRepository *repository.UserRepository) *UserService {
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
