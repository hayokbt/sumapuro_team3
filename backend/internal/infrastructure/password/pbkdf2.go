package password

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	iterations = 600000
	saltLength = 16
	keyLength  = 32
)

type PBKDF2 struct{}

func NewPBKDF2() *PBKDF2 {
	return &PBKDF2{}
}

func (p *PBKDF2) Hash(value string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash, err := pbkdf2.Key(sha256.New, value, salt, iterations, keyLength)
	if err != nil {
		return "", fmt.Errorf("derive password hash: %w", err)
	}
	return fmt.Sprintf(
		"pbkdf2-sha256$%d$%s$%s",
		iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (p *PBKDF2) Verify(value, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false, fmt.Errorf("invalid PBKDF2 password hash")
	}

	hashIterations, err := strconv.Atoi(parts[1])
	if err != nil || hashIterations < 1 || hashIterations > iterations {
		return false, fmt.Errorf("invalid PBKDF2 iteration count")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) != saltLength {
		return false, fmt.Errorf("invalid PBKDF2 salt")
	}
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(expectedHash) != keyLength {
		return false, fmt.Errorf("invalid PBKDF2 password hash")
	}

	actualHash, err := pbkdf2.Key(sha256.New, value, salt, hashIterations, len(expectedHash))
	if err != nil {
		return false, fmt.Errorf("derive password hash: %w", err)
	}
	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}
