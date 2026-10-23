package accountrecovery

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"gomoku-backend/internal/domain"
)

type recoveryRepositoryStub struct {
	email            string
	codeHash         string
	expiresAt        time.Time
	removeCodeHash   string
	tokenHash        string
	validateErr      error
	resetAccountName string
	passwordHash     string
	now              time.Time
	saveErr          error
	removeErr        error
	issueErr         error
	completeErr      error
}

func (r *recoveryRepositoryStub) SaveRecoveryCode(
	_ context.Context,
	email, codeHash string,
	expiresAt time.Time,
) error {
	r.email = email
	r.codeHash = codeHash
	r.expiresAt = expiresAt
	return r.saveErr
}

func (r *recoveryRepositoryStub) RemoveRecovery(_ context.Context, email, codeHash string) error {
	r.email = email
	r.removeCodeHash = codeHash
	return r.removeErr
}

func (r *recoveryRepositoryStub) IssueResetToken(
	_ context.Context,
	email, _, tokenHash string,
	now time.Time,
) error {
	r.email = email
	r.tokenHash = tokenHash
	r.now = now
	return r.issueErr
}

func (r *recoveryRepositoryStub) ValidateResetToken(
	_ context.Context,
	email, tokenHash string,
	now time.Time,
) error {
	r.email = email
	r.tokenHash = tokenHash
	r.now = now
	return r.validateErr
}

func (r *recoveryRepositoryStub) CompleteRecovery(
	_ context.Context,
	email, tokenHash, accountName, passwordHash string,
	now time.Time,
) error {
	r.email = email
	r.tokenHash = tokenHash
	r.resetAccountName = accountName
	r.passwordHash = passwordHash
	r.now = now
	return r.completeErr
}

type recoveryEmailSenderStub struct {
	email string
	code  string
	err   error
}

func (s *recoveryEmailSenderStub) SendAccountRecoveryCode(email, code string) error {
	s.email = email
	s.code = code
	return s.err
}

type recoveryPasswordHasherStub struct {
	password string
	err      error
}

func (h *recoveryPasswordHasherStub) Hash(password string) (string, error) {
	h.password = password
	if h.err != nil {
		return "", h.err
	}
	return "hashed-recovery-password", nil
}

func TestRequestCodeSavesAndSendsCode(t *testing.T) {
	repository := &recoveryRepositoryStub{}
	sender := &recoveryEmailSenderStub{}
	service := NewAccountRecovery(repository, sender, &recoveryPasswordHasherStub{})
	now := time.Now()
	service.now = func() time.Time { return now }

	if err := service.RequestCode(context.Background(), " Player@Example.com "); err != nil {
		t.Fatalf("RequestCode() error = %v", err)
	}
	if repository.email != "player@example.com" || sender.email != repository.email {
		t.Fatalf("email was not normalized: repository=%q sender=%q", repository.email, sender.email)
	}
	if len(sender.code) != 6 || repository.codeHash != domain.HashVerificationCode(sender.code) {
		t.Fatalf("recovery code was not saved as a hash: sent=%q hash=%q", sender.code, repository.codeHash)
	}
	if !repository.expiresAt.Equal(now.Add(domain.VerificationCodeLifetime)) {
		t.Errorf("expiry = %v, want %v", repository.expiresAt, now.Add(domain.VerificationCodeLifetime))
	}
}

func TestRequestCodeDoesNotRevealUnknownAccount(t *testing.T) {
	repository := &recoveryRepositoryStub{saveErr: domain.ErrAccountNotFound}
	sender := &recoveryEmailSenderStub{}
	service := NewAccountRecovery(repository, sender, &recoveryPasswordHasherStub{})

	if err := service.RequestCode(context.Background(), "player@example.com"); err != nil {
		t.Fatalf("RequestCode() error = %v, want nil for unknown account", err)
	}
	if sender.email != "" {
		t.Errorf("sent recovery email to unknown account %q", sender.email)
	}
}

func TestRequestCodeRemovesUnsentCode(t *testing.T) {
	repository := &recoveryRepositoryStub{}
	sender := &recoveryEmailSenderStub{err: errors.New("SMTP unavailable")}
	service := NewAccountRecovery(repository, sender, &recoveryPasswordHasherStub{})

	err := service.RequestCode(context.Background(), "player@example.com")
	if !errors.Is(err, domain.ErrEmailDelivery) {
		t.Fatalf("RequestCode() error = %v, want email delivery error", err)
	}
	if repository.removeCodeHash != domain.HashVerificationCode(sender.code) {
		t.Error("did not remove the recovery code whose email failed")
	}
}

func TestVerifyCodeIssuesResetTokenHash(t *testing.T) {
	repository := &recoveryRepositoryStub{}
	service := NewAccountRecovery(repository, &recoveryEmailSenderStub{}, &recoveryPasswordHasherStub{})
	now := time.Now()
	service.now = func() time.Time { return now }

	token, err := service.VerifyCode(context.Background(), " Player@Example.com ", "012345")
	if err != nil {
		t.Fatalf("VerifyCode() error = %v", err)
	}
	tokenBytes, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(tokenBytes) != 32 {
		t.Fatalf("reset token is not a 32-byte URL-safe token: %q, error=%v", token, err)
	}
	expectedHash := sha256.Sum256([]byte(token))
	if repository.email != "player@example.com" ||
		repository.tokenHash != base64.RawStdEncoding.EncodeToString(expectedHash[:]) {
		t.Error("repository did not receive the normalized email and hashed reset token")
	}
	if !repository.now.Equal(now) {
		t.Errorf("verification time = %v, want %v", repository.now, now)
	}
}

func TestResetAccountValidatesAndUpdatesCredentials(t *testing.T) {
	repository := &recoveryRepositoryStub{}
	hasher := &recoveryPasswordHasherStub{}
	service := NewAccountRecovery(repository, &recoveryEmailSenderStub{}, hasher)
	now := time.Now()
	service.now = func() time.Time { return now }
	token := "one-time-reset-token"

	if err := service.ResetAccount(
		context.Background(),
		" Player@Example.com ",
		token,
		" new-player ",
		"new-strong-password",
	); err != nil {
		t.Fatalf("ResetAccount() error = %v", err)
	}
	expectedHash := sha256.Sum256([]byte(token))
	if repository.email != "player@example.com" ||
		repository.tokenHash != base64.RawStdEncoding.EncodeToString(expectedHash[:]) ||
		repository.resetAccountName != "new-player" ||
		repository.passwordHash != "hashed-recovery-password" ||
		hasher.password != "new-strong-password" ||
		!repository.now.Equal(now) {
		t.Errorf("unexpected recovery update: %+v", repository)
	}
}

func TestResetAccountRejectsInvalidTokenBeforeHashing(t *testing.T) {
	repository := &recoveryRepositoryStub{validateErr: domain.ErrInvalidRecoveryToken}
	hasher := &recoveryPasswordHasherStub{}
	service := NewAccountRecovery(repository, &recoveryEmailSenderStub{}, hasher)

	err := service.ResetAccount(
		context.Background(),
		"player@example.com",
		"invalid-token",
		"new-player",
		"new-strong-password",
	)
	if !errors.Is(err, domain.ErrInvalidRecoveryToken) {
		t.Fatalf("ResetAccount() error = %v, want invalid recovery token", err)
	}
	if hasher.password != "" || repository.resetAccountName != "" {
		t.Error("invalid recovery token was passed to the hasher or completion repository")
	}
}

func TestResetAccountRejectsInvalidCredentialsBeforeHashing(t *testing.T) {
	repository := &recoveryRepositoryStub{}
	hasher := &recoveryPasswordHasherStub{}
	service := NewAccountRecovery(repository, &recoveryEmailSenderStub{}, hasher)

	err := service.ResetAccount(context.Background(), "player@example.com", "token", "", "short")
	if !errors.Is(err, domain.ErrInvalidRegistration) {
		t.Fatalf("ResetAccount() error = %v, want invalid registration", err)
	}
	if hasher.password != "" || repository.resetAccountName != "" {
		t.Error("invalid credentials were passed to the hasher or repository")
	}
}
