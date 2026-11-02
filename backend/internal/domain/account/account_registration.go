package account

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	VerificationCodeLifetime = 10 * time.Minute
	MaxVerificationAttempts  = 5
	MaxPasswordBytes         = 72
)

var (
	ErrInvalidRegistration = errors.New("invalid registration data")
	ErrAccountExists       = errors.New("email or account name already exists")
	ErrInvalidCode         = errors.New("verification code is incorrect")
	ErrExpiredCode         = errors.New("verification code has expired")
	ErrTooManyAttempts     = errors.New("too many verification attempts")
	ErrEmailDelivery       = errors.New("verification email delivery failed")
)

type Registration struct {
	Email       string
	AccountName string
	Password    string
}

type PendingRegistration struct {
	ID                   int64
	Email                string
	AccountName          string
	PasswordHash         string
	VerificationCodeHash string
	VerificationAttempts uint8
	ExpiresAt            time.Time
}

type RegistrationRepository interface {
	SavePending(ctx context.Context, pending PendingRegistration) (int64, error)
	RemovePending(ctx context.Context, id int64) error
	CompleteVerification(ctx context.Context, email, code string, now time.Time) error
}

type VerificationEmailSender interface {
	SendVerificationCode(to, code string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
}

func NewRegistration(email, accountName, password string) (Registration, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	accountName = strings.TrimSpace(accountName)

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || len(email) > 254 {
		return Registration{}, ErrInvalidRegistration
	}
	if accountName == "" || utf8.RuneCountInString(accountName) > 50 {
		return Registration{}, ErrInvalidRegistration
	}
	if len(password) < 8 || len(password) > MaxPasswordBytes {
		return Registration{}, ErrInvalidRegistration
	}

	return Registration{
		Email:       email,
		AccountName: accountName,
		Password:    password,
	}, nil
}

func NormalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email || len(email) > 254 {
		return "", ErrInvalidRegistration
	}
	return email, nil
}

func NewVerificationCode() (string, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", number.Int64()), nil
}

func HashVerificationCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return fmt.Sprintf("%x", hash)
}

func (p *PendingRegistration) Verify(code string, now time.Time) error {
	if !now.Before(p.ExpiresAt) {
		return ErrExpiredCode
	}
	if p.VerificationAttempts >= MaxVerificationAttempts {
		return ErrTooManyAttempts
	}
	if len(code) != 6 || !isDigits(code) {
		return p.recordFailedAttempt()
	}
	if subtle.ConstantTimeCompare(
		[]byte(p.VerificationCodeHash),
		[]byte(HashVerificationCode(code)),
	) != 1 {
		return p.recordFailedAttempt()
	}
	return nil
}

func (p *PendingRegistration) recordFailedAttempt() error {
	p.VerificationAttempts++
	if p.VerificationAttempts >= MaxVerificationAttempts {
		return ErrTooManyAttempts
	}
	return ErrInvalidCode
}

func isDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
