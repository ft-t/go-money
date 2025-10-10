package importers_test

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/ft-t/go-money/pkg/importers"
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
