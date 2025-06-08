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

	cv := currency.NewConverter()

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

	cv := currency.NewConverter()

	resp, err := cv.Convert(
		context.TODO(),
		"USD",
		"PLN",
		decimal.RequireFromString("3.0"),
	)

	assert.ErrorContains(t, err, "rate for USD not found")
	assert.EqualValues(t, decimal.Zero, resp)
}
