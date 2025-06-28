package transactions

import (
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package transactions_test -source=interfaces.go

type StatsSvc interface {
	HandleTransactions(
		ctx context.Context,
		dbTx *gorm.DB,
		newTxs []*database.Transaction,
	) error
}

type LockerSvc interface {
	LockDailyStat(ctx context.Context, dbTx *gorm.DB) error
	LockMonthlyStat(ctx context.Context, dbTx *gorm.DB) error
}

type MapperSvc interface {
	MapTransaction(ctx context.Context, tx *database.Transaction) *v1.Transaction
}

type CurrencyConverterSvc interface {
	Convert(
		ctx context.Context,
		fromCurrency string,
		toCurrency string,
		amount decimal.Decimal,
	) (decimal.Decimal, error)
}

type BaseAmountSvc interface {
	RecalculateAmountInBaseCurrency(
		_ context.Context,
		tx *gorm.DB,
		txs []*database.Transaction,
	) error
}

type RuleSvc interface {
	ProcessTransactions(
		ctx context.Context,
		inputTxs []*database.Transaction,
	) ([]*database.Transaction, error)
}
