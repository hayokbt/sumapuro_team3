package account

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewRegistration(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		accountName string
		password    string
		valid       bool
	}{
		{name: "valid input", email: "Player@Example.com", accountName: "player", password: "strong-pass-123", valid: true},
		{name: "invalid email", email: "not-an-email", accountName: "player", password: "strong-pass-123"},
		{name: "blank account name", email: "player@example.com", accountName: " ", password: "strong-pass-123"},
		{name: "short password", email: "player@example.com", accountName: "player", password: "short"},
		{name: "password too long", email: "player@example.com", accountName: "player", password: strings.Repeat("a", MaxPasswordBytes+1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registration, err := NewRegistration(test.email, test.accountName, test.password)
			if test.valid && err != nil {
				t.Fatalf("expected valid registration, got %v", err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidRegistration) {
				t.Fatalf("expected invalid registration, got %v", err)
			}
			if test.valid && registration.Email != strings.ToLower(test.email) {
				t.Fatalf("email was not normalized: %q", registration.Email)
			}
		})
	}
}

func TestPendingRegistrationVerification(t *testing.T) {
	now := time.Now()
	pending := PendingRegistration{
		VerificationCodeHash: HashVerificationCode("042321"),
		ExpiresAt:            now.Add(time.Minute),
	}

	if err := pending.Verify("000000", now); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("expected incorrect code error, got %v", err)
	}
	if pending.VerificationAttempts != 1 {
		t.Fatalf("expected one failed attempt, got %d", pending.VerificationAttempts)
	}
	if err := pending.Verify("042321", now); err != nil {
		t.Fatalf("expected valid code, got %v", err)
	}
}

func TestPendingRegistrationVerificationLimits(t *testing.T) {
	now := time.Now()
	pending := PendingRegistration{
		VerificationCodeHash: HashVerificationCode("042321"),
		ExpiresAt:            now.Add(time.Minute),
	}
	for attempt := 1; attempt < MaxVerificationAttempts; attempt++ {
		if err := pending.Verify("000000", now); !errors.Is(err, ErrInvalidCode) {
			t.Fatalf("attempt %d: expected incorrect code error, got %v", attempt, err)
		}
	}
	if err := pending.Verify("000000", now); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("expected too many attempts error, got %v", err)
	}
}

func TestPendingRegistrationVerificationExpiry(t *testing.T) {
	now := time.Now()
	pending := PendingRegistration{
		VerificationCodeHash: HashVerificationCode("042321"),
		ExpiresAt:            now,
	}
	if err := pending.Verify("042321", now); !errors.Is(err, ErrExpiredCode) {
		t.Fatalf("expected expired code error, got %v", err)
	}
}

func TestNewVerificationCode(t *testing.T) {
	code, err := NewVerificationCode()
	if err != nil {
		t.Fatalf("generate verification code: %v", err)
	}
	if len(code) != 6 || !isDigits(code) {
		t.Fatalf("generated invalid verification code %q", code)
	}
}
