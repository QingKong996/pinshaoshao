package repository

import (
	"errors"
	"math"
	"testing"

	"pinshaoshao/internal/feature/order/apperr"
)

func TestSafeLineAmount(t *testing.T) {
	tests := []struct {
		name     string
		unit     int64
		quantity int
		want     int64
		wantErr  bool
	}{
		{name: "zero price", unit: 0, quantity: 2, want: 0},
		{name: "regular price", unit: 1999, quantity: 3, want: 5997},
		{name: "negative price", unit: -1, quantity: 1, wantErr: true},
		{name: "zero quantity", unit: 1, quantity: 0, wantErr: true},
		{name: "multiplication overflow", unit: math.MaxInt64, quantity: 2, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeLineAmount(tt.unit, tt.quantity)
			if tt.wantErr {
				if !errors.Is(err, apperr.ErrOrderAmountInvalid) {
					t.Fatalf("error = %v, want %v", err, apperr.ErrOrderAmountInvalid)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("amount/error = %d/%v, want %d/nil", got, err, tt.want)
			}
		})
	}
}

func TestSafeAddAmount(t *testing.T) {
	tests := []struct {
		name    string
		total   int64
		amount  int64
		want    int64
		wantErr bool
	}{
		{name: "regular", total: 1999, amount: 2999, want: 4998},
		{name: "negative amount", total: 1, amount: -1, wantErr: true},
		{name: "addition overflow", total: math.MaxInt64, amount: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeAddAmount(tt.total, tt.amount)
			if tt.wantErr {
				if !errors.Is(err, apperr.ErrOrderAmountInvalid) {
					t.Fatalf("error = %v, want %v", err, apperr.ErrOrderAmountInvalid)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("amount/error = %d/%v, want %d/nil", got, err, tt.want)
			}
		})
	}
}
