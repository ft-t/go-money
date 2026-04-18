package currency_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvert(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	usd := &database.Currency{
		ID:            "USD",
		DecimalPlaces: 2,
		Rate:          decimal.RequireFromString("1.0"),
	}
	assert.NoError(t, gormDB.Create(usd).Error)

	pln := &database.Currency{
		ID:            "PLN",
		DecimalPlaces: 2,
		Rate:          decimal.RequireFromString("3.8"),
	}
	assert.NoError(t, gormDB.Create(pln).Error)

	eur := &database.Currency{
		ID:            "EUR",
		DecimalPlaces: 2,
		Rate:          decimal.RequireFromString("0.8"),
	}
	assert.NoError(t, gormDB.Create(eur).Error)

	cv := currency.NewConverter("USD")

	t.Run("USD -> PLN", func(t *testing.T) {
		resp, err := cv.Convert(
			context.TODO(),
			"USD",
			"PLN",
			decimal.RequireFromString("3.0"),
		)

		assert.NoError(t, err)
		assert.EqualValues(t, "11.40", resp.StringFixed(2))
	})

	t.Run("PLN -> USD -> EUR", func(t *testing.T) {
		resp, err := cv.Convert(
			context.TODO(),
			"PLN",
			"EUR",
			decimal.RequireFromString("4.0"),
		)

		assert.NoError(t, err)
		assert.EqualValues(t, "0.84", resp.StringFixed(2))
	})

	t.Run("USD -> XYZ", func(t *testing.T) {
		resp, err := cv.Convert(
			context.TODO(),
			"USD",
			"XYZ",
			decimal.RequireFromString("3.0"),
		)

		assert.ErrorContains(t, err, "rate for XYZ not found")
		assert.EqualValues(t, decimal.Zero, resp)
	})
}

func TestMissingBase(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	cv := currency.NewConverter("USD")

	resp, err := cv.Convert(
		context.TODO(),
		"USD",
		"PLN",
		decimal.RequireFromString("3.0"),
	)

	assert.ErrorContains(t, err, "rate for USD not found")
	assert.EqualValues(t, decimal.Zero, resp)
}

func TestConverter_Quote_Success(t *testing.T) {
	type caseDef struct {
		name          string
		from          string
		to            string
		amount        decimal.Decimal
		wantConverted string
		wantFromRate  string
		wantToRate    string
	}

	cases := []caseDef{
		{
			name:          "standard conversion USD to EUR",
			from:          "USD",
			to:            "EUR",
			amount:        decimal.NewFromInt(100),
			wantConverted: "90",
			wantFromRate:  "1",
			wantToRate:    "0.9",
		},
		{
			name:          "same currency USD to USD",
			from:          "USD",
			to:            "USD",
			amount:        decimal.NewFromInt(50),
			wantConverted: "50",
			wantFromRate:  "1",
			wantToRate:    "1",
		},
		{
			name:          "cross conversion EUR to PLN",
			from:          "EUR",
			to:            "PLN",
			amount:        decimal.NewFromInt(100),
			wantConverted: "444.4444444444444444",
			wantFromRate:  "0.9",
			wantToRate:    "4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

			usd := &database.Currency{
				ID:            "USD",
				DecimalPlaces: 2,
				Rate:          decimal.RequireFromString("1"),
			}
			assert.NoError(t, gormDB.Create(usd).Error)

			eur := &database.Currency{
				ID:            "EUR",
				DecimalPlaces: 2,
				Rate:          decimal.RequireFromString("0.9"),
			}
			assert.NoError(t, gormDB.Create(eur).Error)

			pln := &database.Currency{
				ID:            "PLN",
				DecimalPlaces: 2,
				Rate:          decimal.RequireFromString("4"),
			}
			assert.NoError(t, gormDB.Create(pln).Error)

			cv := currency.NewConverter("USD")

			quote, err := cv.Quote(context.TODO(), tc.from, tc.to, tc.amount)
			assert.NoError(t, err)
			assert.NotNil(t, quote)
			assert.Equal(t, tc.from, quote.From)
			assert.Equal(t, tc.to, quote.To)
			assert.EqualValues(t, tc.amount.String(), quote.Amount.String())
			assert.Equal(t, "USD", quote.BaseCurrency)
			assert.EqualValues(t, tc.wantFromRate, quote.FromRate.String())
			assert.EqualValues(t, tc.wantToRate, quote.ToRate.String())
			assert.EqualValues(t, tc.wantConverted, quote.Converted.String())
		})
	}
}

func TestConverter_Quote_Failure(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	usd := &database.Currency{
		ID:            "USD",
		DecimalPlaces: 2,
		Rate:          decimal.RequireFromString("1"),
	}
	assert.NoError(t, gormDB.Create(usd).Error)

	cv := currency.NewConverter("USD")

	quote, err := cv.Quote(context.TODO(), "USD", "EUR", decimal.NewFromInt(100))
	assert.ErrorContains(t, err, "rate for EUR not found")
	assert.Nil(t, quote)
}
