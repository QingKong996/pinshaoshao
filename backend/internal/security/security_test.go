package security_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"pinshaoshao/internal/security"
)

func TestPasswordHashing(t *testing.T) {
	password := "correct horse battery staple"

	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == password {
		t.Fatal("password hash contains the plaintext password")
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("hash = %q, want an argon2id encoded hash", hash)
	}

	ok, err := security.VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}
	if !ok {
		t.Fatal("password verification failed for the original password")
	}

	ok, err = security.VerifyPassword("wrong password", hash)
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}
	if ok {
		t.Fatal("password verification succeeded for a wrong password")
	}

	otherHash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password again: %v", err)
	}
	if hash == otherHash {
		t.Fatal("two hashes of the same password are identical")
	}
}

func TestVerifyPasswordRejectsInvalidHashes(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	parts := strings.Split(hash, "$")
	withPart := func(index int, value string) string {
		t.Helper()

		changed := append([]string(nil), parts...)
		changed[index] = value
		return strings.Join(changed, "$")
	}
	encoded := func(value []byte) string {
		return base64.RawStdEncoding.EncodeToString(value)
	}

	tests := map[string]string{
		"empty":                 "",
		"unsupported algorithm": withPart(1, "argon2i"),
		"unsupported version":   withPart(2, "v=18"),
		"malformed parameters":  withPart(3, "m=19456,t=2"),
		"excessive threads":     withPart(3, "m=19456,t=2,p=17"),
		"invalid salt":          withPart(4, "%%%"),
		"short salt":            withPart(4, encoded([]byte("short"))),
		"invalid hash":          withPart(5, "%%%"),
		"short hash":            withPart(5, encoded(make([]byte, 16))),
	}

	for name, invalidHash := range tests {
		t.Run(name, func(t *testing.T) {
			ok, err := security.VerifyPassword(password, invalidHash)
			if err == nil {
				t.Fatal("verify password: expected an error")
			}
			if ok {
				t.Fatal("verify password: invalid hash was accepted")
			}
		})
	}
}

func TestSessionTokens(t *testing.T) {
	token, tokenHash, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate session token: %v", err)
	}
	if token == "" {
		t.Fatal("session token is empty")
	}

	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("decode session token: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("session token length = %d, want 32 bytes", len(decoded))
	}
	if !bytes.Equal(tokenHash, security.HashSessionToken(token)) {
		t.Fatal("returned token hash does not match the token")
	}
	if len(tokenHash) != 32 {
		t.Fatalf("token hash length = %d, want 32 bytes", len(tokenHash))
	}

	otherToken, _, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("generate another session token: %v", err)
	}
	if token == otherToken {
		t.Fatal("two generated session tokens are identical")
	}
}
