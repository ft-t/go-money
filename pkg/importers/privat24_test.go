package importers_test

import (
	"context"
	_ "embed"
	"testing"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/privat24/expense_same_currency.xlsx
var privat24ExpenseSameCurrency []byte

//go:embed testdata/privat24/expense_fx.xlsx
var privat24ExpenseFx []byte

//go:embed testdata/privat24/income.xlsx
var privat24Income []byte

//go:embed testdata/privat24/mixed.xlsx
var privat24Mixed []byte

//go:embed testdata/privat24/unknown_card.xlsx
var privat24UnknownCard []byte

//go:embed testdata/privat24/bad_date.xlsx
var privat24BadDate []byte

//go:embed testdata/privat24/bad_amount.xlsx
var privat24BadAmount []byte

func TestPrivat24_Type(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))
	assert.Equal(t, importv1.ImportSource_IMPORT_SOURCE_PRIVATE_24, srv.Type())
}

func TestPrivat24_ExpenseSuccess(t *testing.T) {
	type tc struct {
		name          string
		data          []byte
		wantSrcAmt    string
		wantSrcCur    string
		wantDestAmt   string
		wantDestCur   string
		wantSrcAcc    string
		wantDate      string
		wantDateShort string
		wantDesc      string
		wantCategory  string
	}

	cases := []tc{
		{
			name:          "same currency",
			data:          privat24ExpenseSameCurrency,
			wantSrcAmt:    "123.45",
			wantSrcCur:    "UAH",
			wantDestAmt:   "123.45",
			wantDestCur:   "UAH",
			wantSrcAcc:    "1111********2222",
			wantDate:      "2026-01-15 12:30:00 +0000",
			wantDateShort: "12:30",
			wantDesc:      "FAKE MARKET, KYIV",
			wantCategory:  "Супермаркети та продукти",
		},
		{
			name:          "fx expense",
			data:          privat24ExpenseFx,
			wantSrcAmt:    "410.20",
			wantSrcCur:    "UAH",
			wantDestAmt:   "40.00",
			wantDestCur:   "PLN",
			wantSrcAcc:    "1111********2222",
			wantDate:      "2026-01-16 09:15:00 +0000",
			wantDateShort: "09:15",
			wantDesc:      "FAKE FUEL, WROCLAW",
			wantCategory:  "Авто",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: c.data}})
			require.NoError(t, err)
			require.Len(t, resp, 1)

			tx := resp[0]
			require.NoError(t, tx.ParsingError)
			assert.Equal(t, importers.TransactionTypeExpense, tx.Type)
			assert.Equal(t, c.wantSrcAmt, tx.SourceAmount.StringFixed(2))
			assert.Equal(t, c.wantSrcCur, tx.SourceCurrency)
			assert.Equal(t, c.wantDestAmt, tx.DestinationAmount.StringFixed(2))
			assert.Equal(t, c.wantDestCur, tx.DestinationCurrency)
			assert.Equal(t, c.wantSrcAcc, tx.SourceAccount)
			assert.Equal(t, "", tx.DestinationAccount)
			assert.Equal(t, c.wantDate, tx.Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, c.wantDateShort, tx.DateFromMessage)
			assert.Equal(t, c.wantDesc, tx.Description)
			assert.Equal(t, c.wantCategory, tx.OriginalTxType)
			require.Len(t, tx.DeduplicationKeys, 1)
			assert.NotEmpty(t, tx.DeduplicationKeys[0])
		})
	}
}

func TestPrivat24_IncomeSuccess(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: privat24Income}})
	require.NoError(t, err)
	require.Len(t, resp, 1)

	tx := resp[0]
	require.NoError(t, tx.ParsingError)
	assert.Equal(t, importers.TransactionTypeIncome, tx.Type)
	assert.Equal(t, "5000.00", tx.DestinationAmount.StringFixed(2))
	assert.Equal(t, "UAH", tx.DestinationCurrency)
	assert.Equal(t, "5000.00", tx.SourceAmount.StringFixed(2))
	assert.Equal(t, "UAH", tx.SourceCurrency)
	assert.Equal(t, "1111********2222", tx.DestinationAccount)
	assert.Equal(t, "", tx.SourceAccount)
	assert.Equal(t, "FAKE SALARY", tx.Description)
	assert.Equal(t, "Зарахування", tx.OriginalTxType)
}

func TestPrivat24_MixedSuccess(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: privat24Mixed}})
	require.NoError(t, err)
	require.Len(t, resp, 3)

	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
	assert.Equal(t, "1111********2222", resp[0].SourceAccount)
	assert.Equal(t, "Переказ на свою картку", resp[0].OriginalTxType)

	assert.Equal(t, importers.TransactionTypeIncome, resp[1].Type)
	assert.Equal(t, "3333********4444", resp[1].DestinationAccount)
	assert.Equal(t, "Зарахування зі своєї картки", resp[1].OriginalTxType)

	assert.Equal(t, importers.TransactionTypeExpense, resp[2].Type)
	assert.Equal(t, "3333********4444", resp[2].SourceAccount)
	assert.Equal(t, "Таксі", resp[2].OriginalTxType)
}

func TestPrivat24_UnknownCardStillParses(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: privat24UnknownCard}})
	require.NoError(t, err)
	require.Len(t, resp, 1)

	tx := resp[0]
	require.NoError(t, tx.ParsingError)
	assert.Equal(t, "9999********8888", tx.SourceAccount)
}

func TestPrivat24_Failure(t *testing.T) {
	type tc struct {
		name    string
		data    []byte
		wantErr string
	}

	cases := []tc{
		{name: "empty data", data: []byte{}, wantErr: "failed to open excel"},
		{name: "bad date", data: privat24BadDate, wantErr: "failed to parse date"},
		{name: "bad amount", data: privat24BadAmount, wantErr: "failed to parse card amount"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{{Data: c.data}})
			require.NoError(t, err)
			require.Len(t, resp, 1)

			require.Error(t, resp[0].ParsingError)
			assert.Contains(t, resp[0].ParsingError.Error(), c.wantErr)
		})
	}
}

func TestPrivat24_ParseDecodeError(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	_, err := srv.Parse(context.Background(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data: []string{"!!!not-base64!!!"},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode")
}
