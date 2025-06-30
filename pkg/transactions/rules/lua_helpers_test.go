package rules_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAccountHelpers(t *testing.T) {
	t.Run("success get account", func(t *testing.T) {
		accSvc := NewMockAccountsSvc(gomock.NewController(t))
		accSvc.EXPECT().GetAccountByID(gomock.Any(), int32(55)).
			Return(&database.Account{
				Currency: "PLN",
			}, nil)

		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{
			AccountsSvc: accSvc,
		})

		script := `
		local account = helpers:getAccountByID(55)
		tx:destinationCurrency(account.Currency)
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)

		assert.EqualValues(t, "PLN", tx.DestinationCurrency)
	})

	t.Run("no arguments", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{
			AccountsSvc: nil,
		})

		script := `
		local account = helpers:getAccountByID()
		tx:destinationCurrency(account.Currency)
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.Error(t, err)
		assert.False(t, result)
		assert.ErrorContains(t, err, "account ID expected")
	})

	t.Run("error getting account", func(t *testing.T) {
		accSvc := NewMockAccountsSvc(gomock.NewController(t))
		accSvc.EXPECT().GetAccountByID(gomock.Any(), int32(55)).
			Return(nil, assert.AnError)

		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{
			AccountsSvc: accSvc,
		})

		script := `
		local account = helpers:getAccountByID(55)
		tx:destinationCurrency(account.Currency)
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.Error(t, err)
		assert.False(t, result)
		assert.ErrorContains(t, err, "failed to get account")
	})
}

func TestConvertHelpers(t *testing.T) {
	t.Run("success convert currency", func(t *testing.T) {
		converterSvc := NewMockCurrencyConverterSvc(gomock.NewController(t))
		decimalSvc := NewMockDecimalSvc(gomock.NewController(t))

		converterSvc.EXPECT().Convert(
			gomock.Any(),
			"USD",
			"EUR",
			decimal.NewFromFloat(100),
		).Return(decimal.NewFromFloat(85.50), nil)

		decimalSvc.EXPECT().GetCurrencyDecimals(gomock.Any(), "EUR").Return(2)

		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{
			CurrencyConverterSvc: converterSvc,
			DecimalSvc:           decimalSvc,
		})

		script := `
		local converted = helpers:convertCurrency("USD", "EUR", 100)
		tx:destinationAmount(converted)
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.EqualValues(t, 85.50, tx.DestinationAmount.Decimal.InexactFloat64())
	})

	t.Run("missing arguments", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{
			CurrencyConverterSvc: nil,
			DecimalSvc:           nil,
		})

		script := `
		local converted = helpers:convertCurrency("USD", "EUR")
		tx:destinationAmount(converted)
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.Error(t, err)
		assert.False(t, result)
		assert.ErrorContains(t, err, "from, to, amount are expected")
	})

	t.Run("conversion error", func(t *testing.T) {
		converterSvc := NewMockCurrencyConverterSvc(gomock.NewController(t))
		decimalSvc := NewMockDecimalSvc(gomock.NewController(t))

		converterSvc.EXPECT().Convert(
			gomock.Any(),
			"USD",
			"EUR",
			decimal.NewFromFloat(100),
		).Return(decimal.Decimal{}, assert.AnError)

		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{
			CurrencyConverterSvc: converterSvc,
			DecimalSvc:           decimalSvc,
		})

		script := `
		local converted = helpers:convertCurrency("USD", "EUR", 100)
		tx:destinationAmount(converted)
	`
		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.False(t, result)
		assert.ErrorContains(t, err, "failed to convert currency")
	})
}
