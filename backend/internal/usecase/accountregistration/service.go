package accountregistration

import "context"

type Service interface {
	Start(ctx context.Context, email, accountName, password string) error
	Verify(ctx context.Context, email, code string) error
}
