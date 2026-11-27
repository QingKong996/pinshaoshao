//go:build integration

package repository_test

import (
	"context"
	"testing"

	"pinshaoshao/internal/feature/address/apperr"
	"pinshaoshao/internal/feature/address/model"
)

func TestAddressRepositoryDefaults(t *testing.T) {
	ctx := context.Background()
	repository := newTestAddressRepository()

	resetAddressDatabase(t)
	user := insertAddressUser(t, "buyer01")

	first, err := repository.Create(ctx, user.ID, model.Address{
		ReceiverName:    "Alice",
		ReceiverPhone:   "123",
		ShippingAddress: "Shanghai",
	})
	if err != nil {
		t.Fatalf("create first address: %v", err)
	}
	if !first.IsDefault {
		t.Fatalf("first address = %+v, want default", first)
	}

	second, err := repository.Create(ctx, user.ID, model.Address{
		ReceiverName:    "Bob",
		ReceiverPhone:   "456",
		ShippingAddress: "Beijing",
	})
	if err != nil {
		t.Fatalf("create second address: %v", err)
	}
	if second.IsDefault {
		t.Fatalf("second address = %+v, want non-default", second)
	}

	third, err := repository.Create(ctx, user.ID, model.Address{
		ReceiverName:    "Carol",
		ReceiverPhone:   "789",
		ShippingAddress: "Shenzhen",
		IsDefault:       true,
	})
	if err != nil {
		t.Fatalf("create default address: %v", err)
	}
	if !third.IsDefault {
		t.Fatalf("third address = %+v, want default", third)
	}

	addresses, err := repository.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list addresses: %v", err)
	}
	if len(addresses) != 3 {
		t.Fatalf("address count = %d, want 3", len(addresses))
	}
	if !addresses[0].IsDefault || addresses[0].ID != third.ID {
		t.Fatalf("addresses = %+v, want new default first", addresses)
	}
	defaultCount := 0
	for _, address := range addresses {
		if address.IsDefault {
			defaultCount++
		}
		if address.ID == first.ID && address.IsDefault {
			t.Fatalf("first address is still default: %+v", address)
		}
	}
	if defaultCount != 1 {
		t.Fatalf("default count = %d, want 1", defaultCount)
	}
}

func TestAddressRepositoryUpdateDeleteAndOwnership(t *testing.T) {
	ctx := context.Background()
	repository := newTestAddressRepository()

	resetAddressDatabase(t)
	user := insertAddressUser(t, "buyer01")
	otherUser := insertAddressUser(t, "buyer02")

	first, err := repository.Create(ctx, user.ID, model.Address{
		ReceiverName:    "Alice",
		ReceiverPhone:   "123",
		ShippingAddress: "Shanghai",
	})
	if err != nil {
		t.Fatalf("create first address: %v", err)
	}
	second, err := repository.Create(ctx, user.ID, model.Address{
		ReceiverName:    "Bob",
		ReceiverPhone:   "456",
		ShippingAddress: "Beijing",
	})
	if err != nil {
		t.Fatalf("create second address: %v", err)
	}

	_, err = repository.Update(ctx, otherUser.ID, first.ID, model.Address{
		ReceiverName:    "Mallory",
		ReceiverPhone:   "000",
		ShippingAddress: "Other",
		IsDefault:       true,
	})
	requireAddressErrorIs(t, err, apperr.ErrAddressNotFound)

	_, err = repository.SetDefault(ctx, otherUser.ID, second.ID)
	requireAddressErrorIs(t, err, apperr.ErrAddressNotFound)

	updated, err := repository.Update(ctx, user.ID, second.ID, model.Address{
		ReceiverName:    "Robert",
		ReceiverPhone:   "456",
		ShippingAddress: "Beijing",
		IsDefault:       true,
	})
	if err != nil {
		t.Fatalf("update address: %v", err)
	}
	if updated.ReceiverName != "Robert" || !updated.IsDefault {
		t.Fatalf("address = %+v, want updated default", updated)
	}

	first, err = repository.SetDefault(ctx, user.ID, first.ID)
	if err != nil {
		t.Fatalf("set default address: %v", err)
	}
	if !first.IsDefault {
		t.Fatalf("address = %+v, want default", first)
	}

	if err := repository.Delete(ctx, user.ID, first.ID); err != nil {
		t.Fatalf("delete default address: %v", err)
	}
	addresses, err := repository.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list addresses: %v", err)
	}
	if len(addresses) != 1 || addresses[0].ID != second.ID || !addresses[0].IsDefault {
		t.Fatalf("addresses = %+v, want second address promoted to default", addresses)
	}

	err = repository.Delete(ctx, user.ID, second.ID)
	if err != nil {
		t.Fatalf("delete last address: %v", err)
	}
	addresses, err = repository.ListByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("list empty addresses: %v", err)
	}
	if len(addresses) != 0 {
		t.Fatalf("addresses = %+v, want empty", addresses)
	}
}
