package rules

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
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
	dbRecord, err := s.transactionSvc.GetTransactionByIDs(ctx, []int64{req.TransactionId})
	if err != nil {
		return nil, err
	}

	if len(dbRecord) != 1 {
		return nil, errors.New("transaction not found")
	}

	tx := dbRecord[0]

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

	finalResp.RuleApplied = executed
	finalResp.After = s.mapperSvc.MapTransaction(ctx, updated)

	return finalResp, nil
}
