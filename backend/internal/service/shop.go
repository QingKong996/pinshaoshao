package service

import (
	"context"
	"errors"
	"fmt"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/model"
	"pinshaoshao/internal/security"
)

type RegisterShopInput struct {
	Username string
	Name     string
	Password string
	Email    string
}

type UpdateShopProfileInput struct {
	Name        string
	Description string
}

// ShopRepository is the persistence the shop service depends on.
type ShopRepository interface {
	GetShopByUsername(ctx context.Context, username string) (*model.Shop, error)
	GetShopByID(ctx context.Context, id int) (*model.Shop, error)
	CreateShopWithSession(ctx context.Context, shop *model.Shop, tokenHash []byte) error
	CreateSession(ctx context.Context, shop *model.Shop, tokenHash []byte) error
	GetShopBySessionTokenHash(ctx context.Context, tokenHash []byte) (*model.Shop, error)
	UpdateShopProfile(ctx context.Context, shop *model.Shop) (*model.Shop, error)
	DeleteSessionByTokenHash(ctx context.Context, tokenHash []byte) error
}

type ShopService struct {
	shopRepository ShopRepository
}

func NewShopService(shopRepository ShopRepository) *ShopService {
	return &ShopService{shopRepository: shopRepository}
}

func (s *ShopService) Login(ctx context.Context, input LoginInput) (string, error) {
	shop, err := s.shopRepository.GetShopByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, apperr.ErrShopNotFound) {
			security.DummyVerify(input.Password)
			return "", fmt.Errorf("service.ShopLogin: %w", apperr.ErrInvalidCredentials)
		}
		return "", fmt.Errorf("service.ShopLogin: get shop by username: %w", err)
	}

	res, err := security.VerifyPassword(input.Password, shop.PasswordHash)
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

func (s *ShopService) Register(ctx context.Context, input RegisterShopInput) (string, error) {
	passwordHash, err := security.HashPassword(input.Password)
	if err != nil {
		return "", fmt.Errorf("service.ShopRegister: hash password: %w", err)
	}

	shop := &model.Shop{
		Username:     input.Username,
		Name:         input.Name,
		PasswordHash: passwordHash,
		Email:        input.Email,
	}

	token, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("service.ShopRegister: generate session token: %w", err)
	}

	if err := s.shopRepository.CreateShopWithSession(ctx, shop, tokenHash); err != nil {
		return "", fmt.Errorf("service.ShopRegister: create shop with session: %w", err)
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

// GetProfile returns the profile of the shop.
func (s *ShopService) GetProfile(ctx context.Context, shopID int) (*model.Shop, error) {
	shop, err := s.shopRepository.GetShopByID(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("service.ShopGetProfile: %w", err)
	}
	return shop, nil
}

// UpdateProfile updates the editable profile fields of the shop.
func (s *ShopService) UpdateProfile(ctx context.Context, shopID int, input UpdateShopProfileInput) (*model.Shop, error) {
	shop, err := s.shopRepository.UpdateShopProfile(ctx, &model.Shop{
		ID:          shopID,
		Name:        input.Name,
		Description: input.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("service.ShopUpdateProfile: %w", err)
	}
	return shop, nil
}
