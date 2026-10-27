package database

import (
	"context"
	"database/sql"
	"fmt"

	"gomoku-backend/internal/domain"
)

type AuthenticationRepository struct {
	db *sql.DB
}

func NewAuthenticationRepository(db *sql.DB) *AuthenticationRepository {
	return &AuthenticationRepository{db: db}
}

func (r *AuthenticationRepository) PasswordHashByAccountName(ctx context.Context, accountName string) (domain.AuthenticatedUser, error) {
	var user domain.AuthenticatedUser
	if err := r.db.QueryRowContext(ctx,
		"SELECT id, auth_version, password_hash FROM users WHERE username = ?",
		accountName,
	).Scan(&user.ID, &user.Version, &user.PasswordHash); err != nil {
		return user, fmt.Errorf("load account password hash: %w", err)
	}
	return user, nil
}
