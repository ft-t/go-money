package accounts

import (
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package accounts_test -source=interfaces.go

type MapperSvc interface {
	MapAccount(ctx context.Context, acc *database.Account) *v1.Account
}
