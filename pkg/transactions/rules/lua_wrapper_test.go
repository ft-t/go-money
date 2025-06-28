package rules_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicFields(t *testing.T) {
	t.Run("title", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:title() == "Old Value" then
			tx:title("New value")
		end
	`

		tx := &database.Transaction{
			Title: "Old Value",
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
	})

	t.Run("source currency", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:sourceCurrency() == "USD" then
			tx:sourceCurrency("EUR")
		end
	`

		tx := &database.Transaction{
			SourceCurrency: "USD",
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, "EUR", tx.SourceCurrency)
	})

	t.Run("destination currency", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:destinationCurrency() == "USD" then
			tx:destinationCurrency("EUR")
		end
	`
		tx := &database.Transaction{
			DestinationCurrency: "USD",
		}
		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)
		assert.True(t, result)
		assert.Equal(t, "EUR", tx.DestinationCurrency)
	})

	t.Run("notes", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:notes() == "Old Notes" then
			tx:notes("New Notes")
		end
	`

		tx := &database.Transaction{
			Notes: "Old Notes",
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, "New Notes", tx.Notes)
	})

	t.Run("reference number", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:referenceNumber() == "12345" then
			tx:referenceNumber("67890")
		end
	`

		tx := &database.Transaction{
			ReferenceNumber: lo.ToPtr("12345"),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, "67890", *tx.ReferenceNumber)
	})

	t.Run("internal reference number", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:internalReferenceNumber() == "12345" then
			tx:internalReferenceNumber("67890")
		end
	`

		tx := &database.Transaction{
			InternalReferenceNumber: lo.ToPtr("12345"),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, "67890", *tx.InternalReferenceNumber)
	})

	t.Run("source account ID", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:sourceAccountID() == 12345 then
			tx:sourceAccountID(67890)
		end
	`

		tx := &database.Transaction{
			SourceAccountID: lo.ToPtr(int32(12345)),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, int32(67890), *tx.SourceAccountID)
	})

	t.Run("source account ID nil", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:sourceAccountID() == nil then
			tx:sourceAccountID(67890)
		end
	`

		tx := &database.Transaction{
			SourceAccountID: nil,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, int32(67890), *tx.SourceAccountID)
	})

	t.Run("set source account ID to nil", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:sourceAccountID() == 12345 then
			tx:sourceAccountID(nil)
		end
	`

		tx := &database.Transaction{
			SourceAccountID: lo.ToPtr(int32(12345)),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Nil(t, tx.SourceAccountID)
	})

	t.Run("destination account ID", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:destinationAccountID() == 12345 then
			tx:destinationAccountID(67890)
		end
	`

		tx := &database.Transaction{
			DestinationAccountID: lo.ToPtr(int32(12345)),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, int32(67890), *tx.DestinationAccountID)
	})

	t.Run("transaction type", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:transactionType() == 1 then
			tx:transactionType(2)
		end
	`

		tx := &database.Transaction{
			TransactionType: 1,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.EqualValues(t, int32(2), tx.TransactionType)
	})

	t.Run("transaction type nil", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter()

		script := `
		if tx:transactionType() == 2 then
			tx:transactionType(nil)
		end
	`

		tx := &database.Transaction{
			TransactionType: 2,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.EqualValues(t, int32(0), tx.TransactionType)
	})
}
