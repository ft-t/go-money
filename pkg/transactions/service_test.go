package transactions_test

import (
	"context"
	"os"
	"testing"
	"time"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
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
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			TransactionDateTime:  time.Now(),
			Title:                "Test Deposit",
			DestinationAccountID: int32(123),
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(11)),
			DestinationCurrency:  "USD",
			Extra:                map[string]string{},
		},
		{
			TransactionType:     gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			TransactionDateTime: time.Now().Add(1 * time.Hour),
			Title:               "Test Withdrawal",
			SourceAccountID:     int32(456),
			SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(22)),
			SourceCurrency:      "EUR",
			Extra:               map[string]string{},
			TagIDs:              []int32{5},
			CategoryID:          lo.ToPtr(int32(5)),
		},
	}

	assert.NoError(t, gormDB.Create(&txs).Error)

	t.Run("list withdrawals", func(t *testing.T) {
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

	t.Run("list by tag", func(t *testing.T) {
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
			ToDate:    nil,
			TagIds:    []int32{5},
			TextQuery: nil,
			Skip:      0,
			Limit:     10,
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

	t.Run("list by category", func(t *testing.T) {
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
			FromDate:    timestamppb.New(time.Now().Add(-24 * time.Hour)),
			ToDate:      nil,
			CategoryIds: []int32{5},
			TextQuery:   nil,
			Skip:        0,
			Limit:       10,
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
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
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

func TestCreateReconciliation(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	statsSvc := transactions.NewStatService()
	mapper := NewMockMapperSvc(gomock.NewController(t))

	baseCurrency := NewMockBaseAmountSvc(gomock.NewController(t))
	baseCurrency.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, db *gorm.DB, i []*database.Transaction) error {
			assert.Len(t, i, 1)

			return nil
		})

	ruleEngine := NewMockRuleSvc(gomock.NewController(t))
	ruleEngine.EXPECT().ProcessTransactions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, i []*database.Transaction) ([]*database.Transaction, error) {
			assert.Len(t, i, 1)
			return i, nil
		})

	accountSvc := NewMockAccountSvc(gomock.NewController(t))
	validationSvc := NewMockValidationSvc(gomock.NewController(t))
	doubleEnty := NewMockDoubleEntrySvc(gomock.NewController(t))

	validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
	doubleEnty.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		RuleSvc:           ruleEngine,
		BaseAmountService: baseCurrency,
		AccountSvc:        accountSvc,
		ValidationSvc:     validationSvc,
		DoubleEntry:       doubleEnty,
	})

	accounts := []*database.Account{
		{
			Name:     "Private [UAH]",
			Currency: "UAH",
			Extra:    map[string]string{},
			Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		},
		{
			Name:     "Adjustment",
			Currency: "UAH",
			Extra:    map[string]string{},
			Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_ADJUSTMENT,
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)
	accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
			return &gomoneypbv1.Transaction{
				Id: transaction.ID,
			}
		})

	expenseDate := time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)
	resp, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Adjustment{
			Adjustment: &transactionsv1.Adjustment{
				DestinationAmount:    "556",
				DestinationCurrency:  accounts[0].Currency,
				DestinationAccountId: accounts[0].ID,
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	var createdTx *database.Transaction
	assert.NoError(t, gormDB.Where("id = ?", resp.Transaction.Id).Find(&createdTx).Error)

	assert.EqualValues(t, gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT, createdTx.TransactionType)
	assert.EqualValues(t, 556, createdTx.DestinationAmount.Decimal.IntPart())
	assert.EqualValues(t, accounts[0].Currency, createdTx.DestinationCurrency)
	assert.EqualValues(t, accounts[0].ID, createdTx.DestinationAccountID)
}

func TestCreateBulk(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	statsSvc := transactions.NewStatService()
	mapper := NewMockMapperSvc(gomock.NewController(t))

	baseCurrency := NewMockBaseAmountSvc(gomock.NewController(t))
	baseCurrency.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, db *gorm.DB, i []*database.Transaction) error {
			assert.Len(t, i, 2)

			return nil
		})

	ruleEngine := NewMockRuleSvc(gomock.NewController(t))
	ruleEngine.EXPECT().ProcessTransactions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, i []*database.Transaction) ([]*database.Transaction, error) {
			assert.Len(t, i, 2)
			return i, nil
		})

	accountSvc := NewMockAccountSvc(gomock.NewController(t))
	validationSvc := NewMockValidationSvc(gomock.NewController(t))
	doubleEnty := NewMockDoubleEntrySvc(gomock.NewController(t))

	validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
	doubleEnty.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		BaseAmountService: baseCurrency,
		RuleSvc:           ruleEngine,
		AccountSvc:        accountSvc,
		ValidationSvc:     validationSvc,
		DoubleEntry:       doubleEnty,
	})

	accounts := []*database.Account{
		{
			Name:     "Private [UAH]",
			Currency: "UAH",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
			return &gomoneypbv1.Transaction{
				Id: transaction.ID,
			}
		}).Times(2)

	expenseDate := time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)
	resp, err := srv.CreateBulk(context.TODO(), []*transactionsv1.CreateTransactionRequest{
		{
			TransactionDate: timestamppb.New(expenseDate),
			Transaction: &transactionsv1.CreateTransactionRequest_Adjustment{
				Adjustment: &transactionsv1.Adjustment{
					DestinationAmount:    "556",
					DestinationCurrency:  accounts[0].Currency,
					DestinationAccountId: accounts[0].ID,
				},
			},
		},
		{
			TransactionDate: timestamppb.New(expenseDate),
			Transaction: &transactionsv1.CreateTransactionRequest_Adjustment{
				Adjustment: &transactionsv1.Adjustment{
					DestinationAmount:    "777",
					DestinationCurrency:  accounts[0].Currency,
					DestinationAccountId: accounts[0].ID,
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Len(t, resp, 2)

	var ids []int64
	for _, r := range resp {
		assert.NotNil(t, r.Transaction)
		ids = append(ids, r.Transaction.Id)
	}

	var createdTx []*database.Transaction
	assert.NoError(t, gormDB.Where("id in ?", ids).Order("id asc").Find(&createdTx).Error)
	assert.Len(t, createdTx, 2)

	assert.EqualValues(t, gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT, createdTx[0].TransactionType)
	assert.EqualValues(t, 556, createdTx[0].DestinationAmount.Decimal.IntPart())
	assert.EqualValues(t, accounts[0].Currency, createdTx[0].DestinationCurrency)
	assert.EqualValues(t, accounts[0].ID, createdTx[0].DestinationAccountID)

	assert.EqualValues(t, gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT, createdTx[1].TransactionType)
	assert.EqualValues(t, 777, createdTx[1].DestinationAmount.Decimal.IntPart())
	assert.EqualValues(t, accounts[0].Currency, createdTx[1].DestinationCurrency)
	assert.EqualValues(t, accounts[0].ID, createdTx[1].DestinationAccountID)
}

func TestGetTransactionsByIDs(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	txs := []*database.Transaction{
		{
			TransactionType:     gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			TransactionDateTime: time.Now(),
			Title:               "Test Deposit",
			Extra:               map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&txs).Error)

	srv := transactions.NewService(&transactions.ServiceConfig{})

	resp, err := srv.GetTransactionByIDs(context.TODO(), []int64{txs[0].ID})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)
}

func TestCreateRawTransaction(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
	accounts := []*database.Account{
		{
			Name:     "Private [UAH]",
			Currency: "UAH",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	t.Run("success", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		baseSvc := NewMockBaseAmountSvc(gomock.NewController(t))
		mapper := NewMockMapperSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))
		doubleEnty := NewMockDoubleEntrySvc(gomock.NewController(t))

		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		doubleEnty.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

		svc := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:          statSvc,
			BaseAmountService: baseSvc,
			MapperSvc:         mapper,
			AccountSvc:        accountSvc,
			ValidationSvc:     validationSvc,
			DoubleEntry:       doubleEnty,
		})

		mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
				return &gomoneypbv1.Transaction{
					Id: transaction.ID,
				}
			})

		statSvc.EXPECT().HandleTransactions(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		baseSvc.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		newTx := &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAccountID: accounts[0].ID,
			DestinationCurrency:  accounts[0].Currency,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
		}
		resp, err := svc.CreateRawTransaction(context.TODO(), newTx)

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.NotEmpty(t, newTx.ID)
		assert.Equal(t, newTx.ID, resp.Transaction.Id)
	})

	t.Run("validation err", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		baseSvc := NewMockBaseAmountSvc(gomock.NewController(t))
		mapper := NewMockMapperSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))

		svc := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:          statSvc,
			BaseAmountService: baseSvc,
			MapperSvc:         mapper,
			AccountSvc:        accountSvc,
			ValidationSvc:     validationSvc,
		})

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)
		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("validation error"))

		newTx := &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAccountID: accounts[0].ID,
			DestinationCurrency:  accounts[0].Currency,
		}
		resp, err := svc.CreateRawTransaction(context.TODO(), newTx)

		assert.ErrorContains(t, err, "failed to validate transactions: validation error")
		assert.Nil(t, resp)
	})

	t.Run("finalize err", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		baseSvc := NewMockBaseAmountSvc(gomock.NewController(t))
		mapper := NewMockMapperSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))
		doubleEnty := NewMockDoubleEntrySvc(gomock.NewController(t))

		svc := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:          statSvc,
			BaseAmountService: baseSvc,
			MapperSvc:         mapper,
			AccountSvc:        accountSvc,
			ValidationSvc:     validationSvc,
			DoubleEntry:       doubleEnty,
		})

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)
		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		statSvc.EXPECT().HandleTransactions(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("unexpected error"))

		newTx := &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAccountID: accounts[0].ID,
			DestinationCurrency:  accounts[0].Currency,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
		}
		resp, err := svc.CreateRawTransaction(context.TODO(), newTx)

		assert.ErrorContains(t, err, "unexpected error")
		assert.Nil(t, resp)
	})

	t.Run("finalize err 2", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		baseSvc := NewMockBaseAmountSvc(gomock.NewController(t))
		mapper := NewMockMapperSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))
		doubleEnty := NewMockDoubleEntrySvc(gomock.NewController(t))

		svc := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:          statSvc,
			BaseAmountService: baseSvc,
			MapperSvc:         mapper,
			AccountSvc:        accountSvc,
			ValidationSvc:     validationSvc,
			DoubleEntry:       doubleEnty,
		})

		statSvc.EXPECT().HandleTransactions(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)
		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		baseSvc.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("unexpected error"))

		newTx := &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAccountID: accounts[0].ID,
			DestinationCurrency:  accounts[0].Currency,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
		}
		resp, err := svc.CreateRawTransaction(context.TODO(), newTx)

		assert.ErrorContains(t, err, "unexpected error")
		assert.Nil(t, resp)
	})

	t.Run("double entry err", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		baseSvc := NewMockBaseAmountSvc(gomock.NewController(t))
		mapper := NewMockMapperSvc(gomock.NewController(t))

		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		validationSvc := NewMockValidationSvc(gomock.NewController(t))
		doubleEnty := NewMockDoubleEntrySvc(gomock.NewController(t))

		svc := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:          statSvc,
			BaseAmountService: baseSvc,
			MapperSvc:         mapper,
			AccountSvc:        accountSvc,
			ValidationSvc:     validationSvc,
			DoubleEntry:       doubleEnty,
		})

		accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)
		validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		statSvc.EXPECT().HandleTransactions(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)
		baseSvc.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		doubleEnty.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(errors.New("unexpected error"))

		newTx := &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAccountID: accounts[0].ID,
			DestinationCurrency:  accounts[0].Currency,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
		}
		resp, err := svc.CreateRawTransaction(context.TODO(), newTx)

		assert.ErrorContains(t, err, "unexpected error")
		assert.Nil(t, resp)
	})

	t.Run("create err", func(t *testing.T) {
		statSvc := NewMockStatsSvc(gomock.NewController(t))
		baseSvc := NewMockBaseAmountSvc(gomock.NewController(t))
		mapper := NewMockMapperSvc(gomock.NewController(t))

		svc := transactions.NewService(&transactions.ServiceConfig{
			StatsSvc:          statSvc,
			BaseAmountService: baseSvc,
			MapperSvc:         mapper,
		})

		newTx := &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAccountID: accounts[0].ID,
			DestinationCurrency:  accounts[0].Currency,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
		}

		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		resp, err := svc.CreateRawTransaction(ctx, newTx)

		assert.ErrorContains(t, err, "context canceled")
		assert.Nil(t, resp)
	})
}
