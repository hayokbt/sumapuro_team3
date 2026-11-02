package authentication

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gomoku-backend/internal/domain/account"
)

type PasswordVerifier interface {
	Verify(password, encodedHash string) (bool, error)
}

type Authentication struct {
	repository account.AuthenticationRepository
	verifier   PasswordVerifier
}

func NewAuthentication(repository account.AuthenticationRepository, verifier PasswordVerifier) *Authentication {
	return &Authentication{
		repository: repository,
		verifier:   verifier,
	}
}

func (a *Authentication) Authenticate(ctx context.Context, accountName, password string) (account.AuthenticatedUser, error) {
	accountName = strings.TrimSpace(accountName)
	if accountName == "" || password == "" {
		return account.AuthenticatedUser{}, account.ErrInvalidCredentials
	}

	user, err := a.repository.PasswordHashByAccountName(ctx, accountName)
	if errors.Is(err, sql.ErrNoRows) {
		return account.AuthenticatedUser{}, account.ErrInvalidCredentials
	}
	if err != nil {
		return account.AuthenticatedUser{}, fmt.Errorf("find account for authentication: %w", err)
	}

	matches, err := a.verifier.Verify(password, user.PasswordHash)
	if err != nil {
		return account.AuthenticatedUser{}, fmt.Errorf("verify account password: %w", err)
	}
	if !matches {
		return account.AuthenticatedUser{}, account.ErrInvalidCredentials
	}
	return user, nil
}
