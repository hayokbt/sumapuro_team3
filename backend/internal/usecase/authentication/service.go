package authentication

import "context"

type Service interface {
	Authenticate(ctx context.Context, accountName, password string) error
}
