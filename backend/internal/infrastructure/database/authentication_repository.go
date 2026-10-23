package database

import (
	"context"
	"database/sql"
	"fmt"
)

type AuthenticationRepository struct {
	db *sql.DB
}

func NewAuthenticationRepository(db *sql.DB) *AuthenticationRepository {
	return &AuthenticationRepository{db: db}
}

func (r *AuthenticationRepository) PasswordHashByAccountName(ctx context.Context, accountName string) (string, error) {
	var passwordHash string
	if err := r.db.QueryRowContext(ctx,
		"SELECT password_hash FROM users WHERE username = ?",
		accountName,
	).Scan(&passwordHash); err != nil {
		return "", fmt.Errorf("load account password hash: %w", err)
	}
	return passwordHash, nil
}
