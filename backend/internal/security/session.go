package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const sessionTokenLength = 32

func HashSessionToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

func GenerateSessionToken() (string, []byte, error) {
	b := make([]byte, sessionTokenLength)

	_, err := rand.Read(b)
	if err != nil {
		return "", nil, err
	}

	sessionToken := base64.URLEncoding.EncodeToString(b)
	return sessionToken, HashSessionToken(sessionToken), nil
}
