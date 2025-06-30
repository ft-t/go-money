package rules

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package rules_test -source=interfaces.go

type Interpreter interface {
	Run(
		_ context.Context,
		script string,
		clonedTx *database.Transaction,
	) (bool, error)
}

type MapperSvc interface {
	MapRule(rule *database.Rule) *gomoneypbv1.Rule
	MapTransaction(ctx context.Context, tx *database.Transaction) *gomoneypbv1.Transaction
}

type AccountsSvc interface {
	GetAccountByID(
		ctx context.Context,
		accountId int32,
	) (*database.Account, error)
}

type ExecutorSvc interface {
	ProcessSingleRule(
		ctx context.Context,
		inputTx *database.Transaction,
		rule *database.Rule,
	) (bool, *database.Transaction, error)
}

type TransactionSvc interface {
	GetTransactionByIDs(
		ctx context.Context,
		ids []int64,
	) ([]*database.Transaction, error)

	ValidateTransaction(
		ctx context.Context,
		dbTx *gorm.DB,
		tx *database.Transaction,
	) error
}

type CurrencyConverterSvc interface {
	Convert(
		ctx context.Context,
		fromCurrency string,
		toCurrency string,
		amount decimal.Decimal,
	) (decimal.Decimal, error)
}

type DecimalSvc interface {
	GetCurrencyDecimals(ctx context.Context, currency string) int32
}
