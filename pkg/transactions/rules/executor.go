package rules

import (
	"context"
	"github.com/barkimedes/go-deepcopy"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"sort"
)

type Executor struct {
	interpreter Interpreter
}

func NewExecutor(interpreter Interpreter) *Executor {
	return &Executor{
		interpreter: interpreter,
	}
}

func (s *Executor) cloneTx(input *database.Transaction) (*database.Transaction, error) {
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

func (s *Executor) ProcessTransactions(
	ctx context.Context,
	inputTxs []*database.Transaction,
) ([]*database.Transaction, error) {
	if len(inputTxs) == 0 {
		return inputTxs, nil // no transactions to process
	}

	var processedTxs []*database.Transaction

	rules, err := s.getRules(ctx)
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

func (s *Executor) executeInternal(
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
			if err != nil {
				return nil, errors.Wrap(err, "failed to clone transaction for rule execution")
			}

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

func (s *Executor) getRules(
	ctx context.Context,
) ([]*RuleGroup, error) {
	var rules []*database.Rule

	if err := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly)).
		Order("sort_order asc").
		Find(&rules).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get rules from database")
	}

	ruleGroups := map[string]*RuleGroup{}

	for _, rule := range rules {
		if _, exists := ruleGroups[rule.GroupName]; !exists {
			ruleGroups[rule.GroupName] = &RuleGroup{
				Name: rule.GroupName,
			}
		}

		ruleGroups[rule.GroupName].Rules = append(ruleGroups[rule.GroupName].Rules, rule)
	}

	groupNames := lo.Keys(ruleGroups)
	sort.Strings(groupNames) // ordering

	var orderedRuleGroups []*RuleGroup
	for _, key := range groupNames {
		ruleGroup := ruleGroups[key]

		sort.Slice(ruleGroup.Rules, func(i, j int) bool {
			return ruleGroup.Rules[i].SortOrder < ruleGroup.Rules[j].SortOrder
		})

		orderedRuleGroups = append(orderedRuleGroups, ruleGroup)
	}

	return orderedRuleGroups, nil
}
