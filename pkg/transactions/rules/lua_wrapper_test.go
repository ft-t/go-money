package rules_test

import (
	"context"
	"testing"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBasicFields(t *testing.T) {
	t.Run("title", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

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
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

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
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

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
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

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
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

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

	t.Run("get internal reference numbers", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		local refs = tx:getInternalReferenceNumbers()
		if refs[1] == "ref1" and refs[2] == "ref2" then
			tx:notes("found_both")
		end
	`

		tx := &database.Transaction{
			InternalReferenceNumbers: []string{"ref1", "ref2"},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, "found_both", tx.Notes)
	})

	t.Run("add internal reference number", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		tx:addInternalReferenceNumber("new_ref")
	`

		tx := &database.Transaction{
			InternalReferenceNumbers: []string{"existing_ref"},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, []string{"existing_ref", "new_ref"}, []string(tx.InternalReferenceNumbers))
	})

	t.Run("set internal reference numbers", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		tx:setInternalReferenceNumbers({"new1", "new2"})
	`

		tx := &database.Transaction{
			InternalReferenceNumbers: []string{"old_ref"},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, []string{"new1", "new2"}, []string(tx.InternalReferenceNumbers))
	})

	t.Run("remove internal reference number", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		tx:removeInternalReferenceNumber("ref2")
	`

		tx := &database.Transaction{
			InternalReferenceNumbers: []string{"ref1", "ref2", "ref3"},
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, []string{"ref1", "ref3"}, []string(tx.InternalReferenceNumbers))
	})

	t.Run("source account ID", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		if tx:sourceAccountID() == 12345 then
			tx:sourceAccountID(67890)
		end
	`

		tx := &database.Transaction{
			SourceAccountID: int32(12345),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, int32(67890), tx.SourceAccountID)
	})

	t.Run("category ID", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		if tx:categoryID() == 12345 then
			tx:categoryID(67890)
		end
	`

		tx := &database.Transaction{
			CategoryID: lo.ToPtr(int32(12345)),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, int32(67890), *tx.CategoryID)
	})

	t.Run("category ID nil", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
			tx:categoryID(nil)

		if tx:categoryID() == nil then
			tx:notes("abc")
		end
	`

		tx := &database.Transaction{
			CategoryID: lo.ToPtr(int32(12345)),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Nil(t, tx.CategoryID)
	})

	t.Run("source account ID nil", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		if tx:sourceAccountID() == 0 then
			tx:sourceAccountID(67890)
		end
	`

		tx := &database.Transaction{
			SourceAccountID: 0,
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, int32(67890), tx.SourceAccountID)
	})

	t.Run("set source account ID to 0", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		if tx:sourceAccountID() == 12345 then
			tx:sourceAccountID(0)
		end
	`

		tx := &database.Transaction{
			SourceAccountID: int32(12345),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.EqualValues(t, 0, tx.SourceAccountID)
	})

	t.Run("destination account ID", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

		script := `
		if tx:destinationAccountID() == 12345 then
			tx:destinationAccountID(67890)
		end
	`

		tx := &database.Transaction{
			DestinationAccountID: int32(12345),
		}

		result, err := interpreter.Run(context.TODO(), script, tx)
		assert.NoError(t, err)

		assert.True(t, result)
		assert.Equal(t, int32(67890), tx.DestinationAccountID)
	})

	t.Run("transaction type", func(t *testing.T) {
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

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
		interpreter := rules.NewLuaInterpreter(&rules.LuaInterpreterConfig{})

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
