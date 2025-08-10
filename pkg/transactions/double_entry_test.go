package transactions_test

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
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

		accSvc.EXPECT().GetAccount(gomock.Any(), sourceAccountID).Return(&database.Account{
			Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_REGULAR,
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

		accSvc.EXPECT().GetAccount(gomock.Any(), sourceAccountID).Return(&database.Account{
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
}
