package maintenance

import (
	"context"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
)

func (s *Service) FixDailyGaps(
	ctx context.Context,
) error {
	var accounts []*database.Account

	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()
	
	if err := tx.Find(&accounts).Error; err != nil {
		return err
	}

	for _, account := range accounts {
		var latestDailyStat database.DailyStat

		if err := tx.Where("account_id = ?", account.ID).
			Order("date DESC").Limit(1).Find(&latestDailyStat).Error; err != nil {
			return err
		}

		startDate := latestDailyStat.Date

		if latestDailyStat.Date.IsZero() { // should not happen, but just in case
			var fistTransaction database.Transaction

			if err := tx.Where("(source_account_id = ? or destination_account_id = ?) AND deleted_at IS NULL", account.ID, account.ID).
				Order("transaction_date_time ASC").
				Limit(1).Find(&fistTransaction).Error; err != nil {
				return err
			}

			if !fistTransaction.TransactionDateTime.IsZero() {
				startDate = fistTransaction.TransactionDateTime
			} else {
				startDate = account.CreatedAt // event more strange, but ok
			}
		}

		if err := s.cfg.StatsSvc.CalculateDailyStat(ctx, tx, transactions.CalculateDailyStatRequest{
			StartDate: startDate,
			AccountID: account.ID,
		}); err != nil {
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
