package transactions_test

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBaseAmountService(t *testing.T) {
	t.Run("success with full update", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		rates := []*database.Currency{
			{
				ID:   "PLN",
				Rate: decimal.NewFromFloat(3.8),
			},
			{
				ID:   "USD",
				Rate: decimal.NewFromFloat(1.0),
			},
			{
				ID:   "EUR",
				Rate: decimal.NewFromFloat(0.85),
			},
		}
		assert.NoError(t, gormDB.Create(&rates).Error)

		txs := []*database.Transaction{
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),

				DestinationCurrency: configuration.BaseCurrency,
				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(999)),
				Extra:               make(map[string]string),

				// should not be updated by script, because foreign currency is set to base
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),
				Extra:           make(map[string]string),

				// source to rate
				// here dest should be null
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  configuration.BaseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
				Extra:           make(map[string]string),

				// source same,
				// dest null
			},

			// transfers
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  configuration.BaseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),

				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(9999)),
				DestinationCurrency: "PLN",
				Extra:               make(map[string]string),

				// source is in USD to should use 55,
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(9999)),

				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(55)),
				DestinationCurrency: configuration.BaseCurrency,
				Extra:               make(map[string]string),

				// destination is in USD to should use 55,
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),

				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(999)),
				DestinationCurrency: "UAH",

				Extra: make(map[string]string),
				// should convert to base currency with same amount from PLN
			},
		}

		assert.NoError(t, gormDB.Create(&txs).Error)

		svc := transactions.NewBaseAmountService()

		err := svc.RecalculateAmountInBaseCurrencyForAll(context.TODO(), gormDB)
		assert.NoError(t, err)

		var updatedTxs []*database.Transaction
		assert.NoError(t, gormDB.Order("id asc").Find(&updatedTxs).Error)

		assert.EqualValues(t, 999, updatedTxs[0].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 999, updatedTxs[0].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, 3, updatedTxs[1].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, false, updatedTxs[1].DestinationAmountInBaseCurrency.Valid)

		assert.EqualValues(t, 55, updatedTxs[2].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, false, updatedTxs[2].DestinationAmountInBaseCurrency.Valid)

		// transfers
		assert.EqualValues(t, 55, updatedTxs[3].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[3].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, 55, updatedTxs[4].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[4].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, 26, updatedTxs[5].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 26, updatedTxs[5].SourceAmountInBaseCurrency.Decimal.IntPart())
	})

	t.Run("success with partial update", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		rates := []*database.Currency{
			{
				ID:   "PLN",
				Rate: decimal.NewFromFloat(3.8),
			},
			{
				ID:   "USD",
				Rate: decimal.NewFromFloat(1.0),
			},
			{
				ID:   "EUR",
				Rate: decimal.NewFromFloat(0.85),
			},
		}
		assert.NoError(t, gormDB.Create(&rates).Error)

		txs := []*database.Transaction{
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),

				DestinationCurrency: configuration.BaseCurrency,
				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(999)),
				Extra:               make(map[string]string),

				// should not be updated by script, because foreign currency is set to base
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),
				Extra:           make(map[string]string),

				// source to rate
				// here dest should be null
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
				SourceCurrency:  configuration.BaseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
				Extra:           make(map[string]string),

				// source same,
				// dest null
			},

			// transfers
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  configuration.BaseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),

				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(9999)),
				DestinationCurrency: "PLN",
				Extra:               make(map[string]string),

				// source is in USD to should use 55,
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(9999)),

				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(55)),
				DestinationCurrency: configuration.BaseCurrency,
				Extra:               make(map[string]string),

				// destination is in USD to should use 55,
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(100)),

				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(999)),
				DestinationCurrency: "UAH",

				Extra: make(map[string]string),
				// should convert to base currency with same amount from PLN
			},
		}

		assert.NoError(t, gormDB.Create(&txs).Error)

		svc := transactions.NewBaseAmountService()

		err := svc.RecalculateAmountInBaseCurrency(context.TODO(), gormDB,
			[]*database.Transaction{
				txs[0],
				txs[2],
				txs[3],
			},
		)
		assert.NoError(t, err)

		var updatedTxs []*database.Transaction
		assert.NoError(t, gormDB.Order("id asc").Find(&updatedTxs).Error)

		assert.EqualValues(t, 999, updatedTxs[0].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 999, updatedTxs[0].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 999, updatedTxs[0].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 999, updatedTxs[0].DestinationAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, false, updatedTxs[1].SourceAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[1].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[1].SourceAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[1].DestinationAmountInBaseCurrency.Valid)

		assert.EqualValues(t, 55, updatedTxs[2].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, false, updatedTxs[2].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, 55, updatedTxs[2].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, false, updatedTxs[2].DestinationAmountInBaseCurrency.Valid)

		// transfers
		assert.EqualValues(t, 55, updatedTxs[3].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[3].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[3].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[3].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, false, updatedTxs[4].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[4].SourceAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[4].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[4].SourceAmountInBaseCurrency.Valid)

		assert.EqualValues(t, false, updatedTxs[5].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[5].SourceAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[5].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[5].SourceAmountInBaseCurrency.Valid)
	})
}
