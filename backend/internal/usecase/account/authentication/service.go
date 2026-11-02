package authentication

import "context"
import "gomoku-backend/internal/domain/account"

type Service interface {
	Authenticate(ctx context.Context, accountName, password string) (account.AuthenticatedUser, error)
}
