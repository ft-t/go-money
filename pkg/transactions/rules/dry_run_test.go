package rules_test

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDryRun(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))

		dry := rules.NewDryRun(executorSvc, transactionSvc, mapperSvc)

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

	t.Run("transactionSvc error", func(t *testing.T) {
		executorSvc := NewMockExecutorSvc(gomock.NewController(t))
		transactionSvc := NewMockTransactionSvc(gomock.NewController(t))
		mapperSvc := NewMockMapperSvc(gomock.NewController(t))

		dry := rules.NewDryRun(executorSvc, transactionSvc, mapperSvc)

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

		dry := rules.NewDryRun(executorSvc, transactionSvc, mapperSvc)

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
