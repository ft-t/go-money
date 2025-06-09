package transactions_test

import (
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm"
	"os"
	"testing"
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
