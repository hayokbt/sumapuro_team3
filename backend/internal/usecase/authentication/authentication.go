package authentication

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gomoku-backend/internal/domain"
)

type PasswordVerifier interface {
	Verify(password, encodedHash string) (bool, error)
}

type Authentication struct {
	repository domain.AuthenticationRepository
	verifier   PasswordVerifier
}

func NewAuthentication(repository domain.AuthenticationRepository, verifier PasswordVerifier) *Authentication {
	return &Authentication{
		repository: repository,
		verifier:   verifier,
	}
}

func (a *Authentication) Authenticate(ctx context.Context, accountName, password string) error {
	accountName = strings.TrimSpace(accountName)
	if accountName == "" || password == "" {
		return domain.ErrInvalidCredentials
	}

	passwordHash, err := a.repository.PasswordHashByAccountName(ctx, accountName)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrInvalidCredentials
	}
	if err != nil {
		return fmt.Errorf("find account for authentication: %w", err)
	}

	matches, err := a.verifier.Verify(password, passwordHash)
	if err != nil {
		return fmt.Errorf("verify account password: %w", err)
	}
	if !matches {
		return domain.ErrInvalidCredentials
	}
	return nil
}
