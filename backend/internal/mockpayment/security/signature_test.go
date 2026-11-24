package security_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"pinshaoshao/internal/mockpayment/security"
)

func TestSignAndVerify(t *testing.T) {
	now := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	body := []byte(`{"id":"evt_1","type":"payment.paid"}`)
	signer := security.NewSigner("test-signing-key")
	signature := signer.Sign(now, body)
	timestamp := strconv.FormatInt(now.Unix(), 10)

	if err := security.Verify("test-signing-key", timestamp, signature, body, now, time.Minute); err != nil {
		t.Fatalf("verify signature: %v", err)
	}
	if err := security.Verify("test-signing-key", timestamp, signature, []byte(`{"tampered":true}`), now, time.Minute); !errors.Is(err, security.ErrInvalidSignature) {
		t.Fatalf("tampered body error = %v, want invalid signature", err)
	}
	if err := security.Verify("wrong-signing-key", timestamp, signature, body, now, time.Minute); !errors.Is(err, security.ErrInvalidSignature) {
		t.Fatalf("wrong secret error = %v, want invalid signature", err)
	}
	if err := security.Verify("test-signing-key", "not-a-timestamp", signature, body, now, time.Minute); !errors.Is(err, security.ErrInvalidSignature) {
		t.Fatalf("invalid timestamp error = %v, want invalid signature", err)
	}
	if err := security.Verify("test-signing-key", timestamp, signature, body, now.Add(2*time.Minute), time.Minute); !errors.Is(err, security.ErrStaleSignature) {
		t.Fatalf("stale timestamp error = %v, want stale signature", err)
	}
}
