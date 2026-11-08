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

// ShopRepository is the persistence the shop service depends on.
type ShopRepository interface {
	GetShopByUsername(ctx context.Context, username string) (*model.Shop, error)
	CreateShop(ctx context.Context, shop *model.Shop) error
	CreateSession(ctx context.Context, shop *model.Shop, tokenHash []byte) error
	GetShopBySessionTokenHash(ctx context.Context, tokenHash []byte) (*model.Shop, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash []byte) error
}

type ShopService struct {
	shopRepository ShopRepository
}

func NewShopService(shopRepository ShopRepository) *ShopService {
	return &ShopService{shopRepository: shopRepository}
}

func (s *ShopService) Login(ctx context.Context, req *dto.LoginRequest) (string, error) {
	shop, err := s.shopRepository.GetShopByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, apperr.ErrShopNotFound) {
			security.DummyVerify(req.Password)
			return "", fmt.Errorf("service.ShopLogin: %w", apperr.ErrInvalidCredentials)
		}
		return "", fmt.Errorf("service.ShopLogin: get shop by username: %w", err)
	}

	res, err := security.VerifyPassword(req.Password, shop.PasswordHash)
	if err != nil {
		return "", fmt.Errorf("service.ShopLogin: verify password: %w: %w", apperr.ErrInvalidCredentials, err)
	}
	if !res {
		return "", fmt.Errorf("service.ShopLogin: verify password: %w", apperr.ErrInvalidCredentials)
	}

	if shop.Status != model.ShopStatusActive {
		return "", fmt.Errorf("service.ShopLogin: %w", apperr.ErrShopBanned)
	}

	token, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("service.ShopLogin: generate session token: %w", err)
	}

	err = s.shopRepository.CreateSession(ctx, shop, tokenHash)
	if err != nil {
		return "", fmt.Errorf("service.ShopLogin: create session: %w", err)
	}

	return token, nil
}

func (s *ShopService) Register(ctx context.Context, req *dto.ShopRegisterRequest) (string, error) {
	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return "", fmt.Errorf("service.ShopRegister: hash password: %w", err)
	}

	shop := &model.Shop{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
	}

	err = s.shopRepository.CreateShop(ctx, shop)
	if err != nil {
		return "", fmt.Errorf("service.ShopRegister: create shop: %w", err)
	}

	token, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("service.ShopRegister: generate session token: %w", err)
	}

	err = s.shopRepository.CreateSession(ctx, shop, tokenHash)
	if err != nil {
		return "", fmt.Errorf("service.ShopRegister: create session: %w", err)
	}

	return token, nil
}

// Authenticate resolves a session token to an active shop.
func (s *ShopService) Authenticate(ctx context.Context, token string) (*model.Shop, error) {
	if token == "" {
		return nil, fmt.Errorf("service.ShopAuthenticate: %w", apperr.ErrUnauthorized)
	}

	shop, err := s.shopRepository.GetShopBySessionTokenHash(ctx, security.HashSessionToken(token))
	if err != nil {
		return nil, fmt.Errorf("service.ShopAuthenticate: %w", err)
	}

	if shop.Status != model.ShopStatusActive {
		return nil, fmt.Errorf("service.ShopAuthenticate: %w", apperr.ErrShopBanned)
	}

	return shop, nil
}

// Logout revokes the session identified by token.
func (s *ShopService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("service.ShopLogout: %w", apperr.ErrUnauthorized)
	}

	err := s.shopRepository.DeleteSessionByTokenHash(ctx, security.HashSessionToken(token))
	if err != nil {
		return fmt.Errorf("service.ShopLogout: delete session: %w", err)
	}

	return nil
}
