package database

import (
	"context"
	"database/sql"

	"gomoku-backend/internal/domain/account"
)

type AuthSessionRepository struct{ db *sql.DB }

func NewAuthSessionRepository(db *sql.DB) *AuthSessionRepository { return &AuthSessionRepository{db} }
func (r *AuthSessionRepository) Create(ctx context.Context, s account.AuthSession) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO auth_sessions (id,user_id,expires_at) VALUES (?,?,?)", s.ID, s.UserID, s.ExpiresAt)
	return err
}
func (r *AuthSessionRepository) Valid(ctx context.Context, s account.AuthSession) (bool, error) {
	var valid bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM auth_sessions s JOIN users u ON u.id=s.user_id WHERE s.id=? AND s.user_id=? AND s.revoked_at IS NULL AND s.expires_at>UTC_TIMESTAMP() AND u.auth_version=?)`, s.ID, s.UserID, s.Version).Scan(&valid)
	return valid, err
}
func (r *AuthSessionRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE auth_sessions SET revoked_at=UTC_TIMESTAMP() WHERE id=?", id)
	return err
}
