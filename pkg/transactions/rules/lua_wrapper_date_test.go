package rules_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAddDate(t *testing.T) {
	t.Run("source amount without lua convert", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `tx:transactionDateTimeAddDate(2,3,4)`

		dt := time.Now().UTC()
		tx := &database.Transaction{
			DestinationAmount:   decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
			TransactionDateTime: dt,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, dt.AddDate(2, 3, 4), tx.TransactionDateTime)
	})

	t.Run("missing argument", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `tx:transactionDateTimeAddDate(2,3)`

		dt := time.Now().UTC()
		tx := &database.Transaction{
			DestinationAmount:   decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
			TransactionDateTime: dt,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.ErrorContains(t, err, "expected 3 arguments")
		assert.False(t, result)
		assert.Equal(t, dt, tx.TransactionDateTime)
	})
}

func TestSetTime(t *testing.T) {
	t.Run("set time", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `tx:transactionDateTimeSetTime(12, 30)`

		dt := time.Now().UTC()
		tx := &database.Transaction{
			DestinationAmount:   decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
			TransactionDateTime: dt,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		expectedTime := dt.Truncate(time.Hour * 24).Add(time.Hour*12 + time.Minute*30)
		assert.Equal(t, expectedTime, tx.TransactionDateTime)
	})

	t.Run("set time to 0", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `tx:transactionDateTimeSetTime(0, 0)`

		dt := time.Now().UTC()
		tx := &database.Transaction{
			DestinationAmount:   decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
			TransactionDateTime: dt,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		expectedTime := dt.Truncate(time.Hour * 24)
		assert.Equal(t, expectedTime, tx.TransactionDateTime)
	})

	t.Run("missing argument", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `tx:transactionDateTimeSetTime(12)`

		dt := time.Now().UTC()
		tx := &database.Transaction{
			DestinationAmount:   decimal.NullDecimal{Decimal: decimal.NewFromFloat(100.50), Valid: true},
			TransactionDateTime: dt,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.ErrorContains(t, err, "expected 2 arguments")
		assert.False(t, result)
		assert.Equal(t, dt, tx.TransactionDateTime)
	})
}
