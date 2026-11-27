package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"pinshaoshao/internal/feature/address/apperr"
	"pinshaoshao/internal/feature/address/model"
	"pinshaoshao/internal/feature/address/service"
)

type fakeAddressRepository struct {
	listFn       func(context.Context, int) ([]model.Address, error)
	createFn     func(context.Context, int, model.Address) (*model.Address, error)
	updateFn     func(context.Context, int, int, model.Address) (*model.Address, error)
	deleteFn     func(context.Context, int, int) error
	setDefaultFn func(context.Context, int, int) (*model.Address, error)
}

func (f fakeAddressRepository) ListByUserID(ctx context.Context, userID int) ([]model.Address, error) {
	return f.listFn(ctx, userID)
}

func (f fakeAddressRepository) Create(ctx context.Context, userID int, address model.Address) (*model.Address, error) {
	return f.createFn(ctx, userID, address)
}

func (f fakeAddressRepository) Update(ctx context.Context, userID int, addressID int, address model.Address) (*model.Address, error) {
	return f.updateFn(ctx, userID, addressID, address)
}

func (f fakeAddressRepository) Delete(ctx context.Context, userID int, addressID int) error {
	return f.deleteFn(ctx, userID, addressID)
}

func (f fakeAddressRepository) SetDefault(ctx context.Context, userID int, addressID int) (*model.Address, error) {
	return f.setDefaultFn(ctx, userID, addressID)
}

func TestAddressServiceCreateNormalizesInput(t *testing.T) {
	ctx := context.Background()
	want := &model.Address{ID: 1, ReceiverName: "Alice"}
	addressService := service.NewAddressService(fakeAddressRepository{
		createFn: func(_ context.Context, userID int, address model.Address) (*model.Address, error) {
			if userID != 7 {
				t.Fatalf("user id = %d, want 7", userID)
			}
			if address.ReceiverName != "Alice" ||
				address.ReceiverPhone != "13800138000" ||
				address.ShippingAddress != "Shanghai" ||
				!address.IsDefault {
				t.Fatalf("address = %+v, want trimmed values and default", address)
			}
			return want, nil
		},
	})

	got, err := addressService.Create(ctx, 7, service.AddressInput{
		ReceiverName:    " Alice ",
		ReceiverPhone:   " 13800138000 ",
		ShippingAddress: " Shanghai ",
		IsDefault:       true,
	})
	if err != nil {
		t.Fatalf("create address: %v", err)
	}
	if got != want {
		t.Fatalf("address = %+v, want %+v", got, want)
	}
}

func TestAddressServiceRejectsInvalidInput(t *testing.T) {
	addressService := service.NewAddressService(fakeAddressRepository{})

	tests := []struct {
		name  string
		input service.AddressInput
	}{
		{name: "blank recipient", input: service.AddressInput{ReceiverPhone: "123", ShippingAddress: "Shanghai"}},
		{name: "blank phone", input: service.AddressInput{ReceiverName: "Alice", ShippingAddress: "Shanghai"}},
		{name: "blank address", input: service.AddressInput{ReceiverName: "Alice", ReceiverPhone: "123"}},
		{name: "long recipient", input: service.AddressInput{ReceiverName: strings.Repeat("a", 51), ReceiverPhone: "123", ShippingAddress: "Shanghai"}},
		{name: "long phone", input: service.AddressInput{ReceiverName: "Alice", ReceiverPhone: strings.Repeat("1", 33), ShippingAddress: "Shanghai"}},
		{name: "long address", input: service.AddressInput{ReceiverName: "Alice", ReceiverPhone: "123", ShippingAddress: strings.Repeat("a", 1001)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := addressService.Create(context.Background(), 7, tt.input)
			requireAddressErrorIs(t, err, apperr.ErrInvalidAddress)
		})
	}
}

func TestAddressServiceRepositoryOperations(t *testing.T) {
	ctx := context.Background()
	address := &model.Address{ID: 3, IsDefault: true}

	t.Run("lists addresses", func(t *testing.T) {
		addressService := service.NewAddressService(fakeAddressRepository{
			listFn: func(context.Context, int) ([]model.Address, error) {
				return []model.Address{*address}, nil
			},
		})

		got, err := addressService.List(ctx, 7)
		if err != nil {
			t.Fatalf("list addresses: %v", err)
		}
		if len(got) != 1 || got[0].ID != address.ID {
			t.Fatalf("addresses = %+v", got)
		}
	})

	t.Run("updates an owned address", func(t *testing.T) {
		addressService := service.NewAddressService(fakeAddressRepository{
			updateFn: func(_ context.Context, userID int, addressID int, input model.Address) (*model.Address, error) {
				if userID != 7 || addressID != 3 || input.ReceiverName != "Alice" {
					t.Fatalf("update args = %d/%d/%+v", userID, addressID, input)
				}
				return address, nil
			},
		})

		got, err := addressService.Update(ctx, 7, 3, service.AddressInput{
			ReceiverName:    "Alice",
			ReceiverPhone:   "123",
			ShippingAddress: "Shanghai",
		})
		if err != nil {
			t.Fatalf("update address: %v", err)
		}
		if got != address {
			t.Fatalf("address = %+v, want %+v", got, address)
		}
	})

	t.Run("deletes an address", func(t *testing.T) {
		addressService := service.NewAddressService(fakeAddressRepository{
			deleteFn: func(_ context.Context, userID int, addressID int) error {
				if userID != 7 || addressID != 3 {
					t.Fatalf("delete args = %d/%d", userID, addressID)
				}
				return nil
			},
		})

		if err := addressService.Delete(ctx, 7, 3); err != nil {
			t.Fatalf("delete address: %v", err)
		}
	})

	t.Run("sets the default address", func(t *testing.T) {
		addressService := service.NewAddressService(fakeAddressRepository{
			setDefaultFn: func(context.Context, int, int) (*model.Address, error) {
				return address, nil
			},
		})

		got, err := addressService.SetDefault(ctx, 7, 3)
		if err != nil {
			t.Fatalf("set default address: %v", err)
		}
		if got != address {
			t.Fatalf("address = %+v, want %+v", got, address)
		}
	})

	t.Run("rejects invalid ids", func(t *testing.T) {
		addressService := service.NewAddressService(fakeAddressRepository{})

		_, err := addressService.Update(ctx, 7, 0, service.AddressInput{})
		requireAddressErrorIs(t, err, apperr.ErrInvalidAddressID)
		err = addressService.Delete(ctx, 7, 0)
		requireAddressErrorIs(t, err, apperr.ErrInvalidAddressID)
		_, err = addressService.SetDefault(ctx, 7, 0)
		requireAddressErrorIs(t, err, apperr.ErrInvalidAddressID)
	})

	t.Run("preserves repository failures", func(t *testing.T) {
		dbErr := errors.New("database is down")
		addressService := service.NewAddressService(fakeAddressRepository{
			listFn: func(context.Context, int) ([]model.Address, error) {
				return nil, dbErr
			},
		})

		_, err := addressService.List(ctx, 7)
		requireAddressErrorIs(t, err, dbErr)
	})
}

func requireAddressErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}
