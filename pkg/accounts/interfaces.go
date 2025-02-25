package accounts

import (
	"context"
	v1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package accounts_test -source=interfaces.go

type MapperSvc interface {
	MapAccount(ctx context.Context, acc *database.Account) *v1.Account
}
