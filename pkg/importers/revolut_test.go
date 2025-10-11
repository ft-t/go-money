package importers_test

import (
	"context"
	"encoding/base64"
	_ "embed"
	"testing"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/revolut/simple_expense.csv
var revolutExpense []byte

//go:embed testdata/revolut/exchange.csv
var revolutExchange []byte

//go:embed testdata/revolut/exchange_swap.csv
var revolutExchangeSwap []byte

func TestRevolutSimple_Success(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	records := []*importers.Record{
		{
			Data: revolutExpense,
		},
	}

	parsedRecords, err := srv.ParseMessages(context.TODO(), records)
	require.NoError(t, err)
	require.NotNil(t, parsedRecords)

	txs := parsedRecords
	require.Len(t, txs, 1)

	assert.EqualValues(t, "2024-09-02 10:31:35", txs[0].Date.Format(time.DateTime))
	assert.EqualValues(t, "TRANSFER.To XXYYZZ", txs[0].Description)
	assert.EqualValues(t, "USD", txs[0].SourceCurrency)
	assert.EqualValues(t, "21.31", txs[0].SourceAmount.StringFixed(2))
}

func TestRevolutExchange_Success(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	records := []*importers.Record{
		{
			Data: revolutExchange,
		},
	}

	parsedRecords, err := srv.ParseMessages(context.TODO(), records)
	require.NoError(t, err)
	require.NotNil(t, parsedRecords)

	txs := parsedRecords
	require.Len(t, txs, 1)

	assert.EqualValues(t, "2024-10-25 11:48:00", txs[0].Date.Format(time.DateTime))
	assert.EqualValues(t, "EXCHANGE.Exchanged to PLN", txs[0].Description)

	assert.EqualValues(t, "USD", txs[0].SourceCurrency)
	assert.EqualValues(t, "revolut_USD", txs[0].SourceAccount)
	assert.EqualValues(t, "469.57", txs[0].SourceAmount.StringFixed(2))

	assert.EqualValues(t, "PLN", txs[0].DestinationCurrency)
	assert.EqualValues(t, "revolut_PLN", txs[0].DestinationAccount)
	assert.EqualValues(t, "1907.07", txs[0].DestinationAmount.StringFixed(2))
}

func TestRevolutExchangeSwap_Success(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	records := []*importers.Record{
		{
			Data: revolutExchangeSwap,
		},
	}

	parsedRecords, err := srv.ParseMessages(context.TODO(), records)
	require.NoError(t, err)
	require.NotNil(t, parsedRecords)

	txs := parsedRecords
	require.Len(t, txs, 1)

	assert.EqualValues(t, "2024-10-25 11:48:00", txs[0].Date.Format(time.DateTime))
	assert.EqualValues(t, "EXCHANGE.Exchanged to PLN", txs[0].Description)

	assert.EqualValues(t, "USD", txs[0].SourceCurrency)
	assert.EqualValues(t, "revolut_USD", txs[0].SourceAccount)
	assert.EqualValues(t, "469.57", txs[0].SourceAmount.StringFixed(2))

	assert.EqualValues(t, "PLN", txs[0].DestinationCurrency)
	assert.EqualValues(t, "revolut_PLN", txs[0].DestinationAccount)
	assert.EqualValues(t, "1907.07", txs[0].DestinationAmount.StringFixed(2))
}

func TestRevolutType(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	assert.Equal(t, importv1.ImportSource_IMPORT_SOURCE_REVOLUT, srv.Type())
}

func TestRevolutParse_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	base := importers.NewBaseParser(currencyConverter, txSvc, mapperSvc)
	srv := importers.NewRevolut(base)

	account1 := &database.Account{
		ID:            1,
		Name:          "Test Account",
		Currency:      "USD",
		AccountNumber: "revolut_USD",
		Type:          gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	}

	expenseAccount := &database.Account{
		ID:       2,
		Name:     "Default Expense",
		Currency: "EUR",
		Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Flags:    database.AccountFlagIsDefault,
	}

	accounts := []*database.Account{account1, expenseAccount}

	csvData := []byte("Type,Product,Started Date,Completed Date,Description,Amount,Fee,Currency,State,Balance\nTRANSFER,Current,2024-09-02 10:31:35,2024-09-02 10:31:35,To XXYYZZ,-21.31,0,USD,COMPLETED,100.00")

	currencyConverter.EXPECT().
		Convert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, from string, to string, amount decimal.Decimal) (decimal.Decimal, error) {
			assert.Equal(t, "USD", from)
			assert.Equal(t, "EUR", to)
			return amount, nil
		}).
		Times(1)

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(csvData)},
			Accounts: accounts,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.CreateRequests, 1)
	assert.Equal(t, "TRANSFER.To XXYYZZ", resp.CreateRequests[0].Title)
}

func TestRevolutParse_EmptyFile(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	emptyCSV := []byte("Type,Product,Started Date,Completed Date,Description,Amount,Fee,Currency,State,Balance\n")

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(emptyCSV)},
			Accounts: nil,
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty file")
	assert.Nil(t, resp)
}

func TestRevolutParse_InvalidBase64(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{"invalid-base64!!!"},
			Accounts: nil,
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode files")
	assert.Nil(t, resp)
}

func TestRevolutParseMessages_EmptyCSV(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	records := []*importers.Record{
		{
			Data: []byte{},
		},
	}

	parsedRecords, err := srv.ParseMessages(context.TODO(), records)
	require.NoError(t, err)
	require.Len(t, parsedRecords, 1)
	assert.Error(t, parsedRecords[0].ParsingError)
	assert.Contains(t, parsedRecords[0].ParsingError.Error(), "empty CSV data")
}

func TestRevolutParseMessages_InvalidData(t *testing.T) {
	srv := importers.NewRevolut(importers.NewBaseParser(nil, nil, nil))

	records := []*importers.Record{
		{
			Data: []byte("invalid csv data"),
		},
	}

	parsedRecords, err := srv.ParseMessages(context.TODO(), records)
	require.NoError(t, err)
	require.Len(t, parsedRecords, 1)
	assert.Error(t, parsedRecords[0].ParsingError)
}

func TestRevolutParse_SplitCsvError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	base := importers.NewBaseParser(currencyConverter, txSvc, mapperSvc)
	srv := importers.NewRevolut(base)

	csvData := []byte("Type,Product,Started Date,Completed Date,Description,Amount,Fee,Currency,State,Balance")

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(csvData)},
			Accounts: []*database.Account{},
		},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to split csv")
}

func TestRevolutParse_GetAccountMapError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	base := importers.NewBaseParser(currencyConverter, txSvc, mapperSvc)
	srv := importers.NewRevolut(base)

	account1 := &database.Account{
		ID:            1,
		AccountNumber: "duplicate",
	}
	account2 := &database.Account{
		ID:            2,
		AccountNumber: "duplicate",
	}

	csvData := []byte("Type,Product,Started Date,Completed Date,Description,Amount,Fee,Currency,State,Balance\nTRANSFER,Current,2024-09-02 10:31:35,2024-09-02 10:31:35,To XXYYZZ,-21.31,0,USD,COMPLETED,100.00")

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(csvData)},
			Accounts: []*database.Account{account1, account2},
		},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get account map by numbers")
}
