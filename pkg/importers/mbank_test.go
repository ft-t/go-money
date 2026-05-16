package importers_test

import (
	"context"
	"encoding/base64"
	_ "embed"
	"testing"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/mbank/expense.csv
var mbankExpense []byte

//go:embed testdata/mbank/income.csv
var mbankIncome []byte

//go:embed testdata/mbank/mixed.csv
var mbankMixed []byte

//go:embed testdata/mbank/bad_date.csv
var mbankBadDate []byte

//go:embed testdata/mbank/bad_amount.csv
var mbankBadAmount []byte

//go:embed testdata/mbank/bad_amount_no_cur.csv
var mbankBadAmountNoCur []byte

//go:embed testdata/mbank/no_account.csv
var mbankNoAccount []byte

const mbankAccountNumber = "00000000000000000000000001"

func TestMbank_Type(t *testing.T) {
	srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))
	assert.Equal(t, importv1.ImportSource_IMPORT_SOURCE_MBANK, srv.Type())
}

func TestMbank_ExpenseSuccess(t *testing.T) {
	type tc struct {
		name        string
		data        []byte
		wantAmt     string
		wantCur     string
		wantSrcAcc  string
		wantDate    string
		wantDesc    string
		wantCategry string
	}

	cases := []tc{
		{
			name:        "card expense",
			data:        mbankExpense,
			wantAmt:     "44.91",
			wantCur:     "PLN",
			wantSrcAcc:  mbankAccountNumber,
			wantDate:    "2026-05-16 00:00:00 +0000",
			wantDesc:    "FAKE SHOP ZAKUP PRZY UŻYCIU KARTY W KRAJU transakcja nierozliczona",
			wantCategry: "Wyjścia i wydarzenia",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: c.data}})
			require.NoError(t, err)
			require.Len(t, resp, 1)

			tx := resp[0]
			require.NoError(t, tx.ParsingError)
			assert.Equal(t, importers.TransactionTypeExpense, tx.Type)
			assert.Equal(t, c.wantAmt, tx.SourceAmount.StringFixed(2))
			assert.Equal(t, c.wantCur, tx.SourceCurrency)
			assert.Equal(t, c.wantAmt, tx.DestinationAmount.StringFixed(2))
			assert.Equal(t, c.wantCur, tx.DestinationCurrency)
			assert.Equal(t, c.wantSrcAcc, tx.SourceAccount)
			assert.Equal(t, "", tx.DestinationAccount)
			assert.Equal(t, c.wantDate, tx.Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, c.wantDesc, tx.Description)
			assert.Equal(t, c.wantCategry, tx.OriginalTxType)
			require.Len(t, tx.DeduplicationKeys, 1)
			assert.NotEmpty(t, tx.DeduplicationKeys[0])
		})
	}
}

func TestMbank_IncomeSuccess(t *testing.T) {
	srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: mbankIncome}})
	require.NoError(t, err)
	require.Len(t, resp, 1)

	tx := resp[0]
	require.NoError(t, tx.ParsingError)
	assert.Equal(t, importers.TransactionTypeIncome, tx.Type)
	assert.Equal(t, "2000.00", tx.DestinationAmount.StringFixed(2))
	assert.Equal(t, "PLN", tx.DestinationCurrency)
	assert.Equal(t, "2000.00", tx.SourceAmount.StringFixed(2))
	assert.Equal(t, "PLN", tx.SourceCurrency)
	assert.Equal(t, mbankAccountNumber, tx.DestinationAccount)
	assert.Equal(t, "", tx.SourceAccount)
	assert.Equal(t, "FAKE SENDER PRZELEW ZEWNĘTRZNY PRZYCHODZĄCY 00000000000000000000000002", tx.Description)
	assert.Equal(t, "Wpływy - inne", tx.OriginalTxType)
}

func TestMbank_MixedSuccess(t *testing.T) {
	srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: mbankMixed}})
	require.NoError(t, err)
	require.Len(t, resp, 3)

	assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)
	assert.Equal(t, "65.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, mbankAccountNumber, resp[0].DestinationAccount)

	assert.Equal(t, importers.TransactionTypeIncome, resp[1].Type)
	assert.Equal(t, "2000.00", resp[1].DestinationAmount.StringFixed(2))
	assert.Equal(t, mbankAccountNumber, resp[1].DestinationAccount)

	assert.Equal(t, importers.TransactionTypeExpense, resp[2].Type)
	assert.Equal(t, "44.91", resp[2].SourceAmount.StringFixed(2))
	assert.Equal(t, mbankAccountNumber, resp[2].SourceAccount)
}

func TestMbank_Failure(t *testing.T) {
	type tc struct {
		name    string
		data    []byte
		wantErr string
	}

	cases := []tc{
		{name: "bad date", data: mbankBadDate, wantErr: "failed to parse date"},
		{name: "bad amount", data: mbankBadAmount, wantErr: "failed to parse amount"},
		{name: "amount missing currency", data: mbankBadAmountNoCur, wantErr: "failed to parse amount"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: c.data}})
			require.NoError(t, err)
			require.Len(t, resp, 1)

			require.Error(t, resp[0].ParsingError)
			assert.Contains(t, resp[0].ParsingError.Error(), c.wantErr)
		})
	}
}

func TestMbank_NoHeader(t *testing.T) {
	srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: []byte("not;an;mbank;file\r\n")}})
	require.NoError(t, err)
	require.Len(t, resp, 1)

	require.Error(t, resp[0].ParsingError)
	assert.Contains(t, resp[0].ParsingError.Error(), "header row not found")
}

func TestMbank_ParseDecodeError(t *testing.T) {
	srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

	_, err := srv.Parse(context.Background(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data: []string{"!!!not-base64!!!"},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode")
}

func TestMbank_NoAccountNumberStillParses(t *testing.T) {
	srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: mbankNoAccount}})
	require.NoError(t, err)
	require.Len(t, resp, 1)

	tx := resp[0]
	require.NoError(t, tx.ParsingError)
	assert.Equal(t, importers.TransactionTypeExpense, tx.Type)
	assert.Equal(t, "", tx.SourceAccount)
	assert.Equal(t, "12.34", tx.SourceAmount.StringFixed(2))
}

func TestMbankParse_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	base := importers.NewBaseParser(currencyConverter, txSvc, mapperSvc)
	srv := importers.NewMbank(base)

	mbankAccount := &database.Account{
		ID:            1,
		Name:          "MBank PLN",
		Currency:      "PLN",
		AccountNumber: mbankAccountNumber,
		Type:          gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	}

	expenseAccount := &database.Account{
		ID:       2,
		Name:     "Default Expense",
		Currency: "USD",
		Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Flags:    database.AccountFlagIsDefault,
	}

	currencyConverter.EXPECT().
		Convert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, from string, to string, amount decimal.Decimal) (decimal.Decimal, error) {
			assert.Equal(t, "PLN", from)
			assert.Equal(t, "USD", to)
			return amount, nil
		}).
		Times(1)

	resp, err := srv.Parse(context.Background(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(mbankExpense)},
			Accounts: []*database.Account{mbankAccount, expenseAccount},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.CreateRequests, 1)
	assert.Equal(t, "FAKE SHOP ZAKUP PRZY UŻYCIU KARTY W KRAJU transakcja nierozliczona", resp.CreateRequests[0].Title)
}

func TestMbankParse_GetAccountMapError(t *testing.T) {
	srv := importers.NewMbank(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.Parse(context.Background(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data: []string{base64.StdEncoding.EncodeToString(mbankExpense)},
			Accounts: []*database.Account{
				{ID: 1, AccountNumber: "duplicate"},
				{ID: 2, AccountNumber: "duplicate"},
			},
		},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get account map by numbers")
}
