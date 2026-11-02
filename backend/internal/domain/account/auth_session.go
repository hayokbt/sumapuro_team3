package account

import (
	"context"
	"errors"
	"time"
)

var ErrUnauthenticated = errors.New("unauthenticated")

type AuthSession struct {
	ID        string
	UserID    int64
	Version   int64
	ExpiresAt time.Time
}
type AuthSessionRepository interface {
	Create(context.Context, AuthSession) error
	Valid(context.Context, AuthSession) (bool, error)
	Revoke(context.Context, string) error
}
