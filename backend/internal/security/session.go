package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func GenerateSessionToken() (string, []byte, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", nil, err
	}

	sessionToken := base64.URLEncoding.EncodeToString(b)
	hash := sha256.Sum256([]byte(sessionToken))
	return sessionToken, hash[:], nil
}
