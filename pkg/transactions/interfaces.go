package transactions

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package transactions_test -source=interfaces.go

type StatsSvc interface {
	ProcessTransaction(
		ctx context.Context,
		dbTx *gorm.DB,
		transaction *database.Transaction,
	) error
}

type LockerSvc interface {
	LockDailyStat(ctx context.Context, dbTx *gorm.DB) error
	LockMonthlyStat(ctx context.Context, dbTx *gorm.DB) error
}
