package account

import (
	"context"
	"errors"
)

var ErrInvalidCredentials = errors.New("invalid account name or password")

type AuthenticatedUser struct {
	ID           int64
	Version      int64
	PasswordHash string
}

type AuthenticationRepository interface {
	PasswordHashByAccountName(ctx context.Context, accountName string) (AuthenticatedUser, error)
}
