package rules

import (
	"context"
	"github.com/barkimedes/go-deepcopy"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
)

type Service struct {
	interpreter Interpreter
}

func NewService(interpreter Interpreter) *Service {
	return &Service{
		interpreter: interpreter,
	}
}

func (s *Service) cloneTx(input *database.Transaction) (*database.Transaction, error) {
	clonedAny, err := deepcopy.Anything(input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to deep clone transaction")
	}

	clonedTx, ok := clonedAny.(*database.Transaction)
	if !ok {
		return nil, errors.New("cloned transaction is not of type *database.Transaction")
	}

	return clonedTx, nil
}

func (s *Service) ProcessTransactions(
	ctx context.Context,
	inputTxs []*database.Transaction,
) ([]*database.Transaction, error) {
	if len(inputTxs) == 0 {
		return inputTxs, nil // no transactions to process
	}

	var processedTxs []*database.Transaction

	rules, err := s.getRules()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rules")
	}

	if len(rules) == 0 {
		return inputTxs, nil // no rules to apply
	}

	for _, inputTx := range inputTxs {
		tx, txErr := s.executeInternal(ctx, inputTx, rules)
		if txErr != nil {
			return nil, errors.Wrap(txErr, "failed to execute rules on transaction")
		}

		processedTxs = append(processedTxs, tx)
	}

	return processedTxs, nil
}

func (s *Service) executeInternal(
	ctx context.Context,
	inputTx *database.Transaction,
	ruleGroups []*RuleGroup,
) (*database.Transaction, error) {
	tx, cloneErr := s.cloneTx(inputTx) // clone initial transaction
	if cloneErr != nil {
		return nil, errors.Wrap(cloneErr, "failed to clone transaction")
	}

	for _, ruleGroup := range ruleGroups {
		for _, rule := range ruleGroup.Rules {
			clonedTxForRule, err := s.cloneTx(tx) // clone of cloned initial transaction for each rule

			result, err := s.interpreter.Run(ctx, rule.Script, clonedTxForRule)
			if err != nil { // errors should be handled in lua scripts
				return nil, err
			}

			if result {
				tx = clonedTxForRule
			}

			if result && rule.IsFinalRule {
				break // break execution for this group
			}
		}
	}

	return tx, nil
}

func (s *Service) getRules() ([]*RuleGroup, error) {
	// This function should return the rules from the database or any other source.
	// For now, we return an empty slice.
	return []*RuleGroup{
		{
			Rules: []*database.Rule{
				{
					Script: "somescript",
				},
				{
					Script: "somescript2",
				},
			},
		},
	}, nil
}
