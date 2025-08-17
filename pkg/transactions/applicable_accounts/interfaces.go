package applicable_accounts

import (
	"context"

	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package applicable_accounts_test -source=interfaces.go
type AccountSvc interface {
	GetAllAccounts(ctx context.Context) ([]*database.Account, error)
}
