package accountregistration

import (
	"context"
	"errors"
	"testing"
	"time"

	"gomoku-backend/internal/domain/account"
)

type registrationRepositoryStub struct {
	pending       account.PendingRegistration
	pendingID     int64
	removedID     int64
	verifiedEmail string
	verifiedCode  string
	verifiedAt    time.Time
	saveErr       error
	verifyErr     error
}

func (r *registrationRepositoryStub) SavePending(_ context.Context, pending account.PendingRegistration) (int64, error) {
	r.pending = pending
	return r.pendingID, r.saveErr
}

func (r *registrationRepositoryStub) RemovePending(_ context.Context, id int64) error {
	r.removedID = id
	return nil
}

func (r *registrationRepositoryStub) CompleteVerification(_ context.Context, email, code string, now time.Time) error {
	r.verifiedEmail = email
	r.verifiedCode = code
	r.verifiedAt = now
	return r.verifyErr
}

type verificationEmailSenderStub struct {
	email string
	code  string
	err   error
}

func (s *verificationEmailSenderStub) SendVerificationCode(email, code string) error {
	s.email = email
	s.code = code
	return s.err
}

type passwordHasherStub struct{}

func (passwordHasherStub) Hash(string) (string, error) {
	return "hashed-password", nil
}

func TestStartSavesRegistrationAndSendsVerificationCode(t *testing.T) {
	repository := &registrationRepositoryStub{pendingID: 42}
	sender := &verificationEmailSenderStub{}
	registration := NewAccountRegistration(repository, sender, passwordHasherStub{})

	if err := registration.Start(context.Background(), " Player@Example.com ", " player ", "strong-pass-123"); err != nil {
		t.Fatalf("start registration: %v", err)
	}
	if repository.pending.Email != "player@example.com" || sender.email != repository.pending.Email {
		t.Fatalf("email was not normalized consistently: pending=%q sent=%q", repository.pending.Email, sender.email)
	}
	if repository.pending.AccountName != "player" || repository.pending.PasswordHash != "hashed-password" {
		t.Fatalf("unexpected pending registration: %+v", repository.pending)
	}
	if repository.pending.VerificationCodeHash != account.HashVerificationCode(sender.code) {
		t.Fatal("saved verification code hash does not match the sent code")
	}
	if len(sender.code) != 6 || repository.pending.ExpiresAt.Before(time.Now()) {
		t.Fatalf("unexpected verification code or expiry: code=%q expires=%v", sender.code, repository.pending.ExpiresAt)
	}
}

func TestStartRemovesRegistrationWhenEmailFails(t *testing.T) {
	repository := &registrationRepositoryStub{pendingID: 42}
	sender := &verificationEmailSenderStub{err: errors.New("smtp unavailable")}
	registration := NewAccountRegistration(repository, sender, passwordHasherStub{})

	err := registration.Start(context.Background(), "player@example.com", "player", "strong-pass-123")
	if !errors.Is(err, account.ErrEmailDelivery) {
		t.Fatalf("expected email delivery error, got %v", err)
	}
	if repository.removedID != repository.pendingID {
		t.Fatalf("expected failed registration %d to be removed, got %d", repository.pendingID, repository.removedID)
	}
}

func TestVerifyNormalizesEmailAndDelegates(t *testing.T) {
	repository := &registrationRepositoryStub{}
	registration := NewAccountRegistration(repository, &verificationEmailSenderStub{}, passwordHasherStub{})
	before := time.Now()

	if err := registration.Verify(context.Background(), " Player@Example.com ", "012345"); err != nil {
		t.Fatalf("verify registration: %v", err)
	}
	if repository.verifiedEmail != "player@example.com" || repository.verifiedCode != "012345" {
		t.Fatalf("unexpected verification request: email=%q code=%q", repository.verifiedEmail, repository.verifiedCode)
	}
	if repository.verifiedAt.Before(before) {
		t.Fatalf("verification used an unexpected timestamp: %v", repository.verifiedAt)
	}
}
