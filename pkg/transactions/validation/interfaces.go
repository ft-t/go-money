package validation

import (
	"context"

	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/applicable_accounts"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package validation_test -source=interfaces.go

type AccountSvc interface {
	GetAccountByID(ctx context.Context, id int32) (*database.Account, error)
	GetAllAccounts(ctx context.Context) ([]*database.Account, error)
}

type ApplicableAccountSvc interface {
	GetApplicableAccounts(
		_ context.Context,
		accounts []*database.Account,
	) map[v1.TransactionType]*applicable_accounts.PossibleAccount
}
