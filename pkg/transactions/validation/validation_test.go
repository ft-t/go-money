package validation_test

import (
	"context"
	"os"
	"testing"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/applicable_accounts"
	"github.com/ft-t/go-money/pkg/transactions/validation"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
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

func TestValidateWithdrawal(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
	tx := gormDB.Begin()
	defer tx.Rollback()

	acc := []*database.Account{
		{
			Name:     "Test USD",
			Currency: "USD",
			Extra:    map[string]string{},
		},
		{
			Name:     "Test EUR",
			Currency: "EUR",
			Extra:    map[string]string{},
		},
		{
			Name:     "Test PLN",
			Currency: "PLN",
			Extra:    map[string]string{},
		},
		{
			Name:     "Default expense account",
			Extra:    map[string]string{},
			Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
			Currency: "USD",
		},
	}
	assert.NoError(t, gormDB.Create(&acc).Error)

	accountMap := buildAccountMap(acc)

	t.Run("valid withdrawal", func(t *testing.T) {
		applicableSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: applicableSvc,
		})

		applicableSvc.EXPECT().GetApplicableAccounts(gomock.Any(), gomock.Any()).
			Return(map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE: {
					SourceAccounts: map[int32]*database.Account{
						acc[0].ID: acc[0],
					},
					DestinationAccounts: map[int32]*database.Account{
						acc[3].ID: acc[3],
					},
				},
			})

		assert.NoError(t, srv.Validate(context.TODO(), gormDB, &validation.Request{
			Accounts: accountMap,
			Txs: []*database.Transaction{
				{
					TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
					SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
					SourceCurrency:  acc[0].Currency,
					SourceAccountID: acc[0].ID,

					DestinationAccountID: acc[3].ID,
					DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
					DestinationCurrency:  acc[3].Currency,
				},
			},
		}))
	})

	t.Run("invalid - positive amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: acc[0].ID,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_amount must be negative for TRANSACTION_TYPE_EXPENSE")
	})

	t.Run("invalid - no source account", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: 0,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_account_id is required for TRANSACTION_TYPE_EXPENSE")
	})

	t.Run("invalid - no source amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NullDecimal{},
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: acc[0].ID,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_amount is required for TRANSACTION_TYPE_EXPENSE")
	})

	t.Run("invalid - no source currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  "",
			SourceAccountID: acc[0].ID,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_currency is required for TRANSACTION_TYPE_EXPENSE")
	})

	t.Run("invalid - no source account ID", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: 0,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_account_id is required for TRANSACTION_TYPE_EXPENSE")
	})

	t.Run("invalid - source currency mismatch", func(t *testing.T) {
		appAcc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: appAcc,
		})

		appAcc.EXPECT().GetApplicableAccounts(gomock.Any(), gomock.Any()).
			Return(map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE: {
					SourceAccounts: map[int32]*database.Account{
						acc[0].ID: acc[1],
					},
					DestinationAccounts: map[int32]*database.Account{
						acc[3].ID: acc[3],
					},
				},
			})

		err := srv.Validate(context.TODO(), gormDB, &validation.Request{
			Accounts: accountMap,
			Txs: []*database.Transaction{
				{
					TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
					SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
					SourceCurrency:  acc[1].Currency,
					SourceAccountID: acc[0].ID,

					DestinationAccountID: acc[3].ID,
					DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
					DestinationCurrency:  acc[3].Currency,
				},
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency USD, expected EUR")
	})

	t.Run("invalid - fx amount without fx currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:  gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:     decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:   acc[0].Currency,
			SourceAccountID:  acc[0].ID,
			FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			FxSourceCurrency: "",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "fx_source_currency is required")
	})

	t.Run("invalid - fx amount with fx currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:  gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:     decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:   acc[0].Currency,
			SourceAccountID:  acc[0].ID,
			FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(100)),
			FxSourceCurrency: acc[1].Currency,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "fx_source_amount must be negative for TRANSACTION_TYPE_EXPENSE")
	})

	t.Run("invalid - destination amount without destination currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:     gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:      acc[0].Currency,
			SourceAccountID:     acc[0].ID,
			DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency: "",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required")
	})

	t.Run("invalid - destination amount with destination currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)

		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:     gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:      acc[0].Currency,
			SourceAccountID:     acc[0].ID,
			DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationCurrency: acc[1].Currency,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount must be positive for TRANSACTION_TYPE_EXPENSE")
	})
}

func TestValidateDeposit(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	acc := []*database.Account{
		{Name: "Test USD", Currency: "USD", Extra: map[string]string{}},
		{Name: "Test EUR", Currency: "EUR", Extra: map[string]string{}},
	}
	assert.NoError(t, gormDB.Create(&acc).Error)

	accountMap := make(map[int32]*database.Account, len(acc))
	for _, a := range acc {
		accountMap[a.ID] = a
	}

	t.Run("valid deposit", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		assert.NoError(t, srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: acc[0].ID,
		}))
	})

	t.Run("invalid - negative amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount must be positive for TRANSACTION_TYPE_INCOME")
	})

	t.Run("invalid - no destination account", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: 0,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_account_id is required for TRANSACTION_TYPE_INCOME")
	})

	t.Run("invalid - no destination amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NullDecimal{},
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount is required for TRANSACTION_TYPE_INCOME")
	})

	t.Run("invalid - no destination currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency:  "",
			DestinationAccountID: acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required for TRANSACTION_TYPE_INCOME")
	})

	t.Run("invalid - destination currency mismatch", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))
		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		accSvc.EXPECT().GetApplicableAccounts(gomock.Any(), gomock.Any()).
			Return(map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME: {
					SourceAccounts: map[int32]*database.Account{
						acc[0].ID: acc[0],
					},
					DestinationAccounts: map[int32]*database.Account{
						acc[1].ID: acc[1],
					},
				},
			})

		err := srv.Validate(context.TODO(), gormDB, &validation.Request{
			Accounts: accountMap,
			Txs: []*database.Transaction{
				{
					TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
					DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
					DestinationCurrency:  "RAND",
					DestinationAccountID: acc[1].ID,

					SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-100)),
					SourceCurrency:  acc[0].Currency,
					SourceAccountID: acc[0].ID,
				},
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency EUR, expected RAND")
	})
}

func TestValidateReconciliation(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	acc := []*database.Account{
		{
			Name:     "Test USD",
			Currency: "USD",
			Extra:    map[string]string{},
		},
		{
			Name:     "reconciliation account",
			Currency: "USD",
			Extra:    map[string]string{},
			Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE, // todo
		},
	}
	assert.NoError(t, gormDB.Create(&acc).Error)

	accMap := buildAccountMap(acc)

	t.Run("valid reconciliation", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		assert.NoError(t, srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: acc[0].ID,
		}))
	})

	t.Run("success - negative amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-200)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: acc[0].ID,
		})

		assert.NoError(t, err)
	})

	t.Run("invalid - no destination account", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: 0,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_account_id is required for TRANSACTION_TYPE_ADJUSTMENT")
	})

	t.Run("invalid - no destination amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NullDecimal{},
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount is required for TRANSACTION_TYPE_ADJUSTMENT")
	})

	t.Run("invalid - no destination currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
			DestinationCurrency:  "",
			DestinationAccountID: acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required for TRANSACTION_TYPE_ADJUSTMENT")
	})

	t.Run("invalid - destination currency mismatch", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		accSvc.EXPECT().GetApplicableAccounts(gomock.Any(), gomock.Any()).
			Return(map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT: {
					SourceAccounts: map[int32]*database.Account{
						acc[0].ID: acc[0],
					},
					DestinationAccounts: map[int32]*database.Account{
						acc[1].ID: acc[1],
					},
				},
			})

		err := srv.Validate(context.TODO(), gormDB, &validation.Request{
			Accounts: accMap,
			Txs: []*database.Transaction{
				{
					TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
					DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
					DestinationCurrency:  "EUR",
					DestinationAccountID: acc[1].ID,

					SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-200)),
					SourceCurrency:  acc[0].Currency,
					SourceAccountID: acc[0].ID,
				},
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency USD, expected EUR")
	})
}

func TestValidateTransferBetweenAccounts(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	acc := []*database.Account{
		{Name: "Test USD", Currency: "USD", Extra: map[string]string{}},
		{Name: "Test EUR", Currency: "EUR", Extra: map[string]string{}},
	}
	assert.NoError(t, gormDB.Create(&acc).Error)

	accountMap := buildAccountMap(acc)

	t.Run("valid transfer", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		assert.NoError(t, srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: acc[1].ID,
		}))
	})

	t.Run("invalid - no source account", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      0,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_account_id is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no destination account", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: 0,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_account_id is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no source amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NullDecimal{},
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_amount is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no destination amount", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      acc[0].ID,
			DestinationAmount:    decimal.NullDecimal{},
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no source currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       "",
			SourceAccountID:      acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_currency is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no destination currency", func(t *testing.T) {
		srv := validation.NewValidationService(nil)
		err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  "",
			DestinationAccountID: acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - source currency mismatch", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		accSvc.EXPECT().GetApplicableAccounts(gomock.Any(), gomock.Any()).
			Return(map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS: {
					SourceAccounts: map[int32]*database.Account{
						acc[0].ID: acc[0],
					},
					DestinationAccounts: map[int32]*database.Account{
						acc[1].ID: acc[1],
					},
				},
			})

		err := srv.Validate(context.TODO(), gormDB, &validation.Request{
			Accounts: accountMap,
			Txs: []*database.Transaction{
				{
					TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
					SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
					SourceCurrency:       acc[1].Currency,
					SourceAccountID:      acc[0].ID,
					DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
					DestinationCurrency:  acc[1].Currency,
					DestinationAccountID: acc[1].ID,
				},
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency USD, expected EUR")
	})

	t.Run("invalid - destination currency mismatch", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		accSvc.EXPECT().GetApplicableAccounts(gomock.Any(), gomock.Any()).
			Return(map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS: {
					SourceAccounts: map[int32]*database.Account{
						acc[0].ID: acc[0],
					},
					DestinationAccounts: map[int32]*database.Account{
						acc[1].ID: acc[1],
					},
				},
			})

		err := srv.Validate(context.TODO(), gormDB, &validation.Request{
			Accounts: accountMap,
			Txs: []*database.Transaction{
				{
					TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
					SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
					SourceCurrency:       acc[0].Currency,
					SourceAccountID:      acc[0].ID,
					DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
					DestinationCurrency:  acc[0].Currency,
					DestinationAccountID: acc[1].ID,
				},
			},
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency EUR, expected USD")
	})
}

func TestValidateInvalidType(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := validation.NewValidationService(nil)

	err := srv.ValidateTransactionData(context.TODO(), gormDB, &database.Transaction{
		TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED,
	})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "unsupported transaction type: TRANSACTION_TYPE_UNSPECIFIED")
}

func TestValidateTxAccounts(t *testing.T) {
	t.Run("no possible accounts for tx type", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		err := srv.ValidateTransactionAccounts(context.TODO(), map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
			gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME: {},
		}, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "no possible accounts found for transaction type")
	})

	t.Run("source account not found", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		err := srv.ValidateTransactionAccounts(context.TODO(), map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
			gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME: {
				SourceAccounts: map[int32]*database.Account{},
			},
		}, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			SourceAccountID: 1,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "is not applicable for transaction type")
	})

	t.Run("destination account not found", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		err := srv.ValidateTransactionAccounts(context.TODO(), map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
			gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME: {
				SourceAccounts: map[int32]*database.Account{
					1: {},
				},
				DestinationAccounts: map[int32]*database.Account{},
			},
		}, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAccountID: 1,
			SourceAccountID:      1,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "is not applicable for transaction type")
	})

	t.Run("success", func(t *testing.T) {
		accSvc := NewMockApplicableAccountSvc(gomock.NewController(t))

		srv := validation.NewValidationService(&validation.ServiceConfig{
			ApplicableAccountSvc: accSvc,
		})

		err := srv.ValidateTransactionAccounts(context.TODO(), map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
			gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME: {
				SourceAccounts: map[int32]*database.Account{
					1: {ID: 1, Currency: "USD"},
				},
				DestinationAccounts: map[int32]*database.Account{
					2: {ID: 2, Currency: "USD"},
				},
			},
		}, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAccountID: 2,
			SourceAccountID:      1,
			DestinationCurrency:  "USD",
			SourceCurrency:       "USD",
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-100)),
		})

		assert.NoError(t, err)
	})
}

func buildAccountMap(acc []*database.Account) map[int32]*database.Account {
	accountMap := make(map[int32]*database.Account, len(acc))

	for _, a := range acc {
		accountMap[a.ID] = a
	}

	return accountMap
}
