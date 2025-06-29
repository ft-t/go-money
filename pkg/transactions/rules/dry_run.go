package rules

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	"context"
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
	dbTransactions, err := s.transactionSvc.GetTransactionByIDs(ctx, req.TransactionIds)
	if err != nil {
		return nil, err
	}

	finalResp := &rulesv1.DryRunRuleResponse{}

	for _, tx := range dbTransactions {
		_, updated, ruleErr := s.executor.ProcessSingleRule(ctx, tx, &database.Rule{
			Script: req.Rule.Script,
			Title:  req.Rule.Title,
		})
		if ruleErr != nil {
			return nil, ruleErr
		}

		finalResp.UpdatedTransactions = append(finalResp.UpdatedTransactions, s.mapperSvc.MapTransaction(ctx, updated))
	}

	return finalResp, nil
}
