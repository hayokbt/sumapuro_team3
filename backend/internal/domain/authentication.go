package domain

import (
	"context"
	"errors"
)

var ErrInvalidCredentials = errors.New("invalid account name or password")

type AuthenticationRepository interface {
	PasswordHashByAccountName(ctx context.Context, accountName string) (string, error)
}
