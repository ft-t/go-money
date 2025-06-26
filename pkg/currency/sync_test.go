package currency_test

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"bytes"
	"context"
	_ "embed"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

//go:embed testdata/rates.json
var mockResponse []byte

func TestSync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
		remoteURL := "https://localhost/rates.json"

		mockClient := NewMockhttpClient(gomock.NewController(t))
		mockClient.EXPECT().Do(gomock.Any()).
			DoAndReturn(func(request *http.Request) (*http.Response, error) {
				assert.EqualValues(t, remoteURL, request.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(mockResponse)),
				}, nil
			})

		syn := currency.NewSyncer(mockClient, configuration.CurrencyConfig{
			UpdateTransactionAmountInBaseCurrency: false,
		})

		err := syn.Sync(context.TODO(), remoteURL)
		assert.NoError(t, err)

		var currencies []*database.Currency
		assert.NoError(t, gormDB.Order("id asc").Find(&currencies).Error)

		assert.Len(t, currencies, 3)
		assert.Equal(t, "EUR", currencies[0].ID)
		assert.EqualValues(t, "0.85", currencies[0].Rate.String())
		assert.EqualValues(t, 2, currencies[0].DecimalPlaces)

		assert.Equal(t, "PLN", currencies[1].ID)
		assert.EqualValues(t, "3.8", currencies[1].Rate.String())
		assert.EqualValues(t, 2, currencies[1].DecimalPlaces)

		assert.Equal(t, "USD", currencies[2].ID)
		assert.EqualValues(t, "1", currencies[2].Rate.String())
		assert.EqualValues(t, 2, currencies[2].DecimalPlaces)
	})

	t.Run("success with update base currency", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
		remoteURL := "https://localhost/rates.json"

		mockClient := NewMockhttpClient(gomock.NewController(t))
		mockClient.EXPECT().Do(gomock.Any()).
			DoAndReturn(func(request *http.Request) (*http.Response, error) {
				assert.EqualValues(t, remoteURL, request.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(mockResponse)),
				}, nil
			})

		txs := []*database.Transaction{
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),

				DestinationCurrency: configuration.BaseCurrency,
				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(999)), // should not be updated by script, because foreign currency is set to base
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),

				// source to rate
				// here dest should be null
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  configuration.BaseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),

				// source same,
				// dest null
			},

			// transfers
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  configuration.BaseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),

				// source same,
				// dest null
			},
		}

		syn := currency.NewSyncer(mockClient, configuration.CurrencyConfig{
			UpdateTransactionAmountInBaseCurrency: true,
		})

		err := syn.Sync(context.TODO(), remoteURL)
		assert.NoError(t, err)

		var currencies []*database.Currency
		assert.NoError(t, gormDB.Order("id asc").Find(&currencies).Error)

		assert.Len(t, currencies, 3)
		assert.Equal(t, "EUR", currencies[0].ID)
		assert.EqualValues(t, "0.85", currencies[0].Rate.String())
		assert.EqualValues(t, 2, currencies[0].DecimalPlaces)

		assert.Equal(t, "PLN", currencies[1].ID)
		assert.EqualValues(t, "3.8", currencies[1].Rate.String())
		assert.EqualValues(t, 2, currencies[1].DecimalPlaces)

		assert.Equal(t, "USD", currencies[2].ID)
		assert.EqualValues(t, "1", currencies[2].Rate.String())
		assert.EqualValues(t, 2, currencies[2].DecimalPlaces)

	})

	t.Run("success update", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
		usd := &database.Currency{
			ID:            "USD",
			DecimalPlaces: 99,
			Rate:          decimal.RequireFromString("-1.0"),
		}
		assert.NoError(t, gormDB.Create(usd).Error)

		eur := &database.Currency{
			ID:            "EUR",
			DecimalPlaces: 88,
			Rate:          decimal.RequireFromString("-240"),
		}
		assert.NoError(t, gormDB.Create(eur).Error)

		pln := &database.Currency{
			ID:            "PLN",
			DecimalPlaces: 77,
			Rate:          decimal.RequireFromString("-3.8"),
		}
		assert.NoError(t, gormDB.Create(pln).Error)

		remoteURL := "https://localhost/rates.json"

		mockClient := NewMockhttpClient(gomock.NewController(t))
		mockClient.EXPECT().Do(gomock.Any()).
			DoAndReturn(func(request *http.Request) (*http.Response, error) {
				assert.EqualValues(t, remoteURL, request.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(mockResponse)),
				}, nil
			})

		syn := currency.NewSyncer(mockClient, configuration.CurrencyConfig{})

		err := syn.Sync(context.TODO(), remoteURL)
		assert.NoError(t, err)

		var currencies []*database.Currency
		assert.NoError(t, gormDB.Order("id asc").Find(&currencies).Error)

		assert.Len(t, currencies, 3)

		assert.Equal(t, "EUR", currencies[0].ID)
		assert.EqualValues(t, "0.85", currencies[0].Rate.String())
		assert.EqualValues(t, 88, currencies[0].DecimalPlaces)

		assert.Equal(t, "PLN", currencies[1].ID)
		assert.EqualValues(t, "3.8", currencies[1].Rate.String())
		assert.EqualValues(t, 77, currencies[1].DecimalPlaces)

		assert.Equal(t, "USD", currencies[2].ID)
		assert.EqualValues(t, "1", currencies[2].Rate.String())
		assert.EqualValues(t, 99, currencies[2].DecimalPlaces)
	})

	t.Run("fail request", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
		remoteURL := "https://localhost/rates.json"

		mockClient := NewMockhttpClient(gomock.NewController(t))
		mockClient.EXPECT().Do(gomock.Any()).
			DoAndReturn(func(request *http.Request) (*http.Response, error) {
				assert.EqualValues(t, remoteURL, request.URL.String())

				return nil, assert.AnError
			})

		syn := currency.NewSyncer(mockClient, configuration.CurrencyConfig{})

		err := syn.Sync(context.TODO(), remoteURL)
		assert.Error(t, err)
	})

	t.Run("fail parse json", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
		remoteURL := "https://localhost/rates.json"

		mockClient := NewMockhttpClient(gomock.NewController(t))
		mockClient.EXPECT().Do(gomock.Any()).
			DoAndReturn(func(request *http.Request) (*http.Response, error) {
				assert.EqualValues(t, remoteURL, request.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte("invalid json"))),
				}, nil
			})

		syn := currency.NewSyncer(mockClient, configuration.CurrencyConfig{})

		err := syn.Sync(context.TODO(), remoteURL)
		assert.Error(t, err)
	})
}
