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
	"gorm.io/gorm"
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
	_ context.Context,
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

		if txDay.Before(lowestDate) {
			lowestDate = txDay
		}
	}

	for accountID, txTime := range impactedAccounts {
		if err := dbTx.Exec(dailyRecalculate,
			sql.Named("startDate", txTime),
			sql.Named("accountID", accountID),
		).Error; err != nil {
			return errors.Wrap(err, "failed to recalculate")
		}
	}

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
