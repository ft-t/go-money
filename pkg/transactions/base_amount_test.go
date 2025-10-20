package transactions_test

import (
	"context"
	"testing"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/ft-t/go-money/pkg/transactions/validation"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBaseAmountService(t *testing.T) {
	baseCurrency := "USD"
	t.Run("success with full update", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		rates := []*database.Currency{
			{
				ID:   "PLN",
				Rate: decimal.NewFromFloat(3.8),
			},
			{
				ID:   baseCurrency,
				Rate: decimal.NewFromFloat(1.0),
			},
			{
				ID:   "EUR",
				Rate: decimal.NewFromFloat(0.85),
			},
		}
		assert.NoError(t, gormDB.Create(&rates).Error)

		acc := []*database.Account{
			{
				Name:     "PLN",
				Currency: "PLN",
				Extra:    make(map[string]string),
				Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				Name:     baseCurrency,
				Currency: baseCurrency,
				Extra:    make(map[string]string),
				Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				Name:     "UAH",
				Currency: "UAH",
				Extra:    make(map[string]string),
				Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				Name:     "EUR",
				Currency: "EUR",
				Extra:    make(map[string]string),
				Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				Name:     "Expense Account",
				Extra:    map[string]string{},
				Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
				Currency: baseCurrency,
			},
			{
				Name:     "Expense Account in RAND currency",
				Extra:    map[string]string{},
				Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
				Currency: "RAND",
			},
			{
				Name:     "Default Income",
				Extra:    map[string]string{},
				Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME,
				Currency: baseCurrency,
			},
		}
		assert.NoError(t, gormDB.Create(&acc).Error)

		txs := []*database.Transaction{
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceAccountID: acc[0].ID,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-10)),

				FxSourceCurrency: baseCurrency,
				FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-999)),
				Extra:            make(map[string]string),

				DestinationCurrency:  acc[5].Currency,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(2)),
				DestinationAccountID: acc[5].ID,

				// should not be updated by script, because foreign currency is set to base
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-10)),
				Extra:           make(map[string]string),
				SourceAccountID: acc[0].ID,

				DestinationCurrency:  acc[5].Currency,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(1)),
				DestinationAccountID: acc[5].ID,
				// source to rate
				// sourceInBase = destInBase
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  baseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
				SourceAccountID: acc[1].ID,
				Extra:           make(map[string]string),

				DestinationCurrency:  acc[5].Currency,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(11)),
				DestinationAccountID: acc[5].ID,

				// [2]
				// source same,
				// dest null
			},

			// transfers
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  baseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
				SourceAccountID: acc[1].ID,

				DestinationAccountID: acc[0].ID,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(9999)),
				DestinationCurrency:  "PLN",
				Extra:                make(map[string]string),

				// source is in USD to should use 55,
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-9999)),
				SourceAccountID: acc[0].ID,

				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
				DestinationCurrency:  baseCurrency,
				DestinationAccountID: acc[1].ID,
				Extra:                make(map[string]string),

				// destination is in USD to should use 55,
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-100)),
				SourceAccountID: acc[0].ID,

				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(999)),
				DestinationCurrency:  "UAH",
				DestinationAccountID: acc[2].ID,

				Extra: make(map[string]string),
				// should convert to base currency with same amount from PLN
			},

			// withdrawal fx

			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  acc[0].Currency, // pln
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
				SourceAccountID: acc[0].ID,
				Extra:           make(map[string]string),

				FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-10)),
				FxSourceCurrency: baseCurrency,

				DestinationCurrency:  acc[5].Currency,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(0)),
				DestinationAccountID: acc[5].ID,

				// [6]
				// source different,
				// fx base
				// dest null
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  acc[0].Currency, // pln
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
				SourceAccountID: acc[0].ID,
				Extra:           make(map[string]string),

				FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-10)),
				FxSourceCurrency: acc[2].Currency, // UAH

				DestinationCurrency:  acc[1].Currency,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(30)),
				DestinationAccountID: acc[1].ID,

				// source different,
				// fx different
				// dest base
			},

			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  acc[0].Currency, // pln
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
				SourceAccountID: acc[0].ID,
				Extra:           make(map[string]string),

				FxSourceAmount:   decimal.NewNullDecimal(decimal.NewFromInt(-10)),
				FxSourceCurrency: acc[2].Currency, // UAH

				DestinationCurrency:  acc[3].Currency,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(30)),
				DestinationAccountID: acc[3].ID,

				// source different,
				// fx different
				// dest different
			},

			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT,
				SourceCurrency:  acc[1].Currency, // usd
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(-55)),
				SourceAccountID: acc[1].ID,
				Extra:           make(map[string]string),

				DestinationCurrency:  acc[3].Currency,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(30)),
				DestinationAccountID: acc[3].ID,

				// [9]
				// source base
				// dest different
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
				SourceCurrency:  baseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(1000)),
				SourceAccountID: acc[6].ID,

				DestinationCurrency:  baseCurrency,
				DestinationAccountID: acc[1].ID,
				DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(1000)),
				Extra:                make(map[string]string),
			},
		}

		txSrv := validation.NewValidationService(&validation.ServiceConfig{})
		for _, tx := range txs {
			assert.NoError(t, txSrv.ValidateTransactionData(context.TODO(), gormDB, tx))
		}

		assert.NoError(t, gormDB.Create(&txs).Error)

		svc := transactions.NewBaseAmountService("USD")

		err := svc.RecalculateAmountInBaseCurrencyForAll(context.TODO(), gormDB)
		assert.NoError(t, err)

		var updatedTxs []*database.Transaction
		assert.NoError(t, gormDB.Order("id asc").Find(&updatedTxs).Error)

		assert.EqualValues(t, 999, updatedTxs[0].DestinationAmountInBaseCurrency.Decimal.IntPart()) // because FX is set to base
		assert.EqualValues(t, -999, updatedTxs[0].SourceAmountInBaseCurrency.Decimal.IntPart())     // because FX is set to base

		assert.EqualValues(t, -3, updatedTxs[1].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 3, updatedTxs[1].DestinationAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, -55, updatedTxs[2].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[2].DestinationAmountInBaseCurrency.Decimal.IntPart()) // because source is in base currency

		// transfers
		assert.EqualValues(t, 55, updatedTxs[3].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, -55, updatedTxs[3].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, 55, updatedTxs[4].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, -55, updatedTxs[4].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, 26, updatedTxs[5].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, -26, updatedTxs[5].SourceAmountInBaseCurrency.Decimal.IntPart())

		// withdrawal fx
		assert.EqualValues(t, -10, updatedTxs[6].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 10, updatedTxs[6].DestinationAmountInBaseCurrency.Decimal.IntPart()) // fx is base

		assert.EqualValues(t, -30, updatedTxs[7].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 30, updatedTxs[7].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, true, updatedTxs[7].DestinationAmountInBaseCurrency.Valid)

		assert.EqualValues(t, -14, updatedTxs[8].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 14, updatedTxs[8].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, true, updatedTxs[8].DestinationAmountInBaseCurrency.Valid)

		// adjustment

		assert.EqualValues(t, -55, updatedTxs[9].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[9].DestinationAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, 1000, updatedTxs[10].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, -1000, updatedTxs[10].SourceAmountInBaseCurrency.Decimal.IntPart())
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
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),

				DestinationCurrency: baseCurrency,
				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(999)),
				Extra:               make(map[string]string),

				// should not be updated by script, because foreign currency is set to base
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),
				Extra:           make(map[string]string),

				// source to rate
				// here dest should be null
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  baseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
				Extra:           make(map[string]string),

				// source same,
				// dest null
			},

			// transfers
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  baseCurrency,
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
				DestinationCurrency: baseCurrency,
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

		svc := transactions.NewBaseAmountService("USD")

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
		assert.EqualValues(t, -999, updatedTxs[0].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, false, updatedTxs[1].SourceAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[1].DestinationAmountInBaseCurrency.Valid)

		assert.EqualValues(t, -55, updatedTxs[2].SourceAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, 55, updatedTxs[2].DestinationAmountInBaseCurrency.Decimal.IntPart())

		// transfers
		assert.EqualValues(t, 55, updatedTxs[3].DestinationAmountInBaseCurrency.Decimal.IntPart())
		assert.EqualValues(t, -55, updatedTxs[3].SourceAmountInBaseCurrency.Decimal.IntPart())

		assert.EqualValues(t, false, updatedTxs[4].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[4].SourceAmountInBaseCurrency.Valid)

		assert.EqualValues(t, false, updatedTxs[5].DestinationAmountInBaseCurrency.Valid)
		assert.EqualValues(t, false, updatedTxs[5].SourceAmountInBaseCurrency.Valid)
	})

	t.Run("db error", func(t *testing.T) {
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
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),

				DestinationCurrency: baseCurrency,
				DestinationAmount:   decimal.NewNullDecimal(decimal.NewFromInt(999)),
				Extra:               make(map[string]string),

				// should not be updated by script, because foreign currency is set to base
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  "PLN",
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(10)),
				Extra:           make(map[string]string),

				// source to rate
				// here dest should be null
			},
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				SourceCurrency:  baseCurrency,
				SourceAmount:    decimal.NewNullDecimal(decimal.NewFromInt(55)),
				Extra:           make(map[string]string),

				// source same,
				// dest null
			},

			// transfers
			{
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
				SourceCurrency:  baseCurrency,
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
				DestinationCurrency: baseCurrency,
				Extra:               make(map[string]string),

				// destination is in USD to should use 55,
			},
		}

		assert.NoError(t, gormDB.Create(&txs).Error)

		svc := transactions.NewBaseAmountService("USD")

		gormMock, _, sql := testingutils.GormMock()

		sql.ExpectQuery("with upd as .*").WithArgs(
			"USD",
			"USD",
			"USD",
			"USD",
			pq.Int64Array{txs[0].ID, txs[2].ID, txs[3].ID},
			pq.Int64Array{txs[0].ID, txs[2].ID, txs[3].ID},
		).WillReturnError(errors.New("db error"))
		err := svc.RecalculateAmountInBaseCurrency(context.TODO(), gormMock,
			[]*database.Transaction{
				txs[0],
				txs[2],
				txs[3],
			},
		)
		assert.ErrorContains(t, err, "db error")
	})
}
