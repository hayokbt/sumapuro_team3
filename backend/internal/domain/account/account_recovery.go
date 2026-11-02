package account

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrInvalidRecoveryToken = errors.New("invalid or expired recovery token")
)

type AccountRecoveryRepository interface {
	SaveRecoveryCode(ctx context.Context, email, codeHash string, expiresAt time.Time) error
	RemoveRecovery(ctx context.Context, email, codeHash string) error
	IssueResetToken(ctx context.Context, email, code, tokenHash string, now time.Time) error
	ValidateResetToken(ctx context.Context, email, tokenHash string, now time.Time) error
	CompleteRecovery(ctx context.Context, email, tokenHash, accountName, passwordHash string, now time.Time) error
}

type RecoveryEmailSender interface {
	SendAccountRecoveryCode(to, code string) error
}
