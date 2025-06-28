package rules_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecuteRule(t *testing.T) {
	interpreter := NewMockInterpreter(gomock.NewController(t))

	srv := rules.NewService(interpreter)

	tx := &database.Transaction{
		ID:     22,
		Title:  "abcd",
		TagIDs: []int32{1, 2},
	}

	var clonedTx1 *database.Transaction
	var clonedTx2 *database.Transaction

	interpreter.EXPECT().Run(gomock.Any(), "somescript", gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ string, transaction *database.Transaction) (bool, error) {
			clonedTx1 = transaction
			assert.NotSame(t, transaction, tx) // ensure the transaction is cloned

			transaction.Title = "modified title"
			return true, nil
		})

	interpreter.EXPECT().Run(gomock.Any(), "somescript2", gomock.Any()).
		DoAndReturn(func(ctx context.Context, s string, transaction *database.Transaction) (bool, error) {
			clonedTx2 = transaction

			assert.NotSame(t, transaction, tx)        // ensure the transaction is cloned
			assert.NotSame(t, clonedTx1, transaction) // ensure tx ptr not same as for first script

			assert.EqualValues(t, "modified title", clonedTx1.Title) // ensure first script modified tx

			transaction.Notes = "new notes"

			return true, nil
		})

	newTx, err := srv.ProcessTransactions(context.TODO(), []*database.Transaction{tx})
	assert.NoError(t, err)
	assert.Len(t, newTx, 1)

	assert.Same(t, newTx[0], clonedTx2) // ptrs should be the same
	assert.EqualValues(t, "modified title", newTx[0].Title)
	assert.EqualValues(t, "new notes", newTx[0].Notes)
}
