package transactions_test

import (
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(
	m *testing.M,
) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestListTransactions(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	txs := []*database.Transaction{
		{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT,
			TransactionDateTime:  time.Now(),
			Title:                "Test Deposit",
			DestinationAccountID: lo.ToPtr(int32(123)),
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(11)),
			DestinationCurrency:  "USD",
			Extra:                map[string]string{},
		},
		{
			TransactionType:     gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
			TransactionDateTime: time.Now().Add(1 * time.Hour),
			Title:               "Test Withdrawal",
			SourceAccountID:     lo.ToPtr(int32(456)),
			SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(22)),
			SourceCurrency:      "EUR",
			Extra:               map[string]string{},
		},
	}

	assert.NoError(t, gormDB.Create(&txs).Error)

	t.Run("list list withdrawals", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:  nil,
			MapperSvc: mapper,
		})

		mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
				return &gomoneypbv1.Transaction{
					Id: transaction.ID,
				}
			})

		resp, err := srv.List(context.TODO(), &transactionsv1.ListTransactionsRequest{
			SourceAccountIds: []int32{456},
			FromDate:         timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:           nil,
			TextQuery:        nil,
			Skip:             0,
			Limit:            10,
			Sort: []*transactionsv1.ListTransactionsRequest_Sort{
				{
					Field:     transactionsv1.SortField_SORT_FIELD_TRANSACTION_DATE,
					Ascending: false,
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Transactions, 1)
		assert.EqualValues(t, 1, resp.TotalCount)
		assert.EqualValues(t, txs[1].ID, resp.Transactions[0].Id)
	})

	t.Run("any account id", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:  nil,
			MapperSvc: mapper,
		})

		mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
				return &gomoneypbv1.Transaction{
					Id: transaction.ID,
				}
			})

		resp, err := srv.List(context.TODO(), &transactionsv1.ListTransactionsRequest{
			FromDate:  timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:    timestamppb.New(time.Now().Add(24 * time.Hour)),
			TextQuery: lo.ToPtr("Withdrawal"),
			Skip:      0,
			Limit:     10,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Transactions, 1)
		assert.EqualValues(t, 1, resp.TotalCount)
		assert.EqualValues(t, txs[1].ID, resp.Transactions[0].Id)
	})

	t.Run("text query", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:  nil,
			MapperSvc: mapper,
		})

		mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
				return &gomoneypbv1.Transaction{
					Id: transaction.ID,
				}
			})

		resp, err := srv.List(context.TODO(), &transactionsv1.ListTransactionsRequest{
			AnyAccountIds: []int32{456},
			FromDate:      timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:        nil,
			TextQuery:     nil,
			Skip:          0,
			Limit:         10,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Transactions, 1)
		assert.EqualValues(t, 1, resp.TotalCount)
		assert.EqualValues(t, txs[1].ID, resp.Transactions[0].Id)
	})

	t.Run("text query", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:  nil,
			MapperSvc: mapper,
		})

		mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
				return &gomoneypbv1.Transaction{
					Id: transaction.ID,
				}
			})

		resp, err := srv.List(context.TODO(), &transactionsv1.ListTransactionsRequest{
			DestinationAccountIds: []int32{123},
			FromDate:              timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:                nil,
			TextQuery:             nil,
			Skip:                  0,
			Limit:                 10,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Transactions, 1)
		assert.EqualValues(t, 1, resp.TotalCount)
		assert.EqualValues(t, txs[0].ID, resp.Transactions[0].Id)
	})

	t.Run("transaction type", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:  nil,
			MapperSvc: mapper,
		})

		mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
				return &gomoneypbv1.Transaction{
					Id: transaction.ID,
				}
			})

		resp, err := srv.List(context.TODO(), &transactionsv1.ListTransactionsRequest{
			TransactionTypes: []gomoneypbv1.TransactionType{
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT,
			},
			FromDate:  timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:    nil,
			TextQuery: nil,
			Skip:      0,
			Limit:     10,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Transactions, 1)
		assert.EqualValues(t, 1, resp.TotalCount)
		assert.EqualValues(t, txs[0].ID, resp.Transactions[0].Id)
	})

	t.Run("amounts success", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:  nil,
			MapperSvc: mapper,
		})

		mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
				return &gomoneypbv1.Transaction{
					Id: transaction.ID,
				}
			})

		resp, err := srv.List(context.TODO(), &transactionsv1.ListTransactionsRequest{
			AmountFrom: lo.ToPtr("22"),
			AmountTo:   lo.ToPtr("23"),
			FromDate:   timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:     nil,
			TextQuery:  nil,
			Skip:       0,
			Limit:      10,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Transactions, 1)
		assert.EqualValues(t, 1, resp.TotalCount)
		assert.EqualValues(t, txs[1].ID, resp.Transactions[0].Id)
	})

	t.Run("amounts no record", func(t *testing.T) {
		mapper := NewMockMapperSvc(gomock.NewController(t))
		srv := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:  nil,
			MapperSvc: mapper,
		})

		resp, err := srv.List(context.TODO(), &transactionsv1.ListTransactionsRequest{
			AmountFrom: lo.ToPtr("24"),
			AmountTo:   lo.ToPtr("25"),
			FromDate:   timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:     nil,
			TextQuery:  nil,
			Skip:       0,
			Limit:      10,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.Len(t, resp.Transactions, 0)
		assert.EqualValues(t, 0, resp.TotalCount)
	})
}

//
//func TestCreateWithdrawal(
//	t *testing.T,
//) {
//	t.Run("success", func(t *testing.T) {
//		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
//
//		statSvc := NewMockStatsSvc(gomock.NewController(t))
//		srv := transactions.NewService(
//			&transactions.ServiceConfig{
//				StatsSvc: statSvc,
//			},
//		)
//
//		account := &database.Account{
//			Currency: "USD",
//			Extra:    map[string]string{},
//		}
//		assert.NoError(t, gormDB.Create(account).Error)
//
//		timeNow := time.Now().UTC()
//
//		statSvc.EXPECT().
//			ProcessTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
//			DoAndReturn(func(ctx context.Context, db *gorm.DB, transaction *database.Transaction) error {
//				assert.EqualValues(
//					t,
//					account.ID,
//					*transaction.SourceAccountID,
//				)
//				assert.EqualValues(
//					t,
//					gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
//					transaction.TransactionType,
//				)
//				return nil
//			})
//
//		resp, err := srv.Create(
//			context.TODO(),
//			&transactionsv1.CreateTransactionRequest{
//				Notes:    "",
//				Extra:    nil,
//				LabelIds: nil,
//				TransactionDate: timestamppb.New(
//					timeNow,
//				),
//				Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
//					Withdrawal: &transactionsv1.Withdrawal{
//						SourceAccountId: account.ID,
//						SourceAmount:    "-55.21",
//						SourceCurrency:  "USD",
//					},
//				},
//			},
//		)
//		assert.NoError(t, err)
//		assert.NotNil(t, resp)
//	})
//
//	t.Run("invalid amount format", func(t *testing.T) {
//		srv := transactions.NewService(
//			&transactions.ServiceConfig{},
//		)
//
//		resp, err := srv.Create(
//			context.TODO(),
//			&transactionsv1.CreateTransactionRequest{
//				Notes:    "",
//				Extra:    nil,
//				LabelIds: nil,
//				TransactionDate: timestamppb.New(
//					time.Now().
//						UTC(),
//				),
//				Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
//					Withdrawal: &transactionsv1.Withdrawal{
//						SourceAccountId: 1,
//						SourceAmount:    "invalid",
//						SourceCurrency:  "USD",
//					},
//				},
//			},
//		)
//		assert.Error(t, err)
//		assert.Nil(t, resp)
//	})
//
//	t.Run("invalid source account id", func(t *testing.T) {
//		srv := transactions.NewService(
//			&transactions.ServiceConfig{},
//		)
//
//		resp, err := srv.Create(
//			context.TODO(),
//			&transactionsv1.CreateTransactionRequest{
//				Notes:    "",
//				Extra:    nil,
//				LabelIds: nil,
//				TransactionDate: timestamppb.New(
//					time.Now().
//						UTC(),
//				),
//				Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
//					Withdrawal: &transactionsv1.Withdrawal{
//						SourceAccountId: -100,
//						SourceAmount:    "55.21",
//						SourceCurrency:  "USD",
//					},
//				},
//			},
//		)
//		assert.ErrorContains(
//			t,
//			err,
//			"source account id is required",
//		)
//		assert.Nil(t, resp)
//	})
//
//	t.Run("source amount should not be positive", func(t *testing.T) {
//		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
//		srv := transactions.NewService(
//			&transactions.ServiceConfig{},
//		)
//
//		account := &database.Account{
//			Currency: "USD",
//			Extra:    map[string]string{},
//		}
//		assert.NoError(t, gormDB.Create(account).Error)
//
//		resp, err := srv.Create(
//			context.TODO(),
//			&transactionsv1.CreateTransactionRequest{
//				Notes:    "",
//				Extra:    nil,
//				LabelIds: nil,
//				TransactionDate: timestamppb.New(
//					time.Now().
//						UTC(),
//				),
//				Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
//					Withdrawal: &transactionsv1.Withdrawal{
//						SourceAccountId: 1,
//						SourceAmount:    "55.21",
//						SourceCurrency:  "USD",
//					},
//				},
//			},
//		)
//		assert.ErrorContains(
//			t,
//			err,
//			"source amount must be negative",
//		)
//		assert.Nil(t, resp)
//	})
//
//	t.Run("invalid account currency", func(t *testing.T) {
//		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
//		srv := transactions.NewService(
//			&transactions.ServiceConfig{},
//		)
//
//		account := &database.Account{
//			Currency: "USD",
//			Extra:    map[string]string{},
//		}
//		assert.NoError(t, gormDB.Create(account).Error)
//
//		resp, err := srv.Create(
//			context.TODO(),
//			&transactionsv1.CreateTransactionRequest{
//				Notes:    "",
//				Extra:    nil,
//				LabelIds: nil,
//				TransactionDate: timestamppb.New(
//					time.Now().
//						UTC(),
//				),
//				Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
//					Withdrawal: &transactionsv1.Withdrawal{
//						SourceAccountId: account.ID,
//						SourceAmount:    "-55.21",
//						SourceCurrency:  "EUR",
//					},
//				},
//			},
//		)
//
//		assert.ErrorContains(
//			t,
//			err,
//			"has currency USD, expected EUR",
//		)
//		assert.Nil(t, resp)
//	})
//}
