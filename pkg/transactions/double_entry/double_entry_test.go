package double_entry_test

import (
	"context"
	"testing"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/double_entry"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

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
