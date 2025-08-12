package transactions_test

import (
	"context"
	"testing"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

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
	}
	assert.NoError(t, gormDB.Create(&acc).Error)

	t.Run("valid withdrawal", func(t *testing.T) {
		srv := transactions.NewService(nil)

		assert.NoError(t, srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: &acc[0].ID,
		}))
	})

	t.Run("invalid - positive amount", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: &acc[0].ID,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_amount must be negative for TRANSACTION_TYPE_WITHDRAWAL")
	})

	t.Run("invalid - no source account", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: nil,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_account_id is required for TRANSACTION_TYPE_WITHDRAWAL")
	})

	t.Run("invalid - no source amount", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NullDecimal{},
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: &acc[0].ID,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_amount is required for TRANSACTION_TYPE_WITHDRAWAL")
	})

	t.Run("invalid - no source currency", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  "",
			SourceAccountID: &acc[0].ID,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_currency is required for TRANSACTION_TYPE_WITHDRAWAL")
	})

	t.Run("invalid - no source account ID", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  acc[0].Currency,
			SourceAccountID: nil,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_account_id is required for TRANSACTION_TYPE_WITHDRAWAL")
	})

	t.Run("invalid - source currency mismatch", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:  acc[1].Currency,
			SourceAccountID: &acc[0].ID,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency USD, expected EUR")
	})

	t.Run("invalid - fx amount without fx currency", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:  gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:     decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:   acc[0].Currency,
			SourceAccountID:  &acc[0].ID,
			FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			FxSourceCurrency: "",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "fx_source_currency is required")
	})

	t.Run("invalid - fx amount with fx currency", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:  gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:     decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:   acc[0].Currency,
			SourceAccountID:  &acc[0].ID,
			FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(100)),
			FxSourceCurrency: acc[1].Currency,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "fx_source_amount must be negative for TRANSACTION_TYPE_WITHDRAWAL")
	})

	t.Run("invalid - destination amount without destination currency", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:     gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:      acc[0].Currency,
			SourceAccountID:     &acc[0].ID,
			DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency: "",
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required")
	})

	t.Run("invalid - destination amount with destination currency", func(t *testing.T) {
		srv := transactions.NewService(nil)

		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:     gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(-55)),
			SourceCurrency:      acc[0].Currency,
			SourceAccountID:     &acc[0].ID,
			DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationCurrency: acc[1].Currency,
		})

		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount must be positive for TRANSACTION_TYPE_WITHDRAWAL")
	})
}

func TestValidateDeposit(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	acc := []*database.Account{
		{Name: "Test USD", Currency: "USD", Extra: map[string]string{}},
		{Name: "Test EUR", Currency: "EUR", Extra: map[string]string{}},
	}
	assert.NoError(t, gormDB.Create(&acc).Error)

	t.Run("valid deposit", func(t *testing.T) {
		srv := transactions.NewService(nil)
		assert.NoError(t, srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: &acc[0].ID,
		}))
	})

	t.Run("invalid - negative amount", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: &acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount must be positive for TRANSACTION_TYPE_DEPOSIT")
	})

	t.Run("invalid - no destination account", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: nil,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_account_id is required for TRANSACTION_TYPE_DEPOSIT")
	})

	t.Run("invalid - no destination amount", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NullDecimal{},
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: &acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount is required for TRANSACTION_TYPE_DEPOSIT")
	})

	t.Run("invalid - no destination currency", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency:  "",
			DestinationAccountID: &acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required for TRANSACTION_TYPE_DEPOSIT")
	})

	t.Run("invalid - destination currency mismatch", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: &acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency USD, expected EUR")
	})
}

func TestValidateReconciliation(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	acc := []*database.Account{
		{Name: "Test USD", Currency: "USD", Extra: map[string]string{}},
	}
	assert.NoError(t, gormDB.Create(&acc).Error)

	t.Run("valid reconciliation", func(t *testing.T) {
		srv := transactions.NewService(nil)
		assert.NoError(t, srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: &acc[0].ID,
		}))
	})

	t.Run("success - negative amount", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-200)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: &acc[0].ID,
		})

		assert.NoError(t, err)
	})

	t.Run("invalid - no destination account", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: nil,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_account_id is required for TRANSACTION_TYPE_RECONCILIATION")
	})

	t.Run("invalid - no destination amount", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NullDecimal{},
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: &acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount is required for TRANSACTION_TYPE_RECONCILIATION")
	})

	t.Run("invalid - no destination currency", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
			DestinationCurrency:  "",
			DestinationAccountID: &acc[0].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required for TRANSACTION_TYPE_RECONCILIATION")
	})

	t.Run("invalid - destination currency mismatch", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(200)),
			DestinationCurrency:  "EUR",
			DestinationAccountID: &acc[0].ID,
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

	t.Run("valid transfer", func(t *testing.T) {
		srv := transactions.NewService(nil)
		assert.NoError(t, srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: &acc[1].ID,
		}))
	})

	t.Run("invalid - no source account", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      nil,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: &acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_account_id is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no destination account", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: nil,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_account_id is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no source amount", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NullDecimal{},
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: &acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_amount is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no destination amount", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NullDecimal{},
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: &acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_amount is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no source currency", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       "",
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: &acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "source_currency is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - no destination currency", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  "",
			DestinationAccountID: &acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_currency is required for TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS")
	})

	t.Run("invalid - source currency mismatch", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[1].Currency,
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[1].Currency,
			DestinationAccountID: &acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency USD, expected EUR")
	})

	t.Run("invalid - destination currency mismatch", func(t *testing.T) {
		srv := transactions.NewService(nil)
		err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
			SourceCurrency:       acc[0].Currency,
			SourceAccountID:      &acc[0].ID,
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
			DestinationCurrency:  acc[0].Currency,
			DestinationAccountID: &acc[1].ID,
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "has currency EUR, expected USD")
	})
}

func TestValidateInvalidType(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	srv := transactions.NewService(nil)

	err := srv.ValidateTransaction(context.TODO(), gormDB, &database.Transaction{
		TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED,
	})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "unsupported transaction type: TRANSACTION_TYPE_UNSPECIFIED")
}
