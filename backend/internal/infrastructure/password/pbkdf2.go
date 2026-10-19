package password

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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
