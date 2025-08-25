package double_entry_test

import (
	"context"
	"os"
	"testing"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/double_entry"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestDoubleEntry_Withdrawals(t *testing.T) {
	baseCurrency := "USD"
	sourceAccountID := int32(1)
	destinationAccountID := int32(2)

	t.Run("basic expense", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		sourceAccount := &database.Account{
			ID:   sourceAccountID,
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		}

		resp, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Coffee",
			},
			SourceAccount: sourceAccount,
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("expense from credit account", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		sourceAccount := &database.Account{
			ID:   sourceAccountID,
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY,
		}

		resp, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Coffee",
			},
			SourceAccount: sourceAccount,
		})
		assert.NoError(t, err)

		assert.True(t, resp[0].IsDebit)
		assert.False(t, resp[1].IsDebit)
	})

	t.Run("income", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		sourceAccount := &database.Account{
			ID:   sourceAccountID,
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME,
		}

		resp, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				Title:                           "Salary",
			},
			SourceAccount: sourceAccount,
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("transfer between accounts", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		sourceAccount := &database.Account{
			ID:   sourceAccountID,
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		}

		resp, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Transfer",
			},
			SourceAccount: sourceAccount,
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("transfer to credit account", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})
		sourceAccount := &database.Account{
			ID:   sourceAccountID,
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		}

		resp, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Transfer to credit account",
			},
			SourceAccount: sourceAccount,
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("transfer from credit account", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		sourceAccount := &database.Account{
			ID:   sourceAccountID,
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY,
		}

		resp, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Transfer from credit account",
			},
			SourceAccount: sourceAccount,
		})
		assert.NoError(t, err)

		assert.True(t, resp[0].IsDebit)
		assert.False(t, resp[1].IsDebit)
	})
}

func TestDoubleEntry(t *testing.T) {
	baseCurrency := "USD"
	sourceAccountID := int32(1)
	destinationAccountID := int32(2)

	sourceAccount := &database.Account{
		ID: sourceAccountID,
	}

	t.Run("amount miss match", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		_, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(200)),
				Title:                           "Test",
			},
			SourceAccount: sourceAccount,
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "source and destination amounts in base currency must be equal for double entry transactions")
	})

	t.Run("amount signs match", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		_, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Test",
			},
			SourceAccount: sourceAccount,
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "source and destination amounts must have opposite signs for double entry transactions")
	})

	t.Run("source account is not set", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		_, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Test",
			},
			SourceAccount: sourceAccount,
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "source_account_id is required for double entry transactions")
	})

	t.Run("destination account is not set", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		_, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Test",
			},
			SourceAccount: sourceAccount,
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "destination_account_id is required for double entry transactions")
	})

	t.Run("get account error", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
		})

		_, err := srv.Calculate(context.TODO(), &double_entry.RecordRequest{
			Transaction: &database.Transaction{
				SourceAccountID:                 sourceAccountID,
				DestinationAccountID:            destinationAccountID,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
				Title:                           "Test",
			},
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "source account is required for double entry transactions")
	})
}

func TestRecord(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: "USD",
		})

		testRecords := []*database.DoubleEntry{
			{
				TransactionID: 55,
				IsDebit:       false,
				BaseCurrency:  "any",
			},
			{
				TransactionID: 55,
				IsDebit:       true,
				BaseCurrency:  "any",
			},
		}
		assert.NoError(t, gormDB.Create(&testRecords).Error)

		err := srv.Record(context.TODO(), gormDB, []*database.Transaction{
			{
				ID:                              55,
				SourceAccountID:                 1,
				DestinationAccountID:            2,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			},
		}, map[int32]*database.Account{
			1: {
				ID:   1,
				Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		})
		assert.NoError(t, err)

		var updatedRecords []*database.DoubleEntry
		assert.NoError(t, gormDB.Where("transaction_id = ?", 55).Find(&updatedRecords).Error)

		assert.Len(t, updatedRecords, 2)

		assert.Equal(t, "USD", updatedRecords[0].BaseCurrency)
		assert.Equal(t, "USD", updatedRecords[1].BaseCurrency)
		assert.EqualValues(t, decimal.NewFromInt(100), updatedRecords[0].AmountInBaseCurrency)
	})

	t.Run("no records", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: "USD",
		})

		err := srv.Record(context.TODO(), nil, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("missing destination account", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: "USD",
		})

		err := srv.Record(context.TODO(), nil, []*database.Transaction{
			{
				ID:                              55,
				SourceAccountID:                 1,
				DestinationAccountID:            0,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			},
		}, map[int32]*database.Account{
			1: {
				ID:   1,
				Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "destination_account_id is required for double entry transactions")
	})

	t.Run("missing source account", func(t *testing.T) {
		srv := double_entry.NewDoubleEntryService(&double_entry.DoubleEntryConfig{
			BaseCurrency: "USD",
		})

		err := srv.Record(context.TODO(), nil, []*database.Transaction{
			{
				ID:                              55,
				SourceAccountID:                 0,
				DestinationAccountID:            2,
				SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			},
		}, map[int32]*database.Account{
			2: {
				ID:   2,
				Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "source account not found for double entry transaction")
	})
}
