package rules_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLuaAmounts(t *testing.T) {
	t.Run("source amount with lua convert", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		function decimal(x)
			return string.format("%.3f",x)
		end

		if decimal(tx:sourceAmount()) == decimal(100.5) then
			tx:sourceAmount(200.75)
		end
	`

		tx := &database.Transaction{
			SourceAmount: decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, decimal.NewFromFloat(200.75), tx.SourceAmount.Decimal)
		assert.True(t, tx.SourceAmount.Valid)
	})

	t.Run("source amount without lua convert", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:destinationAmount() == 100.5 then
			tx:destinationAmount(200.75)
		end
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, decimal.NewFromFloat(200.75), tx.DestinationAmount.Decimal)
		assert.True(t, tx.DestinationAmount.Valid)
	})

	t.Run("source amount with decimal places", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:getSourceAmountWithDecimalPlaces(3) == 100.500 then
			tx:sourceAmount(200.75)
		end
	`

		tx := &database.Transaction{
			SourceAmount: decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, decimal.NewFromFloat(200.75), tx.SourceAmount.Decimal)
		assert.True(t, tx.SourceAmount.Valid)
	})

	t.Run("destination amount with decimal places", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:getDestinationAmountWithDecimalPlaces(3) == 100.500 then
			tx:destinationAmount(200.75)
		end
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, decimal.NewFromFloat(200.75), tx.DestinationAmount.Decimal)
		assert.True(t, tx.DestinationAmount.Valid)
	})

	t.Run("decimal places with nil amount", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:getSourceAmountWithDecimalPlaces(3) == nil then
			tx:sourceAmount(200.75)
		end
	`

		tx := &database.Transaction{
			SourceAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, decimal.NewFromFloat(200.75), tx.SourceAmount.Decimal)
		assert.True(t, tx.SourceAmount.Valid)
	})

	t.Run("decimal places are missing", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:getDestinationAmountWithDecimalPlaces() == nil then
			tx:destinationAmount(200.75)
		end
	`

		tx := &database.Transaction{
			DestinationAmount: decimal.NullDecimal{Valid: false},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.ErrorContains(t, err, "decimalPlaces expected")

		assert.False(t, result)
	})
}
