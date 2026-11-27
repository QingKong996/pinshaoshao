package service

import (
	"context"
	"fmt"
	"pinshaoshao/internal/feature/address/apperr"
	"pinshaoshao/internal/feature/address/model"
	"strings"
	"unicode/utf8"
)

const (
	maxReceiverNameLen  = 50
	maxReceiverPhoneLen = 32
	maxAddressLen       = 1000
)

type AddressInput struct {
	ReceiverName    string
	ReceiverPhone   string
	ShippingAddress string
	IsDefault       bool
}

type AddressRepository interface {
	ListByUserID(ctx context.Context, userID int) ([]model.Address, error)
	Create(ctx context.Context, userID int, address model.Address) (*model.Address, error)
	Update(ctx context.Context, userID int, addressID int, address model.Address) (*model.Address, error)
	Delete(ctx context.Context, userID int, addressID int) error
	SetDefault(ctx context.Context, userID int, addressID int) (*model.Address, error)
}

type AddressService struct {
	addressRepository AddressRepository
}

func NewAddressService(addressRepository AddressRepository) *AddressService {
	return &AddressService{addressRepository: addressRepository}
}

func (s *AddressService) List(ctx context.Context, userID int) ([]model.Address, error) {
	addresses, err := s.addressRepository.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.ListAddresses: %w", err)
	}
	return addresses, nil
}

func (s *AddressService) Create(ctx context.Context, userID int, input AddressInput) (*model.Address, error) {
	address, err := normalizeAddress(input)
	if err != nil {
		return nil, fmt.Errorf("service.CreateAddress: %w", err)
	}

	created, err := s.addressRepository.Create(ctx, userID, address)
	if err != nil {
		return nil, fmt.Errorf("service.CreateAddress: %w", err)
	}
	return created, nil
}

func (s *AddressService) Update(ctx context.Context, userID int, addressID int, input AddressInput) (*model.Address, error) {
	if addressID <= 0 {
		return nil, fmt.Errorf("service.UpdateAddress: %w", apperr.ErrInvalidAddressID)
	}

	address, err := normalizeAddress(input)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateAddress: %w", err)
	}

	updated, err := s.addressRepository.Update(ctx, userID, addressID, address)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateAddress: %w", err)
	}
	return updated, nil
}

func (s *AddressService) Delete(ctx context.Context, userID int, addressID int) error {
	if addressID <= 0 {
		return fmt.Errorf("service.DeleteAddress: %w", apperr.ErrInvalidAddressID)
	}
	if err := s.addressRepository.Delete(ctx, userID, addressID); err != nil {
		return fmt.Errorf("service.DeleteAddress: %w", err)
	}
	return nil
}

func (s *AddressService) SetDefault(ctx context.Context, userID int, addressID int) (*model.Address, error) {
	if addressID <= 0 {
		return nil, fmt.Errorf("service.SetDefaultAddress: %w", apperr.ErrInvalidAddressID)
	}

	address, err := s.addressRepository.SetDefault(ctx, userID, addressID)
	if err != nil {
		return nil, fmt.Errorf("service.SetDefaultAddress: %w", err)
	}
	return address, nil
}

func normalizeAddress(input AddressInput) (model.Address, error) {
	receiverName := strings.TrimSpace(input.ReceiverName)
	receiverPhone := strings.TrimSpace(input.ReceiverPhone)
	shippingAddress := strings.TrimSpace(input.ShippingAddress)

	if receiverName == "" || utf8.RuneCountInString(receiverName) > maxReceiverNameLen {
		return model.Address{}, apperr.ErrInvalidAddress
	}
	if receiverPhone == "" || utf8.RuneCountInString(receiverPhone) > maxReceiverPhoneLen {
		return model.Address{}, apperr.ErrInvalidAddress
	}
	if shippingAddress == "" || utf8.RuneCountInString(shippingAddress) > maxAddressLen {
		return model.Address{}, apperr.ErrInvalidAddress
	}

	return model.Address{
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		ShippingAddress: shippingAddress,
		IsDefault:       input.IsDefault,
	}, nil
}
