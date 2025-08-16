package rules_test

import (
	"context"
	"testing"

	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDryRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return([]*database.Account{
			{
				ID: 1,
			},
		}, nil)

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})

		tx := &database.Transaction{
			Title: "abcd",
		}

		executorSvc.EXPECT().ProcessSingleRule(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction, rule *database.Rule) (bool, *database.Transaction, error) {
				assert.EqualValues(t, "hello world", rule.Script)
				assert.Same(t, tx, transaction)

				return false, &database.Transaction{
					Title: "modified",
				}, nil
			})

		mapperSvc.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {

				return &gomoneypbv1.Transaction{
					Title: transaction.Title,
				}
			}).Times(2)

		transactionSvc.EXPECT().GetTransactionByIDs(gomock.Any(), []int64{1}).
			Return([]*database.Transaction{tx}, nil)

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "hello world",
			},
			TransactionId: 1,
		})
		assert.NoError(t, err)

		assert.Equal(t, "abcd", resp.Before.Title)
		assert.Equal(t, "modified", resp.After.Title)
	})

	t.Run("get account err", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(nil, assert.AnError)

		executorSvc.EXPECT().ProcessSingleRule(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction, rule *database.Rule) (bool, *database.Transaction, error) {
				assert.EqualValues(t, "hello world", rule.Script)

				return false, &database.Transaction{
					Title: "modified",
				}, nil
			})

		tx := &database.Transaction{
			Title: "abcd",
		}
		transactionSvc.EXPECT().GetTransactionByIDs(gomock.Any(), []int64{1}).
			Return([]*database.Transaction{tx}, nil)
		mapperSvc.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {

				return &gomoneypbv1.Transaction{
					Title: transaction.Title,
				}
			}).Times(1)

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "hello world",
			},
			TransactionId: 1,
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("success with tx id = 0", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return([]*database.Account{
			{
				ID: 1,
			},
		}, nil)

		executorSvc.EXPECT().ProcessSingleRule(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction, rule *database.Rule) (bool, *database.Transaction, error) {
				return false, &database.Transaction{
					Title: "modified",
				}, nil
			})

		mapperSvc.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {

				return &gomoneypbv1.Transaction{
					Title: transaction.Title,
				}
			}).Times(2)

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "tx:title(modified)",
			},
			TransactionId: 0,
		})
		assert.NoError(t, err)

		assert.Equal(t, "", resp.Before.Title)
		assert.Equal(t, "modified", resp.After.Title)
	})

	t.Run("transaction not found", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})
		transactionSvc.EXPECT().GetTransactionByIDs(gomock.Any(), []int64{1}).
			Return([]*database.Transaction{
				{},
				{},
			}, nil)

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "hello world",
			},
			TransactionId: 1,
		})

		assert.ErrorContains(t, err, "transaction not found")
		assert.Nil(t, resp)
	})

	t.Run("get transaction error", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})
		transactionSvc.EXPECT().GetTransactionByIDs(gomock.Any(), gomock.Any()).
			Return(nil, assert.AnError)

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "hello world",
			},
			TransactionId: 1,
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("validate transaction error", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return([]*database.Account{
			{
				ID: 1,
			},
		}, nil)

		tx := &database.Transaction{
			Title: "abcd",
		}

		transactionSvc.EXPECT().GetTransactionByIDs(gomock.Any(), []int64{1}).
			Return([]*database.Transaction{tx}, nil)

		mapperSvc.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			Return(&gomoneypbv1.Transaction{})

		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(assert.AnError)

		executorSvc.EXPECT().ProcessSingleRule(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction, rule *database.Rule) (bool, *database.Transaction, error) {
				assert.EqualValues(t, "hello world", rule.Script)
				assert.Same(t, tx, transaction)

				return false, &database.Transaction{
					Title: "modified",
				}, nil
			})

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "hello world",
			},
			TransactionId: 1,
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("transactionSvc error", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})
		transactionSvc.EXPECT().GetTransactionByIDs(gomock.Any(), gomock.Any()).
			Return(nil, assert.AnError)

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "hello world",
			},
			TransactionId: 1,
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("executorSvc error", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))

		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))

		dry := rules.NewDryRun(&rules.DryRunConfig{
			Executor:       executorSvc,
			TransactionSvc: transactionSvc,
			MapperSvc:      mapperSvc,
			ValidationSvc:  validationSvc,
			AccountSvc:     accountSvc,
		})

		mapperSvc.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {

				return &gomoneypbv1.Transaction{
					Title: transaction.Title,
				}
			})

		tx := &database.Transaction{
			Title: "abcd",
		}

		executorSvc.EXPECT().ProcessSingleRule(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(false, nil, assert.AnError)

		transactionSvc.EXPECT().GetTransactionByIDs(gomock.Any(), []int64{1}).
			Return([]*database.Transaction{tx}, nil)

		resp, err := dry.DryRunRule(context.TODO(), &rulesv1.DryRunRuleRequest{
			Rule: &gomoneypbv1.Rule{
				Script: "hello world",
			},
			TransactionId: 1,
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
