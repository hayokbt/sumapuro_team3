package accountregistration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gomoku-backend/internal/domain/account"
)

type AccountRegistration struct {
	repository account.RegistrationRepository
	sender     account.VerificationEmailSender
	hasher     account.PasswordHasher
	now        func() time.Time
}

func NewAccountRegistration(
	repository account.RegistrationRepository,
	sender account.VerificationEmailSender,
	hasher account.PasswordHasher,
) *AccountRegistration {
	return &AccountRegistration{
		repository: repository,
		sender:     sender,
		hasher:     hasher,
		now:        time.Now,
	}
}

func (r *AccountRegistration) Start(ctx context.Context, email, accountName, password string) error {
	registration, err := account.NewRegistration(email, accountName, password)
	if err != nil {
		return err
	}
	passwordHash, err := r.hasher.Hash(registration.Password)
	if err != nil {
		return fmt.Errorf("hash registration password: %w", err)
	}
	code, err := account.NewVerificationCode()
	if err != nil {
		return err
	}

	pending := account.PendingRegistration{
		Email:                registration.Email,
		AccountName:          registration.AccountName,
		PasswordHash:         passwordHash,
		VerificationCodeHash: account.HashVerificationCode(code),
		ExpiresAt:            r.now().Add(account.VerificationCodeLifetime),
	}
	id, err := r.repository.SavePending(ctx, pending)
	if err != nil {
		return err
	}
	if err := r.sender.SendVerificationCode(registration.Email, code); err != nil {
		if cleanupErr := r.repository.RemovePending(ctx, id); cleanupErr != nil {
			return errors.Join(
				account.ErrEmailDelivery,
				fmt.Errorf("send verification email: %w", err),
				fmt.Errorf("remove unsent registration: %w", cleanupErr),
			)
		}
		return errors.Join(account.ErrEmailDelivery, fmt.Errorf("send verification email: %w", err))
	}
	return nil
}

func (r *AccountRegistration) Verify(ctx context.Context, email, code string) error {
	normalizedEmail, err := account.NormalizeEmail(email)
	if err != nil {
		return err
	}
	return r.repository.CompleteVerification(ctx, normalizedEmail, code, r.now())
}
