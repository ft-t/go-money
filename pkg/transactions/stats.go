package transactions

import (
	"context"
	"database/sql"
	_ "embed"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"gorm.io/gorm"
)

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

	accounts = append(accounts, tx.SourceAccountID)
	accounts = append(accounts, tx.DestinationAccountID)

	return accounts
}

func (s *StatService) BuildImpactedAccounts(newTxs []*database.Transaction) map[int32]time.Time {
	impactedAccounts := map[int32]time.Time{} // tx with the lowest date

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

	return impactedAccounts
}

func (s *StatService) HandleTransactions(
	ctx context.Context,
	dbTx *gorm.DB,
	newTxs []*database.Transaction,
) error {
	impactedAccounts := s.BuildImpactedAccounts(newTxs)

	if len(impactedAccounts) == 0 {
		return nil // nothing to do
	}

	for accountID, txTime := range impactedAccounts {
		if err := s.CalculateDailyStat(ctx, dbTx, CalculateDailyStatRequest{
			StartDate: txTime,
			AccountID: accountID,
		}); err != nil {
			return errors.Wrap(err, "failed to recalculate")
		}
	}

	return nil
}

func (s *StatService) CalculateDailyStat(
	_ context.Context,
	dbTx *gorm.DB,
	req CalculateDailyStatRequest,
) error {
	return dbTx.Exec(dailyRecalculate,
		sql.Named("startDate", req.StartDate),
		sql.Named("accountID", req.AccountID),
	).Error
}
