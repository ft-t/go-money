package transactions_test

import (
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestBasicExpenseWithMultiCurrency(t *testing.T) {
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

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		BaseAmountService: baseCurrency,
		RuleSvc:           ruleEngine,
	})

	accounts := []*database.Account{
		{
			Name:     "Private [UAH]",
			Currency: "UAH",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		Return(&gomoneypbv1.Transaction{})

	expenseDate := time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)
	_, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
			Withdrawal: &transactionsv1.Withdrawal{
				SourceAmount:    "-765.76",
				SourceCurrency:  "UAH",
				SourceAccountId: accounts[0].ID,
				ForeignAmount:   lo.ToPtr("-67.54"),
				ForeignCurrency: lo.ToPtr("PLN"),
			},
		},
	})
	assert.NoError(t, err)
}

func TestUpdateTransaction(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	statsSvc := transactions.NewStatService()
	mapper := NewMockMapperSvc(gomock.NewController(t))

	baseCurrency := NewMockBaseAmountSvc(gomock.NewController(t))
	baseCurrency.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, db *gorm.DB, i []*database.Transaction) error {
			assert.Len(t, i, 1)

			return nil
		}).AnyTimes()

	ruleEngine := NewMockRuleSvc(gomock.NewController(t))
	ruleEngine.EXPECT().ProcessTransactions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, i []*database.Transaction) ([]*database.Transaction, error) {
			assert.Len(t, i, 1)
			return i, nil
		}).AnyTimes()

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		BaseAmountService: baseCurrency,
		RuleSvc:           ruleEngine,
	})

	accounts := []*database.Account{
		{
			Name:     "Private [UAH]",
			Currency: "UAH",
			Extra:    map[string]string{},
		},
		{
			Name:     "Bank 2 [USD]",
			Currency: "USD",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, transaction *database.Transaction) *gomoneypbv1.Transaction {
			return &gomoneypbv1.Transaction{
				Id: transaction.ID,
			}
		}).AnyTimes()

	expenseDate := time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)
	tx1Result, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
			Withdrawal: &transactionsv1.Withdrawal{
				SourceAmount:    "-765.76",
				SourceCurrency:  "UAH",
				SourceAccountId: accounts[0].ID,
				ForeignAmount:   lo.ToPtr("-67.54"),
				ForeignCurrency: lo.ToPtr("PLN"),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, tx1Result)

	expenseDate2 := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	tx2Result, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate2),
		Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
			Withdrawal: &transactionsv1.Withdrawal{
				SourceAmount:    "-200.76",
				SourceCurrency:  "UAH",
				SourceAccountId: accounts[0].ID,
				ForeignAmount:   lo.ToPtr("-20.54"),
				ForeignCurrency: lo.ToPtr("PLN"),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, tx2Result)

	var stats []database.DailyStat
	assert.NoError(t, gormDB.
		Where("date >= ?", expenseDate).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "-765.76", stats[0].Amount.String()) // day of transaction
	assert.EqualValues(t, "-765.76", stats[1].Amount.String()) // next day no transaction
	assert.EqualValues(t, "-966.52", stats[2].Amount.String()) // new transaction

	var statBefore []database.DailyStat
	assert.NoError(t, gormDB.
		Where("date < ?", expenseDate).
		Order("date asc").Find(&statBefore).Error)

	assert.Len(t, statBefore, 0)

	// lets move last transaction to -1 day and change amount

	expenseDate3 := expenseDate2.Add(-24 * time.Hour)
	tx3Result, err := srv.Update(context.TODO(), &transactionsv1.UpdateTransactionRequest{
		Id: tx2Result.Transaction.Id,
		Transaction: &transactionsv1.CreateTransactionRequest{
			TransactionDate: timestamppb.New(expenseDate3),
			Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
				Withdrawal: &transactionsv1.Withdrawal{
					SourceAmount:    "-100.0",
					SourceCurrency:  "UAH",
					SourceAccountId: accounts[0].ID,
					ForeignAmount:   lo.ToPtr("-20.54"),
					ForeignCurrency: lo.ToPtr("PLN"),
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, tx3Result)

	assert.NoError(t, gormDB.
		Where("date >= ?", expenseDate).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "-765.76", stats[0].Amount.String()) // day of transaction
	assert.EqualValues(t, "-865.76", stats[1].Amount.String()) // next day no transaction
	assert.EqualValues(t, "-865.76", stats[2].Amount.String()) // new transaction

	// lets switch first transaction to another account

	_, err = srv.Update(context.TODO(), &transactionsv1.UpdateTransactionRequest{
		Id: tx1Result.Transaction.Id,
		Transaction: &transactionsv1.CreateTransactionRequest{
			TransactionDate: timestamppb.New(expenseDate),
			Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
				Withdrawal: &transactionsv1.Withdrawal{
					SourceAmount:    "-20.5",
					SourceCurrency:  "USD",
					SourceAccountId: accounts[1].ID,
					ForeignAmount:   lo.ToPtr("-4.54"),
					ForeignCurrency: lo.ToPtr("PLN"),
				},
			},
		},
	})
	assert.NoError(t, err)

	// first check first account stats
	assert.NoError(t, gormDB.
		Where("date >= ?", expenseDate).
		Where("account_id = ?", accounts[0].ID).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "0", stats[0].Amount.String())    // day of transaction
	assert.EqualValues(t, "-100", stats[1].Amount.String()) // from tx2 update
	assert.EqualValues(t, "-100", stats[2].Amount.String()) // new transaction

	assert.NoError(t, gormDB.
		Where("date >= ?", expenseDate).
		Where("account_id = ?", accounts[1].ID).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "-20.5", stats[0].Amount.String()) // day of transaction
	assert.EqualValues(t, "-20.5", stats[1].Amount.String()) // from tx2 update
	assert.EqualValues(t, "-20.5", stats[2].Amount.String()) // new transaction

	var updatedAccounts []database.Account
	assert.NoError(t, gormDB.Order("id asc").Find(&updatedAccounts).Error)

	assert.Len(t, updatedAccounts, 2)

	assert.EqualValues(t, "-100", updatedAccounts[0].CurrentBalance.String())
	assert.EqualValues(t, "-20.5", updatedAccounts[1].CurrentBalance.String())
}

func TestBasicCalc(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	statsSvc := transactions.NewStatService()
	mapper := NewMockMapperSvc(gomock.NewController(t))

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		Return(&gomoneypbv1.Transaction{}).AnyTimes()

	ruleEngine := NewMockRuleSvc(gomock.NewController(t))
	ruleEngine.EXPECT().ProcessTransactions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, i []*database.Transaction) ([]*database.Transaction, error) {
			assert.Len(t, i, 1)
			return i, nil
		}).Times(6)

	baseCurrency := NewMockBaseAmountSvc(gomock.NewController(t))
	baseCurrency.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, db *gorm.DB, i []*database.Transaction) error {
			assert.Len(t, i, 1)

			return nil
		}).Times(6)

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		BaseAmountService: baseCurrency,
		RuleSvc:           ruleEngine,
	})

	accounts := []*database.Account{
		{
			Name:     "BNP Paribas [USD]",
			Currency: "USD",
			Extra:    map[string]string{},
		},
		{
			Name:     "BNP Paribas [PLN]",
			Currency: "PLN",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	txDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	_, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(txDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "500",
				DestinationCurrency:  "USD",
				DestinationAccountId: accounts[0].ID,
			},
		},
	})
	assert.NoError(t, err)

	txTransferDate := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(txTransferDate),
		Transaction: &transactionsv1.CreateTransactionRequest_TransferBetweenAccounts{
			TransferBetweenAccounts: &transactionsv1.TransferBetweenAccounts{
				SourceAccountId:      accounts[0].ID,
				SourceAmount:         "-120",
				SourceCurrency:       "USD",
				DestinationAccountId: accounts[1].ID,
				DestinationAmount:    "480",
				DestinationCurrency:  "PLN",
			},
		},
	})
	assert.NoError(t, err)

	expenseDate := time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
			Withdrawal: &transactionsv1.Withdrawal{
				SourceAmount:    "-10",
				SourceCurrency:  "USD",
				SourceAccountId: accounts[0].ID,
			},
		},
	})
	assert.NoError(t, err)

	expenseDate2 := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate2),
		Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
			Withdrawal: &transactionsv1.Withdrawal{
				SourceAmount:    "-15",
				SourceCurrency:  "PLN",
				SourceAccountId: accounts[1].ID,
			},
		},
	})
	assert.NoError(t, err)

	depositDate2 := time.Date(2025, 6, 7, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(depositDate2),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "11",
				DestinationCurrency:  "PLN",
				DestinationAccountId: accounts[1].ID,
			},
		},
	})
	assert.NoError(t, err)

	depositDate3 := time.Date(2025, 6, 9, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(depositDate3),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "55",
				DestinationCurrency:  "PLN",
				DestinationAccountId: accounts[1].ID,
			},
		},
	})

	assert.NoError(t, err)
}

func TestBasicCalcWithGap(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
	statsSvc := transactions.NewStatService()
	mapper := NewMockMapperSvc(gomock.NewController(t))

	baseCurrency := NewMockBaseAmountSvc(gomock.NewController(t))
	baseCurrency.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, db *gorm.DB, i []*database.Transaction) error {
			assert.Len(t, i, 1)

			return nil
		}).AnyTimes()

	ruleEngine := NewMockRuleSvc(gomock.NewController(t))
	ruleEngine.EXPECT().ProcessTransactions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, i []*database.Transaction) ([]*database.Transaction, error) {
			assert.Len(t, i, 1)
			return i, nil
		}).AnyTimes()

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		Return(&gomoneypbv1.Transaction{}).AnyTimes()

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		BaseAmountService: baseCurrency,
		RuleSvc:           ruleEngine,
	})

	accounts := []*database.Account{
		{
			Name:     "BNP Paribas [USD]",
			Currency: "USD",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	txDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	_, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(txDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "500",
				DestinationCurrency:  "USD",
				DestinationAccountId: accounts[0].ID,
			},
		},
	})
	assert.NoError(t, err)

	var stats []database.DailyStat
	assert.NoError(t, gormDB.
		Where("date >= ?", txDate).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "500", stats[0].Amount.String()) // day of transaction
	assert.EqualValues(t, "500", stats[1].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[2].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[3].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[4].Amount.String()) // new day

	assert.NoError(t, gormDB.Exec("delete from daily_stat where date > ?", txDate).Error)

	assert.NoError(t, statsSvc.CalculateDailyStat(context.TODO(), gormDB, transactions.CalculateDailyStatRequest{
		StartDate: txDate.AddDate(0, 0, 1),
		AccountID: accounts[0].ID,
	}))

	assert.NoError(t, gormDB.
		Where("date >= ?", txDate).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "500", stats[0].Amount.String()) // day of transaction
	assert.EqualValues(t, "500", stats[1].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[2].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[3].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[4].Amount.String()) // new day
}

func TestNoDailyStatOnRecalculate(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
	statsSvc := transactions.NewStatService()
	mapper := NewMockMapperSvc(gomock.NewController(t))

	baseCurrency := NewMockBaseAmountSvc(gomock.NewController(t))
	baseCurrency.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, db *gorm.DB, i []*database.Transaction) error {
			assert.Len(t, i, 1)

			return nil
		}).AnyTimes()

	ruleEngine := NewMockRuleSvc(gomock.NewController(t))
	ruleEngine.EXPECT().ProcessTransactions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, i []*database.Transaction) ([]*database.Transaction, error) {
			assert.Len(t, i, 1)
			return i, nil
		}).AnyTimes()

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		Return(&gomoneypbv1.Transaction{}).AnyTimes()

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		BaseAmountService: baseCurrency,
		RuleSvc:           ruleEngine,
	})

	accounts := []*database.Account{
		{
			Name:     "BNP Paribas [USD]",
			Currency: "USD",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	txDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	_, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(txDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "500",
				DestinationCurrency:  "USD",
				DestinationAccountId: accounts[0].ID,
			},
		},
	})
	assert.NoError(t, err)

	var stats []database.DailyStat
	assert.NoError(t, gormDB.
		Where("date >= ?", txDate).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "500", stats[0].Amount.String()) // day of transaction
	assert.EqualValues(t, "500", stats[1].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[2].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[3].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[4].Amount.String()) // new day

	assert.NoError(t, gormDB.Exec("delete from daily_stat where date >= ?", txDate).Error)

	assert.NoError(t, statsSvc.CalculateDailyStat(context.TODO(), gormDB, transactions.CalculateDailyStatRequest{
		StartDate: txDate.AddDate(0, 0, 1),
		AccountID: accounts[0].ID,
	}))

	assert.NoError(t, gormDB.
		Where("date >= ?", txDate).
		Order("date asc").Find(&stats).Error)

	assert.EqualValues(t, "500", stats[0].Amount.String()) // day of transaction
	assert.EqualValues(t, "500", stats[1].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[2].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[3].Amount.String()) // new day
	assert.EqualValues(t, "500", stats[4].Amount.String()) // new day
}
