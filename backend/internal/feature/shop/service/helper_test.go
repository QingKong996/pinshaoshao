package service_test

import (
	"errors"
	"testing"

	"pinshaoshao/internal/security"
)

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return hash
}

func requireErrorIs(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("error = %v, want %v", err, target)
	}
}
