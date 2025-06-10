package transactions

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jinzhu/now"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

//go:embed scripts/daily_gap_detect.sql
var dailyGapDetect string

//go:embed scripts/daily_recalculate.sql
var dailyRecalculate string

type StatService struct {
	noGapTillTime *expirable.LRU[string, time.Time]
}

func NewStatService() *StatService {
	return &StatService{
		noGapTillTime: expirable.NewLRU[string, time.Time](1000, nil, 10*time.Minute),
	}
}

func (s *StatService) getAccountsForTx(tx *database.Transaction) []int32 {
	var accounts []int32

	if tx.SourceAccountID != nil {
		accounts = append(accounts, *tx.SourceAccountID)
	}
	if tx.DestinationAccountID != nil {
		accounts = append(accounts, *tx.DestinationAccountID)
	}

	return accounts
}

func (s *StatService) HandleTransactions(
	ctx context.Context,
	dbTx *gorm.DB,
	newTxs []*database.Transaction,
) error {
	impactedAccounts := map[int32]time.Time{} // tx with lowest date

	for _, newTx := range newTxs {
		for _, accountID := range s.getAccountsForTx(newTx) {
			if rec, ok := impactedAccounts[accountID]; !ok {
				impactedAccounts[accountID] = newTx.TransactionDateTime
			} else {
				if rec.After(newTx.TransactionDateTime) {
					impactedAccounts[accountID] = newTx.TransactionDateTime
				}
			}
		}
	}

	if len(impactedAccounts) == 0 {
		return nil // nothing to do
	}

	lowestDate := newTxs[0].TransactionDateTime

	for _, newTx := range newTxs {
		txDate := now.New(newTx.TransactionDateTime)
		txDay := txDate.BeginningOfDay()
		//txMonth := txDate.BeginningOfMonth()

		if txDay.Before(lowestDate) {
			lowestDate = txDay
		}

		accountsForTx := s.getAccountsForTx(newTx)

		for _, accountID := range accountsForTx {
			dailyStat := &database.DailyStat{
				AccountID: accountID,
				Date:      txDay,
				Balance:   decimal.Decimal{},
			}

			fmt.Println(dailyStat)

			//if err := dbTx.Clauses(clause.OnConflict{
			//	DoNothing: true,
			//}).Create(dailyStat).Error; err != nil {
			//	return err
			//}

			//monthStat := &database.MonthlyStat{
			//	AccountID: accountID,
			//	Date:      txMonth,
			//	Balance:   decimal.Decimal{},
			//}

			//if err := dbTx.Clauses(clause.OnConflict{
			//	DoNothing: true,
			//}).Create(monthStat).Error; err != nil {
			//	return err
			//}
		}
	}

	for accountID, txTime := range impactedAccounts {
		if err := dbTx.Debug().Exec(dailyRecalculate,
			sql.Named("startDate", txTime),
			sql.Named("accountID", accountID),
		).Error; err != nil {
			return errors.Wrap(err, "failed to recalculate")
		}
	}
	
	return nil
}

func (s *StatService) EnsureNoGapDaily(
	dbTx *gorm.DB,
	newTx *database.Transaction,
	accounts map[int32]*database.Account,
) error {
	//if newTx.SourceAccountID != nil {
	//	if err := s.CheckDailyGapForAccount(dbTx, *newTx.SourceAccountID,
	//		newTx.TransactionDateOnly, *accounts[*newTx.SourceAccountID].FirstTransactionAt); err != nil { // todo
	//
	//	}
	//}

	return nil
}

func (s *StatService) CheckDailyGapForAccount(
	dbTx *gorm.DB,
	accountID int32,
	transactionTime time.Time,
	firstAccountTransactionAt time.Time,
) (GapMeta, error) {
	key := fmt.Sprintf("daily_gap_%d", accountID)

	dateNow := now.New(time.Now().UTC()).EndOfDay()   // 05.11.2025
	targetTime := now.New(transactionTime).EndOfDay() // 05.11.2020

	if dateNow.After(targetTime) {
		targetTime = dateNow
	}

	cached, ok := s.noGapTillTime.Get(key)

	if ok {
		if time.Now().Before(cached) || targetTime.Equal(cached) { // we are good
			return GapMeta{
				FromCache: true,
			}, nil
		}
	}

	var gap []struct {
		Rec int32
	}

	if err := dbTx.Debug().Raw(dailyGapDetect,
		firstAccountTransactionAt,
		targetTime,
		accountID,
	).Find(&gap).Error; err != nil {
		return GapMeta{}, err
	}

	if len(gap) == 0 {
		return GapMeta{}, nil
	}

	fromCache := true
	if targetTime.After(cached) {
		cached = targetTime
		fromCache = false
	}

	// todo acquire lock

	return GapMeta{
		KeysToSet: map[string]time.Time{
			key: cached,
		},
		FromCache: fromCache,
	}, nil
}

type GapMeta struct {
	KeysToSet map[string]time.Time
	FromCache bool
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
