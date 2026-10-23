package authentication

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"gomoku-backend/internal/domain"
)

type repositoryStub struct {
	passwordHash string
	err          error
	accountName  string
}

func (r *repositoryStub) PasswordHashByAccountName(_ context.Context, accountName string) (string, error) {
	r.accountName = accountName
	return r.passwordHash, r.err
}

type verifierStub struct {
	matches bool
	err     error
	called  bool
}

func (v *verifierStub) Verify(_, _ string) (bool, error) {
	v.called = true
	return v.matches, v.err
}

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name            string
		accountName     string
		password        string
		repositoryError error
		matches         bool
		verifierError   error
		wantError       error
		wantVerifier    bool
		wantAccount     string
	}{
		{
			name:         "valid credentials",
			accountName:  " player ",
			password:     "strong-password",
			matches:      true,
			wantVerifier: true,
			wantAccount:  "player",
		},
		{
			name:         "incorrect password",
			accountName:  "player",
			password:     "incorrect",
			wantError:    domain.ErrInvalidCredentials,
			wantVerifier: true,
			wantAccount:  "player",
		},
		{
			name:            "unknown account",
			accountName:     "unknown",
			password:        "password",
			repositoryError: sql.ErrNoRows,
			wantError:       domain.ErrInvalidCredentials,
			wantAccount:     "unknown",
		},
		{
			name:            "repository failure",
			accountName:     "player",
			password:        "password",
			repositoryError: errors.New("database unavailable"),
			wantError:       errors.New("find account for authentication: database unavailable"),
			wantAccount:     "player",
		},
		{
			name:          "verifier failure",
			accountName:   "player",
			password:      "password",
			verifierError: errors.New("invalid stored hash"),
			wantError:     errors.New("verify account password: invalid stored hash"),
			wantVerifier:  true,
			wantAccount:   "player",
		},
		{
			name:        "blank account name",
			accountName: "  ",
			password:    "password",
			wantError:   domain.ErrInvalidCredentials,
		},
		{
			name:        "blank password",
			accountName: "player",
			wantError:   domain.ErrInvalidCredentials,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &repositoryStub{passwordHash: "stored-hash", err: test.repositoryError}
			verifier := &verifierStub{matches: test.matches, err: test.verifierError}
			service := NewAuthentication(repository, verifier)

			err := service.Authenticate(context.Background(), test.accountName, test.password)
			if test.wantError == nil {
				if err != nil {
					t.Fatalf("Authenticate() error = %v, want nil", err)
				}
			} else if err == nil || err.Error() != test.wantError.Error() {
				t.Fatalf("Authenticate() error = %v, want %v", err, test.wantError)
			}
			if verifier.called != test.wantVerifier {
				t.Errorf("verifier called = %v, want %v", verifier.called, test.wantVerifier)
			}
			if repository.accountName != test.wantAccount {
				t.Errorf("repository account name = %q, want %q", repository.accountName, test.wantAccount)
			}
		})
	}
}
