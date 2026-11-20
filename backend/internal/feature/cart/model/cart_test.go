package model_test

import (
	"testing"

	"pinshaoshao/internal/feature/cart/model"
)

func TestCartItemAvailability(t *testing.T) {
	onShelf := "on_shelf"
	offShelf := "off_shelf"
	enabled := true
	disabled := false
	stock := 5

	tests := []struct {
		name string
		item model.CartItem
		want model.CartItemAvailability
	}{
		{
			name: "available",
			item: model.CartItem{
				Quantity:       2,
				ProductStatus:  &onShelf,
				SKUEnabled:     &enabled,
				AvailableStock: &stock,
			},
			want: model.CartItemAvailable,
		},
		{
			name: "deleted product",
			item: model.CartItem{Quantity: 1},
			want: model.CartItemProductDeleted,
		},
		{
			name: "off shelf product",
			item: model.CartItem{ProductStatus: &offShelf, SKUEnabled: &enabled, AvailableStock: &stock},
			want: model.CartItemProductOffShelf,
		},
		{
			name: "deleted sku",
			item: model.CartItem{ProductStatus: &onShelf, AvailableStock: &stock},
			want: model.CartItemSKUDeleted,
		},
		{
			name: "disabled sku",
			item: model.CartItem{ProductStatus: &onShelf, SKUEnabled: &disabled, AvailableStock: &stock},
			want: model.CartItemSKUDisabled,
		},
		{
			name: "out of stock",
			item: model.CartItem{Quantity: 1, ProductStatus: &onShelf, SKUEnabled: &enabled, AvailableStock: intPtr(0)},
			want: model.CartItemOutOfStock,
		},
		{
			name: "insufficient stock",
			item: model.CartItem{Quantity: 3, ProductStatus: &onShelf, SKUEnabled: &enabled, AvailableStock: intPtr(2)},
			want: model.CartItemInsufficientStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.item.Availability(); got != tt.want {
				t.Fatalf("availability = %q, want %q", got, tt.want)
			}
		})
	}
}

func intPtr(value int) *int {
	return &value
}
