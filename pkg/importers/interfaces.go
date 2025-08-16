package importers

import (
	"context"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package importers_test -source=interfaces.go

type Implementation interface {
	Import(ctx context.Context, req *ImportRequest) (*importv1.ImportTransactionsResponse, error)
	Type() importv1.ImportSource
}

type AccountSvc interface {
	GetAllAccounts(ctx context.Context) ([]*database.Account, error)
}

type TagSvc interface {
	GetAllTags(ctx context.Context) ([]*database.Tag, error)
}

type CategoriesSvc interface {
	GetAllCategories(ctx context.Context) ([]*database.Category, error)
}

type TransactionSvc interface {
	CreateBulkInternal(
		ctx context.Context,
		reqs []*transactions.BulkRequest,
		tx *gorm.DB,
	) ([]*transactionsv1.CreateTransactionResponse, error)
}

type CurrencySvc interface {
	Convert(
		ctx context.Context,
		fromCurrency string,
		toCurrency string,
		amount decimal.Decimal,
	) (decimal.Decimal, error)
}
