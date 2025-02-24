package accounts

import (
	"context"
	v1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
)

type MapperSvc interface {
	MapAccount(ctx context.Context, acc *database.Account) *v1.Account
}
