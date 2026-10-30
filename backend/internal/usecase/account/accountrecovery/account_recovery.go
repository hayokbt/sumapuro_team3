package accountrecovery

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gomoku-backend/internal/domain/account"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
}

type AccountRecovery struct {
	repository account.AccountRecoveryRepository
	sender     account.RecoveryEmailSender
	hasher     PasswordHasher
	now        func() time.Time
}

func NewAccountRecovery(
	repository account.AccountRecoveryRepository,
	sender account.RecoveryEmailSender,
	hasher PasswordHasher,
) *AccountRecovery {
	return &AccountRecovery{
		repository: repository,
		sender:     sender,
		hasher:     hasher,
		now:        time.Now,
	}
}

func (r *AccountRecovery) RequestCode(ctx context.Context, email string) error {
	normalizedEmail, err := account.NormalizeEmail(email)
	if err != nil {
		return err
	}
	code, err := account.NewVerificationCode()
	if err != nil {
		return fmt.Errorf("generate account recovery code: %w", err)
	}
	codeHash := account.HashVerificationCode(code)
	err = r.repository.SaveRecoveryCode(
		ctx,
		normalizedEmail,
		codeHash,
		r.now().Add(account.VerificationCodeLifetime),
	)
	if errors.Is(err, account.ErrAccountNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("save account recovery code: %w", err)
	}
	if err := r.sender.SendAccountRecoveryCode(normalizedEmail, code); err != nil {
		if cleanupErr := r.repository.RemoveRecovery(ctx, normalizedEmail, codeHash); cleanupErr != nil {
			return errors.Join(
				account.ErrEmailDelivery,
				fmt.Errorf("send account recovery code: %w", err),
				fmt.Errorf("remove unsent account recovery: %w", cleanupErr),
			)
		}
		return errors.Join(account.ErrEmailDelivery, fmt.Errorf("send account recovery code: %w", err))
	}
	return nil
}

func (r *AccountRecovery) VerifyCode(ctx context.Context, email, code string) (string, error) {
	normalizedEmail, err := account.NormalizeEmail(email)
	if err != nil {
		return "", err
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate account recovery token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	if err := r.repository.IssueResetToken(
		ctx,
		normalizedEmail,
		code,
		base64.RawStdEncoding.EncodeToString(tokenHash[:]),
		r.now(),
	); err != nil {
		return "", err
	}
	return token, nil
}

func (r *AccountRecovery) ResetAccount(
	ctx context.Context,
	email, token, accountName, password string,
) error {
	normalizedEmail, err := account.NormalizeEmail(email)
	if err != nil {
		return err
	}
	accountName = strings.TrimSpace(accountName)
	if accountName == "" || utf8.RuneCountInString(accountName) > 50 ||
		len(password) < 8 || len(password) > account.MaxPasswordBytes {
		return account.ErrInvalidRegistration
	}
	if token == "" {
		return account.ErrInvalidRecoveryToken
	}
	tokenHash := sha256.Sum256([]byte(token))
	encodedTokenHash := base64.RawStdEncoding.EncodeToString(tokenHash[:])
	now := r.now()
	if err := r.repository.ValidateResetToken(ctx, normalizedEmail, encodedTokenHash, now); err != nil {
		return err
	}
	passwordHash, err := r.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("hash recovered account password: %w", err)
	}
	if err := r.repository.CompleteRecovery(
		ctx,
		normalizedEmail,
		encodedTokenHash,
		accountName,
		passwordHash,
		now,
	); err != nil {
		return err
	}
	return nil
}
