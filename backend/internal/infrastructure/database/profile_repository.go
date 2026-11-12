package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gomoku-backend/internal/domain/account"
)

type ProfileRepository struct{ db *sql.DB }

func NewProfileRepository(db *sql.DB) *ProfileRepository { return &ProfileRepository{db} }
func (r *ProfileRepository) Get(ctx context.Context, id int64) (account.Profile, error) {
	var p account.Profile
	err := r.db.QueryRowContext(ctx, `SELECT u.id,u.username,u.email,u.password_hash,u.bgm_volume,u.se_volume,
 COALESCE(s.wins,0),COALESCE(s.losses,0),COALESCE(s.draws,0),COALESCE(s.rating_score,1000)
 FROM users u LEFT JOIN user_scores s ON s.user_id=u.id WHERE u.id=?`, id).Scan(&p.UserID, &p.AccountName, &p.Email, &p.PasswordHash, &p.BGMVolume, &p.SEVolume, &p.Wins, &p.Losses, &p.Draws, &p.Rating)
	if errors.Is(err, sql.ErrNoRows) {
		return p, account.ErrUnauthenticated
	}
	return p, err
}
func (r *ProfileRepository) Update(ctx context.Context, id int64, name, oldHash, newHash string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET username=?,auth_version=auth_version+IF(password_hash<>?,1,0),password_hash=? WHERE id=? AND password_hash=?`, name, newHash, newHash, id, oldHash)
	if err != nil {
		var duplicate *mysql.MySQLError
		if errors.As(err, &duplicate) && duplicate.Number == 1062 {
			return account.ErrAccountExists
		}
		return err
	}
	// MySQL reports zero changed rows for a no-op; still distinguish a concurrent reset.
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var exists bool
		if err = r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=? AND password_hash=?)`, id, oldHash).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return account.ErrUnauthenticated
		}
	}
	return nil
}
func (r *ProfileRepository) SaveSettings(ctx context.Context, id int64, s account.Settings) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET bgm_volume=?,se_volume=? WHERE id=?`, s.BGMVolume, s.SEVolume, id)
	return err
}
