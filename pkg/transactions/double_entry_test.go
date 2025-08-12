package transactions_test

import (
	"context"
	"testing"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDoubleEntry_Withdrawals(t *testing.T) {
	baseCurrency := "USD"
	sourceAccountID := int32(1)
	destinationAccountID := int32(2)

	t.Run("basic expense", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})

		accSvc.EXPECT().GetAccountByID(gomock.Any(), sourceAccountID).Return(&database.Account{
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		}, nil)

		resp, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Coffee",
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("expense from credit account", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})

		accSvc.EXPECT().GetAccountByID(gomock.Any(), sourceAccountID).Return(&database.Account{
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY,
		}, nil)

		resp, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Coffee",
		})
		assert.NoError(t, err)

		assert.True(t, resp[0].IsDebit)
		assert.False(t, resp[1].IsDebit)
	})

	t.Run("income", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})

		accSvc.EXPECT().GetAccountByID(gomock.Any(), sourceAccountID).Return(&database.Account{
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME,
		}, nil)

		resp, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			Title:                           "Salary",
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("transfer between accounts", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})

		accSvc.EXPECT().GetAccountByID(gomock.Any(), sourceAccountID).Return(&database.Account{
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		}, nil)

		resp, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Transfer",
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("transfer to credit account", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})

		accSvc.EXPECT().GetAccountByID(gomock.Any(), sourceAccountID).Return(&database.Account{
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		}, nil)

		resp, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Transfer to credit account",
		})
		assert.NoError(t, err)

		assert.False(t, resp[0].IsDebit)
		assert.True(t, resp[1].IsDebit)
	})

	t.Run("transfer from credit account", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})

		accSvc.EXPECT().GetAccountByID(gomock.Any(), sourceAccountID).Return(&database.Account{
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY,
		}, nil)

		resp, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Transfer from credit account",
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

	t.Run("amount miss match", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})
		_, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(200)),
			Title:                           "Test",
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "source and destination amounts in base currency must be equal for double entry transactions")
	})

	t.Run("amount signs match", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})
		_, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Test",
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "source and destination amounts must have opposite signs for double entry transactions")
	})

	t.Run("source account is not set", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})
		_, err := srv.Record(context.TODO(), &database.Transaction{
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Test",
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "source_account_id is required for double entry transactions")
	})

	t.Run("destination account is not set", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})
		_, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Test",
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "destination_account_id is required for double entry transactions")
	})

	t.Run("get account error", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		srv := transactions.NewDoubleEntryService(&transactions.DoubleEntryConfig{
			BaseCurrency: baseCurrency,
			AccountSvc:   accSvc,
		})

		accSvc.EXPECT().GetAccountByID(gomock.Any(), sourceAccountID).Return(nil, assert.AnError)

		_, err := srv.Record(context.TODO(), &database.Transaction{
			SourceAccountID:                 &sourceAccountID,
			DestinationAccountID:            &destinationAccountID,
			SourceAmountInBaseCurrency:      decimal.NewNullDecimal(decimal.NewFromInt(-100)),
			DestinationAmountInBaseCurrency: decimal.NewNullDecimal(decimal.NewFromInt(100)),
			Title:                           "Test",
		})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to get source account")
	})
}
