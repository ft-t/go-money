package rules

import (
	"context"
	"time"

	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
)

type DryRun struct {
	cfg *DryRunConfig
}

type DryRunConfig struct {
	Executor       ExecutorSvc
	TransactionSvc TransactionSvc
	MapperSvc      MapperSvc
	ValidationSvc  ValidationSvc
	AccountSvc     AccountSvc
}

func NewDryRun(
	cfg *DryRunConfig,
) *DryRun {
	return &DryRun{
		cfg: cfg,
	}
}

func (s *DryRun) DryRunRule(ctx context.Context, req *rulesv1.DryRunRuleRequest) (*rulesv1.DryRunRuleResponse, error) {
	var tx *database.Transaction
	if req.TransactionId != 0 {
		dbRecord, err := s.cfg.TransactionSvc.GetTransactionByIDs(ctx, []int64{req.TransactionId})
		if err != nil {
			return nil, err
		}

		if len(dbRecord) != 1 {
			return nil, errors.New("transaction not found")
		}

		tx = dbRecord[0]
	} else {
		tx = &database.Transaction{
			CreatedAt: time.Now().UTC(),
		} // scheduled transaction
	}

	finalResp := &rulesv1.DryRunRuleResponse{
		Before:      s.cfg.MapperSvc.MapTransaction(ctx, tx),
		After:       nil,
		RuleApplied: false,
	}

	executed, updated, ruleErr := s.cfg.Executor.ProcessSingleRule(ctx, tx, &database.Rule{
		Script: req.Rule.Script,
		Title:  req.Rule.Title,
	})
	if ruleErr != nil {
		return nil, ruleErr
	}

	accounts, err := s.cfg.AccountSvc.GetAllAccounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts for dry run")
	}

	accountMap := make(map[int32]*database.Account, len(accounts))
	for _, account := range accounts {
		accountMap[account.ID] = account
	}

	if err = s.cfg.ValidationSvc.Validate(
		ctx,
		database.FromContext(ctx, database.GetDb(database.DbTypeReadonly)),
		[]*database.Transaction{updated},
		accountMap,
	); err != nil {
		return nil, errors.Wrap(err, "transaction validation failed after rule execution")
	}

	finalResp.RuleApplied = executed
	finalResp.After = s.cfg.MapperSvc.MapTransaction(ctx, updated)

	return finalResp, nil
}
