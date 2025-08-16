package transactions_test

import (
	"context"
	"testing"

	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestGetAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		acc := NewMockAccountSvc(gomock.NewController(t))

		svc := transactions.NewApplicableAccountService(acc)

		accounts := []*database.Account{
			{ID: 1, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET},
			{ID: 2, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET},
			{ID: 3, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET},
			{ID: 4, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY},
			{ID: 5, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME},
			{ID: 6, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE},
		}

		acc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

		res, err := svc.GetAll(context.TODO())
		assert.NoError(t, err)

		transfer := res[gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS]
		assert.ElementsMatch(t, []int32{1, 2, 3, 4}, lo.Keys(transfer.SourceAccounts))
		assert.ElementsMatch(t, []int32{1, 2, 3, 4}, lo.Keys(transfer.DestinationAccounts))
	})

	t.Run("error", func(t *testing.T) {
		acc := NewMockAccountSvc(gomock.NewController(t))

		svc := transactions.NewApplicableAccountService(acc)

		acc.EXPECT().GetAllAccounts(gomock.Any()).Return(nil, assert.AnError)

		res, err := svc.GetAll(context.TODO())
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestService_getPossibleAccountsForTransactionType(t *testing.T) {
	svc := transactions.NewApplicableAccountService(nil)

	accounts := []*database.Account{
		{ID: 1, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET},
		{ID: 2, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET},
		{ID: 3, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET},
		{ID: 4, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY},
		{ID: 5, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME},
		{ID: 6, Type: gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE},
	}

	res := svc.GetApplicableAccounts(context.TODO(), accounts)

	transfer := res[gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS]
	assert.ElementsMatch(t, []int32{1, 2, 3, 4}, lo.Keys(transfer.SourceAccounts))
	assert.ElementsMatch(t, []int32{1, 2, 3, 4}, lo.Keys(transfer.DestinationAccounts))

	deposit := res[gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME]
	assert.ElementsMatch(t, []int32{5}, lo.Keys(deposit.SourceAccounts))
	assert.ElementsMatch(t, []int32{1, 2, 3}, lo.Keys(deposit.DestinationAccounts))

	withdrawal := res[gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE]
	assert.ElementsMatch(t, []int32{1, 2, 3, 4}, lo.Keys(withdrawal.SourceAccounts))
	assert.ElementsMatch(t, []int32{6}, lo.Keys(withdrawal.DestinationAccounts))
}
