package rules_test

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestExecuteRule(t *testing.T) {
	t.Run("two rules", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		dbRules := []*database.Rule{
			{
				Script:    "somescript",
				SortOrder: 1,
			},
			{
				Script:    "somescript2",
				SortOrder: 2,
			},
		}
		assert.NoError(t, gormDB.Create(dbRules).Error)

		interpreter := NewMockInterpreter(gomock.NewController(t))

		srv := rules.NewExecutor(interpreter)

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
	})

	t.Run("no changes on rule error", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		dbRules := []*database.Rule{
			{
				Script:    "somescript",
				SortOrder: 1,
			},
			{
				Script:    "somescript2",
				SortOrder: 1,
			},
		}
		assert.NoError(t, gormDB.Create(dbRules).Error)

		interpreter := NewMockInterpreter(gomock.NewController(t))

		srv := rules.NewExecutor(interpreter)

		tx := &database.Transaction{
			ID:     22,
			Title:  "abcd",
			Notes:  "original notes",
			TagIDs: []int32{1, 2},
		}

		var clonedTx1 *database.Transaction

		interpreter.EXPECT().Run(gomock.Any(), "somescript", gomock.Any()).
			DoAndReturn(func(ctx context.Context, _ string, transaction *database.Transaction) (bool, error) {
				clonedTx1 = transaction
				assert.NotSame(t, transaction, tx) // ensure the transaction is cloned

				transaction.Title = "modified title"
				return true, nil
			})

		interpreter.EXPECT().Run(gomock.Any(), "somescript2", gomock.Any()).
			DoAndReturn(func(ctx context.Context, s string, transaction *database.Transaction) (bool, error) {
				assert.NotSame(t, transaction, tx)        // ensure the transaction is cloned
				assert.NotSame(t, clonedTx1, transaction) // ensure tx ptr not same as for first script

				assert.EqualValues(t, "modified title", clonedTx1.Title) // ensure first script modified tx

				transaction.Notes = "new notes"

				return false, errors.New("some error")
			})

		newTx, err := srv.ProcessTransactions(context.TODO(), []*database.Transaction{tx})
		assert.ErrorContains(t, err, "some error")
		assert.Len(t, newTx, 0)

		assert.EqualValues(t, "abcd", tx.Title)
		assert.EqualValues(t, "original notes", tx.Notes)
	})

	t.Run("second rule break execution", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		dbRules := []*database.Rule{
			{
				Script:    "somescript",
				SortOrder: 1,
			},
			{
				Script:      "somescript2",
				SortOrder:   2,
				IsFinalRule: true,
			},
			{
				Script:    "somescript3",
				SortOrder: 3,
			},
		}
		assert.NoError(t, gormDB.Create(dbRules).Error)

		interpreter := NewMockInterpreter(gomock.NewController(t))

		srv := rules.NewExecutor(interpreter)

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
	})

	t.Run("multiple groups will execute", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		dbRules := []*database.Rule{
			{
				Script:    "somescript",
				SortOrder: 1,
			},
			{
				Script:      "somescript2",
				SortOrder:   2,
				IsFinalRule: true,
			},
			{
				Script:    "somescript3",
				SortOrder: 3,
			},
			{
				Script:    "somescript4",
				SortOrder: 5,
				GroupName: "Some Group",
			},
		}
		assert.NoError(t, gormDB.Create(dbRules).Error)

		interpreter := NewMockInterpreter(gomock.NewController(t))

		srv := rules.NewExecutor(interpreter)

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

		interpreter.EXPECT().Run(gomock.Any(), "somescript4", gomock.Any()).
			DoAndReturn(func(ctx context.Context, s string, transaction *database.Transaction) (bool, error) {
				clonedTx2 = transaction

				transaction.DestinationCurrency = "PLN"

				return true, nil
			})

		newTx, err := srv.ProcessTransactions(context.TODO(), []*database.Transaction{tx})
		assert.NoError(t, err)
		assert.Len(t, newTx, 1)

		assert.Same(t, newTx[0], clonedTx2) // ptrs should be the same

		assert.EqualValues(t, "modified title", newTx[0].Title)
		assert.EqualValues(t, "new notes", newTx[0].Notes)
		assert.EqualValues(t, "PLN", newTx[0].DestinationCurrency)
	})
}
