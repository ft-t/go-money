package validation

import (
	"context"

	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package validation_test -source=interfaces.go

type AccountSvc interface {
	GetAccountByID(ctx context.Context, id int32) (*database.Account, error)
	GetAllAccounts(ctx context.Context) ([]*database.Account, error)
}
