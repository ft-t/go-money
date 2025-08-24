package maintenance

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
)

type RecalculateService struct {
	cfg *RecalculateServiceConfig
}

type RecalculateServiceConfig struct {
	AccountSvc     AccountSvc
	TransactionSvc TransactionSvc
}

func NewRecalculateService(cfg *RecalculateServiceConfig) *RecalculateService {
	return &RecalculateService{
		cfg: cfg,
	}
}

func (s *RecalculateService) RecalculateAll(
	ctx context.Context,
) error {
	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	accounts, err := s.cfg.AccountSvc.GetAllAccounts(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get accounts")
	}

	accountMap := make(map[int32]*database.Account, len(accounts))
	for _, acc := range accounts {
		accountMap[acc.ID] = acc

		if err = tx.Update("current_balance", 0).Error; err != nil {
			return errors.Wrap(err, "failed to reset account balance")
		}
	}

	tables := []string{
		"double_entries",
		"daily_stat",
	}

	for _, table := range tables {
		if err = tx.Exec("truncate table " + table).Error; err != nil {
			return err
		}
	}

	var txs []*database.Transaction
	if err = tx.Order("transaction_date_time asc").Find(&txs).Error; err != nil {
		return errors.Wrap(err, "failed to get transactions")
	}

	if err = s.cfg.TransactionSvc.StoreStat(ctx, tx, txs, nil, accountMap); err != nil {
		return errors.Wrap(err, "failed to store stats")
	}

	return tx.Commit().Error
}
