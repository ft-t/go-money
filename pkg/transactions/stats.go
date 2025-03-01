package transactions

import (
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	gomoneypbv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jinzhu/now"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type StatService struct {
	noGapTillTime *expirable.LRU[string, time.Time]
}

func NewStatService() *StatService {
	return &StatService{
		noGapTillTime: expirable.NewLRU[string, time.Time](1000, nil, 10*time.Minute),
	}
}

func (s *StatService) HandleTransaction(
	ctx context.Context,
	dbTx *gorm.DB,
	newTx *database.Transaction,
	impactedAccounts map[int32]*database.Account,
) error {
	switch newTx.TransactionType {
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT:
		return s.handleDeposit(ctx, dbTx, newTx)
	default:
		return errors.Newf("unsupported transaction type: %s", newTx.TransactionType)
	}
}

func (s *StatService) ensureNoGapDaily(
	dbTx *gorm.DB,
	newTx *database.Transaction,
) {
	if newTx.SourceAccountID != nil {

	}
}

func (s *StatService) checkDailyGapForWallet(
	dbTx *gorm.DB,
	accountID int32,
	transactionTime time.Time,
	firstAccountTransactionAt time.Time,
) {
	key := fmt.Sprintf("daily_gap_%d", accountID)

	dateNow := now.New(time.Now().UTC()).EndOfDay()
	targetTime := now.New(transactionTime).EndOfDay()

	if dateNow.After(targetTime) {
		targetTime = dateNow
	}

	cached, ok := s.noGapTillTime.Get(key)

	if ok {
		if time.Now().Before(cached) || targetTime.Equal(cached) { // we are good
			return
		}
	}

	// todo acquire lock

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
