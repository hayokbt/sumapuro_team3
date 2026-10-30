package accountrecovery

import "context"

type Service interface {
	RequestCode(ctx context.Context, email string) error
	VerifyCode(ctx context.Context, email, code string) (string, error)
	ResetAccount(ctx context.Context, email, token, accountName, password string) error
}
