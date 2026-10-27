package database

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gomoku-backend/internal/domain"
)

type AccountRecoveryRepository struct {
	db *sql.DB
}

func NewAccountRecoveryRepository(db *sql.DB) *AccountRecoveryRepository {
	return &AccountRecoveryRepository{db: db}
}

func (r *AccountRecoveryRepository) SaveRecoveryCode(
	ctx context.Context,
	email, codeHash string,
	expiresAt time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin account recovery request: %w", err)
	}
	defer tx.Rollback()

	var userID int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE email = ?", email).Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return fmt.Errorf("find account for recovery: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM account_recoveries WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("replace account recovery request: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO account_recoveries
			(user_id, verification_code_hash, expires_at)
			VALUES (?, ?, ?)`,
		userID, codeHash, expiresAt,
	); err != nil {
		return fmt.Errorf("save account recovery code: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit account recovery request: %w", err)
	}
	return nil
}

func (r *AccountRecoveryRepository) RemoveRecovery(ctx context.Context, email, codeHash string) error {
	if _, err := r.db.ExecContext(ctx,
		`DELETE recovery FROM account_recoveries recovery
		 JOIN users ON users.id = recovery.user_id
		 WHERE users.email = ? AND recovery.verification_code_hash = ?`,
		email, codeHash,
	); err != nil {
		return fmt.Errorf("remove account recovery request: %w", err)
	}
	return nil
}

func (r *AccountRecoveryRepository) IssueResetToken(
	ctx context.Context,
	email, code, tokenHash string,
	now time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin account recovery verification: %w", err)
	}
	defer tx.Rollback()

	var userID int64
	var storedCodeHash string
	var attempts uint8
	var expiresAt time.Time
	err = tx.QueryRowContext(ctx,
		`SELECT users.id, recovery.verification_code_hash, recovery.verification_attempts, recovery.expires_at
		 FROM account_recoveries recovery
		 JOIN users ON users.id = recovery.user_id
		 WHERE users.email = ?
		 FOR UPDATE`,
		email,
	).Scan(&userID, &storedCodeHash, &attempts, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrExpiredCode
	}
	if err != nil {
		return fmt.Errorf("load account recovery request: %w", err)
	}

	pending := domain.PendingRegistration{
		VerificationCodeHash: storedCodeHash,
		VerificationAttempts: attempts,
		ExpiresAt:            expiresAt,
	}
	if err := pending.Verify(code, now); err != nil {
		if errors.Is(err, domain.ErrExpiredCode) || errors.Is(err, domain.ErrTooManyAttempts) {
			if _, deleteErr := tx.ExecContext(ctx, "DELETE FROM account_recoveries WHERE user_id = ?", userID); deleteErr != nil {
				return fmt.Errorf("remove expired account recovery: %w", deleteErr)
			}
		} else if _, updateErr := tx.ExecContext(ctx,
			"UPDATE account_recoveries SET verification_attempts = ? WHERE user_id = ?",
			pending.VerificationAttempts, userID,
		); updateErr != nil {
			return fmt.Errorf("record invalid recovery code attempt: %w", updateErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return fmt.Errorf("commit account recovery verification failure: %w", commitErr)
		}
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE account_recoveries
		 SET reset_token_hash = ?, reset_expires_at = ?
		 WHERE user_id = ?`,
		tokenHash, now.Add(domain.VerificationCodeLifetime), userID,
	); err != nil {
		return fmt.Errorf("save account recovery token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit account recovery verification: %w", err)
	}
	return nil
}

func (r *AccountRecoveryRepository) ValidateResetToken(
	ctx context.Context,
	email, tokenHash string,
	now time.Time,
) error {
	var storedTokenHash sql.NullString
	var tokenExpiresAt sql.NullTime
	err := r.db.QueryRowContext(ctx,
		`SELECT recovery.reset_token_hash, recovery.reset_expires_at
		 FROM account_recoveries recovery
		 JOIN users ON users.id = recovery.user_id
		 WHERE users.email = ?`,
		email,
	).Scan(&storedTokenHash, &tokenExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrInvalidRecoveryToken
	}
	if err != nil {
		return fmt.Errorf("load account recovery token: %w", err)
	}
	if !storedTokenHash.Valid || !tokenExpiresAt.Valid ||
		subtle.ConstantTimeCompare([]byte(storedTokenHash.String), []byte(tokenHash)) != 1 ||
		!now.Before(tokenExpiresAt.Time) {
		return domain.ErrInvalidRecoveryToken
	}
	return nil
}

func (r *AccountRecoveryRepository) CompleteRecovery(
	ctx context.Context,
	email, tokenHash, accountName, passwordHash string,
	now time.Time,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin account recovery completion: %w", err)
	}
	defer tx.Rollback()

	var userID int64
	var storedTokenHash sql.NullString
	var tokenExpiresAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT users.id, recovery.reset_token_hash, recovery.reset_expires_at
		 FROM account_recoveries recovery
		 JOIN users ON users.id = recovery.user_id
		 WHERE users.email = ?
		 FOR UPDATE`,
		email,
	).Scan(&userID, &storedTokenHash, &tokenExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrInvalidRecoveryToken
		}
		return fmt.Errorf("load account recovery token: %w", err)
	}
	if !storedTokenHash.Valid || !tokenExpiresAt.Valid ||
		subtle.ConstantTimeCompare([]byte(storedTokenHash.String), []byte(tokenHash)) != 1 ||
		!now.Before(tokenExpiresAt.Time) {
		return domain.ErrInvalidRecoveryToken
	}

	if _, err := tx.ExecContext(ctx,
		"UPDATE users SET username = ?, password_hash = ?, auth_version = auth_version + 1 WHERE id = ?",
		accountName, passwordHash, userID,
	); err != nil {
		if isDuplicateKey(err) {
			return domain.ErrAccountExists
		}
		return fmt.Errorf("update recovered account: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM account_recoveries WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("consume account recovery token: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit recovered account: %w", err)
	}
	return nil
}
