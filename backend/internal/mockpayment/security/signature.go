package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const signatureVersion = "v1="

var (
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrStaleSignature   = errors.New("webhook signature timestamp is outside the allowed window")
)

type Signer struct {
	secret []byte
}

func NewSigner(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

func (s *Signer) Sign(timestamp time.Time, body []byte) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = fmt.Fprintf(mac, "%d.", timestamp.Unix())
	_, _ = mac.Write(body)
	return signatureVersion + hex.EncodeToString(mac.Sum(nil))
}

func Verify(secret string, timestamp string, signature string, body []byte, now time.Time, tolerance time.Duration) error {
	unix, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}
	signedAt := time.Unix(unix, 0)
	if tolerance > 0 && now.Sub(signedAt).Abs() > tolerance {
		return ErrStaleSignature
	}

	expected := NewSigner(secret).Sign(signedAt, body)
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature))) {
		return ErrInvalidSignature
	}
	return nil
}
