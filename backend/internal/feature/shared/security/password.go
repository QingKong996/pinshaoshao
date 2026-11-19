package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory  uint32 = 19 * 1024
	time    uint32 = 2
	threads uint8  = 1
	keyLen  uint32 = 32
)

// DummyVerify performs the same Argon2id work as VerifyPassword and discards
// the result, so logins for unknown users take just as long.
func DummyVerify(password string) {
	var salt [16]byte
	_ = argon2.IDKey([]byte(password), salt[:], time, memory, threads, keyLen)
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		time,
		threads,
		encodedSalt,
		encodedHash,
	)

	return encoded, nil
}

func VerifyPassword(password string, encodedHash string) (bool, error) {

	var invalidEncodedHashError error = errors.New("security.VerifyPassword: invalid encoded hash")

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 ||
		parts[0] != "" ||
		parts[1] != "argon2id" ||
		parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false, invalidEncodedHashError
	}

	var memory uint32
	var time uint32
	var threads uint8

	n, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&memory,
		&time,
		&threads,
	)
	if err != nil || n != 3 {
		return false, invalidEncodedHashError
	}
	if parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, time, threads) ||
		memory < 8*uint32(threads) ||
		memory > 256*1024 ||
		time == 0 || time > 10 ||
		threads == 0 || threads > 16 {
		return false, invalidEncodedHashError
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return false, invalidEncodedHashError
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(storedHash) != 32 {
		return false, invalidEncodedHashError
	}

	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		threads,
		uint32(len(storedHash)),
	)

	if subtle.ConstantTimeCompare(computedHash, storedHash) != 1 {
		return false, nil
	}

	return true, nil
}
