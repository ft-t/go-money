package rules

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"time"
)

type DryRun struct {
	executor       ExecutorSvc
	transactionSvc TransactionSvc
	mapperSvc      MapperSvc
}

func NewDryRun(
	executor ExecutorSvc,
	transactionSvc TransactionSvc,
	mapperSvc MapperSvc,
) *DryRun {
	return &DryRun{
		executor:       executor,
		transactionSvc: transactionSvc,
		mapperSvc:      mapperSvc,
	}
}

func (s *DryRun) DryRunRule(ctx context.Context, req *rulesv1.DryRunRuleRequest) (*rulesv1.DryRunRuleResponse, error) {
	var tx *database.Transaction
	if req.TransactionId != 0 {
		dbRecord, err := s.transactionSvc.GetTransactionByIDs(ctx, []int64{req.TransactionId})
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
		Before:      s.mapperSvc.MapTransaction(ctx, tx),
		After:       nil,
		RuleApplied: false,
	}

	executed, updated, ruleErr := s.executor.ProcessSingleRule(ctx, tx, &database.Rule{
		Script: req.Rule.Script,
		Title:  req.Rule.Title,
	})
	if ruleErr != nil {
		return nil, ruleErr
	}

	if err := s.transactionSvc.ValidateTransactionData(
		ctx,
		database.FromContext(ctx, database.GetDb(database.DbTypeReadonly)),
		updated,
	); err != nil {
		return nil, errors.Wrap(err, "transaction validation failed after rule execution")
	}

	finalResp.RuleApplied = executed
	finalResp.After = s.mapperSvc.MapTransaction(ctx, updated)

	return finalResp, nil
}
