package maintenance

import (
	"context"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"gorm.io/gorm"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package maintenance_test -source=interfaces.go

type StatsSvc interface {
	CalculateDailyStat(
		_ context.Context,
		dbTx *gorm.DB,
		req transactions.CalculateDailyStatRequest,
	) error
}

type TransactionSvc interface {
	StoreStat(
		ctx context.Context,
		tx *gorm.DB,
		created []*database.Transaction,
		originalTxs []*database.Transaction,
		accountMap map[int32]*database.Account,
	) error
}

type AccountSvc interface {
	GetAllAccounts(ctx context.Context) ([]*database.Account, error)
}
