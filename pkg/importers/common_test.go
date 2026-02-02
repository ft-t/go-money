package importers_test

import (
	"context"
	"testing"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAccountMapByNumbers(t *testing.T) {
	bp := importers.NewBaseParser(nil, nil, nil)

	t.Run("single account with single number", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int32(1), result["1234"].ID)
	})

	t.Run("single account with multiple numbers", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234,5678",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int32(1), result["1234"].ID)
		assert.Equal(t, int32(1), result["5678"].ID)
	})

	t.Run("multiple accounts with different numbers", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				ID:            2,
				AccountNumber: "5678",
				Currency:      "USD",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int32(1), result["1234"].ID)
		assert.Equal(t, int32(2), result["5678"].ID)
	})

	t.Run("account with empty number gets uuid", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		var foundAccount *database.Account
		for _, acc := range result {
			foundAccount = acc
			break
		}
		assert.NotNil(t, foundAccount)
		assert.Equal(t, int32(1), foundAccount.ID)
	})

	t.Run("account with whitespace-only number gets uuid", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "  ",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("duplicate account numbers returns error", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				ID:            2,
				AccountNumber: "1234",
				Currency:      "USD",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "duplicate account number: 1234")
	})

	t.Run("account with comma-separated numbers and whitespace", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234 , 5678 , 9012",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, int32(1), result["1234"].ID)
		assert.Equal(t, int32(1), result["5678"].ID)
		assert.Equal(t, int32(1), result["9012"].ID)
	})

	t.Run("empty accounts list", func(t *testing.T) {
		accounts := []*database.Account{}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestToCreateRequests_ParsingError_Success(t *testing.T) {
	testCases := []struct {
		name         string
		transaction  *importers.Transaction
		expectedErr  string
	}{
		{
			name: "expense with parsing error skips expense processing",
			transaction: &importers.Transaction{
				ID:                  "test-1",
				Type:                importers.TransactionTypeExpense,
				SourceAmount:        decimal.Zero,
				SourceCurrency:      "PLN",
				DestinationAmount:   decimal.Zero,
				DestinationCurrency: "PLN",
				Date:                time.Now(),
				Description:         "Pending transaction",
				ParsingError:        errors.New("transaction is still pending"),
			},
			expectedErr: "transaction is still pending",
		},
		{
			name: "income with parsing error skips income processing",
			transaction: &importers.Transaction{
				ID:                  "test-2",
				Type:                importers.TransactionTypeIncome,
				SourceAmount:        decimal.Zero,
				SourceCurrency:      "PLN",
				DestinationAmount:   decimal.Zero,
				DestinationCurrency: "PLN",
				Date:                time.Now(),
				Description:         "Invalid income",
				ParsingError:        errors.New("invalid amount format"),
			},
			expectedErr: "invalid amount format",
		},
		{
			name: "transfer with parsing error skips transfer processing",
			transaction: &importers.Transaction{
				ID:                  "test-3",
				Type:                importers.TransactionTypeInternalTransfer,
				SourceAmount:        decimal.Zero,
				SourceCurrency:      "PLN",
				DestinationAmount:   decimal.Zero,
				DestinationCurrency: "PLN",
				Date:                time.Now(),
				Description:         "Invalid transfer",
				ParsingError:        errors.New("missing account"),
			},
			expectedErr: "missing account",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			currencyConverter := NewMockCurrencyConverterSvc(ctrl)
			bp := importers.NewBaseParser(currencyConverter, nil, nil)

			accountMap := map[string]*database.Account{
				"test-account": {
					ID:       1,
					Currency: "PLN",
					Type:     v1.AccountType_ACCOUNT_TYPE_ASSET,
				},
			}

			requests, err := bp.ToCreateRequests(
				context.Background(),
				[]*importers.Transaction{tc.transaction},
				false,
				accountMap,
				importv1.ImportSource_IMPORT_SOURCE_BNP_PARIBAS_POLSKA,
			)

			require.NoError(t, err)
			require.Len(t, requests, 1)

			req := requests[0]
			assert.Nil(t, req.Transaction)
			assert.Equal(t, tc.expectedErr, req.Extra["parsing_error"])
			assert.Equal(t, tc.transaction.Description, req.Title)
		})
	}
}

func TestToCreateRequests_UnknownType_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	bp := importers.NewBaseParser(currencyConverter, nil, nil)

	transaction := &importers.Transaction{
		ID:                  "test-unknown",
		Type:                importers.TransactionType(99),
		SourceAmount:        decimal.NewFromInt(100),
		SourceCurrency:      "PLN",
		DestinationAmount:   decimal.NewFromInt(100),
		DestinationCurrency: "PLN",
		Date:                time.Now(),
		Description:         "Unknown type transaction",
	}

	accountMap := map[string]*database.Account{
		"test-account": {
			ID:       1,
			Currency: "PLN",
			Type:     v1.AccountType_ACCOUNT_TYPE_ASSET,
		},
	}

	requests, err := bp.ToCreateRequests(
		context.Background(),
		[]*importers.Transaction{transaction},
		false,
		accountMap,
		importv1.ImportSource_IMPORT_SOURCE_BNP_PARIBAS_POLSKA,
	)

	require.NoError(t, err)
	require.Len(t, requests, 1)

	req := requests[0]
	assert.Nil(t, req.Transaction)
	assert.Equal(t, "99", req.Extra["unknown_type"])
}

func TestToCreateRequests_ExpenseWithValidData_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	currencyConverter.EXPECT().
		Convert(gomock.Any(), "PLN", "EUR", gomock.Any()).
		Return(decimal.NewFromInt(25), nil).
		Times(1)

	bp := importers.NewBaseParser(currencyConverter, nil, nil)

	transaction := &importers.Transaction{
		ID:                  "test-expense",
		Type:                importers.TransactionTypeExpense,
		SourceAmount:        decimal.NewFromInt(100),
		SourceCurrency:      "PLN",
		SourceAccount:       "source-acc",
		DestinationAmount:   decimal.NewFromInt(100),
		DestinationCurrency: "PLN",
		Date:                time.Now(),
		Description:         "Valid expense",
	}

	accountMap := map[string]*database.Account{
		"source-acc": {
			ID:       1,
			Currency: "PLN",
			Type:     v1.AccountType_ACCOUNT_TYPE_ASSET,
		},
		"default-expense": {
			ID:       2,
			Currency: "EUR",
			Type:     v1.AccountType_ACCOUNT_TYPE_EXPENSE,
			Flags:    database.AccountFlagIsDefault,
		},
	}

	requests, err := bp.ToCreateRequests(
		context.Background(),
		[]*importers.Transaction{transaction},
		false,
		accountMap,
		importv1.ImportSource_IMPORT_SOURCE_BNP_PARIBAS_POLSKA,
	)

	require.NoError(t, err)
	require.Len(t, requests, 1)

	req := requests[0]
	require.NotNil(t, req.Transaction)

	expense, ok := req.Transaction.(*transactionsv1.CreateTransactionRequest_Expense)
	require.True(t, ok)
	assert.Equal(t, int32(1), expense.Expense.SourceAccountId)
	assert.Equal(t, "-100", expense.Expense.SourceAmount)
	assert.Equal(t, "PLN", expense.Expense.SourceCurrency)
	assert.Equal(t, int32(2), expense.Expense.DestinationAccountId)
	assert.Equal(t, "25", expense.Expense.DestinationAmount)
	assert.Equal(t, "EUR", expense.Expense.DestinationCurrency)
	assert.Empty(t, req.Extra["parsing_error"])
}

func TestDecodeFiles(t *testing.T) {
	bp := importers.NewBaseParser(nil, nil, nil)

	t.Run("valid base64", func(t *testing.T) {
		data := []string{"SGVsbG8gV29ybGQ="}
		result, err := bp.DecodeFiles(data)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Hello World", string(result[0]))
	})

	t.Run("invalid base64", func(t *testing.T) {
		data := []string{"!!!invalid!!!"}
		result, err := bp.DecodeFiles(data)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("multiple files", func(t *testing.T) {
		data := []string{"Zmlyc3Q=", "c2Vjb25k"}
		result, err := bp.DecodeFiles(data)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "first", string(result[0]))
		assert.Equal(t, "second", string(result[1]))
	})
}
