package authentication

import "context"
import "gomoku-backend/internal/domain"

type Service interface {
	Authenticate(ctx context.Context, accountName, password string) (domain.AuthenticatedUser, error)
}
