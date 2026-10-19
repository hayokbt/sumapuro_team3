package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gomoku-backend/internal/domain"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type RegistrationRepository struct {
	db *sql.DB
}

func NewRegistrationRepository(db *sql.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) SavePending(ctx context.Context, pending domain.PendingRegistration) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin pending registration transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = ? OR username = ?)",
		pending.Email, pending.AccountName,
	).Scan(&exists); err != nil {
		return 0, fmt.Errorf("check existing account: %w", err)
	}
	if exists {
		return 0, domain.ErrAccountExists
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM pending_registrations WHERE expires_at <= ?", time.Now()); err != nil {
		return 0, fmt.Errorf("remove expired registrations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM pending_registrations WHERE email = ?", pending.Email); err != nil {
		return 0, fmt.Errorf("replace pending registration: %w", err)
	}
	result, err := tx.ExecContext(ctx,
		`INSERT INTO pending_registrations
			(username, email, password_hash, verification_code_hash, expires_at)
			VALUES (?, ?, ?, ?, ?)`,
		pending.AccountName, pending.Email, pending.PasswordHash,
		pending.VerificationCodeHash, pending.ExpiresAt,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return 0, domain.ErrAccountExists
		}
		return 0, fmt.Errorf("save pending registration: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get pending registration ID: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit pending registration: %w", err)
	}
	return id, nil
}

func (r *RegistrationRepository) RemovePending(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM pending_registrations WHERE id = ?", id); err != nil {
		return fmt.Errorf("remove pending registration: %w", err)
	}
	return nil
}

func (r *RegistrationRepository) CompleteVerification(ctx context.Context, email, code string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin verification transaction: %w", err)
	}
	defer tx.Rollback()

	pending, err := loadPendingRegistration(ctx, tx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrExpiredCode
	}
	if err != nil {
		return err
	}
	if err := pending.Verify(code, now); err != nil {
		if errors.Is(err, domain.ErrExpiredCode) || errors.Is(err, domain.ErrTooManyAttempts) {
			if _, deleteErr := tx.ExecContext(ctx, "DELETE FROM pending_registrations WHERE id = ?", pending.ID); deleteErr != nil {
				return fmt.Errorf("remove invalid pending registration: %w", deleteErr)
			}
		} else if _, updateErr := tx.ExecContext(ctx,
			"UPDATE pending_registrations SET verification_attempts = ? WHERE id = ?",
			pending.VerificationAttempts, pending.ID,
		); updateErr != nil {
			return fmt.Errorf("record failed verification attempt: %w", updateErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("commit failed verification: %w", commitErr)
		}
		return err
	}

	result, err := tx.ExecContext(ctx,
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		pending.AccountName, pending.Email, pending.PasswordHash,
	)
	if err != nil {
		if isDuplicateKey(err) {
			return domain.ErrAccountExists
		}
		return fmt.Errorf("create account: %w", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get created account ID: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO user_scores (user_id) VALUES (?)", userID); err != nil {
		return fmt.Errorf("create account score: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM pending_registrations WHERE id = ?", pending.ID); err != nil {
		return fmt.Errorf("remove verified registration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit account creation: %w", err)
	}
	return nil
}

type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func loadPendingRegistration(ctx context.Context, queryer queryRower, email string) (domain.PendingRegistration, error) {
	var pending domain.PendingRegistration
	err := queryer.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, verification_code_hash, verification_attempts, expires_at
		 FROM pending_registrations WHERE email = ? FOR UPDATE`,
		email,
	).Scan(
		&pending.ID, &pending.AccountName, &pending.Email, &pending.PasswordHash,
		&pending.VerificationCodeHash, &pending.VerificationAttempts, &pending.ExpiresAt,
	)
	if err != nil {
		return domain.PendingRegistration{}, fmt.Errorf("load pending registration: %w", err)
	}
	return pending, nil
}

func isDuplicateKey(err error) bool {
	var mysqlError *mysqlDriver.MySQLError
	return errors.As(err, &mysqlError) && mysqlError.Number == 1062
}
