package transactions

import (
	"context"
	"github.com/cockroachdb/errors"
	gomoneypbv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/jinzhu/now"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatService struct {
}

func NewStatService() *StatService {
	return &StatService{}
}

func (s *StatService) HandleTransaction(
	ctx context.Context,
	dbTx *gorm.DB,
	newTX *database.Transaction,
) error {
	switch newTX.TransactionType {
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT:
		return s.handleDeposit(ctx, dbTx, newTX)
	default:
		return errors.Newf("unsupported transaction type: %s", newTX.TransactionType)
	}
}

func (s *StatService) handleDeposit(
	_ context.Context,
	tx *gorm.DB,
	newTX *database.Transaction,
) error {
	if newTX.SourceAccountID != nil {
		return errors.New("source account is not allowed for deposit transactions")
	}

	if newTX.DestinationAccountID == nil {
		return errors.New("destination account is required for deposit transactions")
	}

	if newTX.DestinationAmount.LessThan(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}

	txDate := now.New(newTX.TransactionDateTime)

	txDay := txDate.BeginningOfDay()

	dailyStat := &database.DailyStat{
		AccountID: *newTX.DestinationAccountID,
		Date:      txDay,
		Balance:   decimal.Decimal{},
	}

	if err := tx.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(dailyStat).Error; err != nil {
		return err
	}

	if err := tx.Exec("UPDATE daily_stats SET balance = balance + ? WHERE account_id = ? AND date >= ?",
		newTX.DestinationAmount, *newTX.DestinationAccountID, txDay).Error; err != nil {
		return err
	}

	// month stats
	txMonth := txDate.BeginningOfMonth()

	monthStat := &database.MonthlyStat{
		AccountID: *newTX.DestinationAccountID,
		Date:      txMonth,
		Balance:   decimal.Decimal{},
	}
	if err := tx.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(monthStat).Error; err != nil {
		return err
	}

	if err := tx.Exec("UPDATE monthly_stats SET balance = balance + ? WHERE account_id = ? AND date >= ?",
		newTX.DestinationAmount, *newTX.DestinationAccountID, txMonth).Error; err != nil {
		return err
	}

	return nil
}
