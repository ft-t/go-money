package maintenance

import (
	"context"
	"github.com/ft-t/go-money/pkg/transactions"
	"gorm.io/gorm"
)

type StatsSvc interface {
	CalculateDailyStat(
		_ context.Context,
		dbTx *gorm.DB,
		req transactions.CalculateDailyStatRequest,
	) error
}
