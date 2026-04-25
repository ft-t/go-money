package rules_test

import (
	"context"
	"testing"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_NoChange_NoEvent(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	dbRules := []*database.Rule{
		{
			Script:    "noop",
			SortOrder: 1,
		},
	}
	require.NoError(t, gormDB.Create(dbRules).Error)

	interpreter := NewMockInterpreter(gomock.NewController(t))
	srv := rules.NewExecutor(interpreter)

	tx := &database.Transaction{
		ID:    7,
		Title: "stable",
	}

	interpreter.EXPECT().Run(gomock.Any(), "noop", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ *database.Transaction) (bool, error) {
			return false, nil
		})

	out, err := srv.ProcessTransactions(context.Background(), []*database.Transaction{tx})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Empty(t, out[0].RuleAppliedEvents)
}

func TestExecutor_RuleChangedTitle_OneEvent(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	dbRules := []*database.Rule{
		{
			Script:    "rename",
			SortOrder: 1,
		},
	}
	require.NoError(t, gormDB.Create(dbRules).Error)

	interpreter := NewMockInterpreter(gomock.NewController(t))
	srv := rules.NewExecutor(interpreter)

	tx := &database.Transaction{
		ID:    11,
		Title: "old",
	}

	interpreter.EXPECT().Run(gomock.Any(), "rename", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, transaction *database.Transaction) (bool, error) {
			transaction.Title = "new"
			return true, nil
		})

	out, err := srv.ProcessTransactions(context.Background(), []*database.Transaction{tx})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].RuleAppliedEvents, 1)

	ev := out[0].RuleAppliedEvents[0]
	assert.Equal(t, dbRules[0].ID, ev.RuleID)
	require.NotNil(t, ev.Before)
	require.NotNil(t, ev.After)
	assert.Equal(t, "old", ev.Before.Title)
	assert.Equal(t, "new", ev.After.Title)
}

func TestExecutor_TwoRulesEachChange_TwoEvents(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	dbRules := []*database.Rule{
		{
			Script:    "set_title",
			SortOrder: 1,
			GroupName: "a",
		},
		{
			Script:    "set_notes",
			SortOrder: 1,
			GroupName: "b",
		},
	}
	require.NoError(t, gormDB.Create(dbRules).Error)

	interpreter := NewMockInterpreter(gomock.NewController(t))
	srv := rules.NewExecutor(interpreter)

	tx := &database.Transaction{
		ID:    21,
		Title: "old",
		Notes: "no notes",
	}

	interpreter.EXPECT().Run(gomock.Any(), "set_title", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, transaction *database.Transaction) (bool, error) {
			transaction.Title = "renamed"
			return true, nil
		})

	interpreter.EXPECT().Run(gomock.Any(), "set_notes", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, transaction *database.Transaction) (bool, error) {
			transaction.Notes = "with notes"
			return true, nil
		})

	out, err := srv.ProcessTransactions(context.Background(), []*database.Transaction{tx})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].RuleAppliedEvents, 2)

	first := out[0].RuleAppliedEvents[0]
	assert.Equal(t, dbRules[0].ID, first.RuleID)
	assert.Equal(t, "old", first.Before.Title)
	assert.Equal(t, "renamed", first.After.Title)

	second := out[0].RuleAppliedEvents[1]
	assert.Equal(t, dbRules[1].ID, second.RuleID)
	assert.Equal(t, "no notes", second.Before.Notes)
	assert.Equal(t, "with notes", second.After.Notes)
	assert.Equal(t, "renamed", second.Before.Title)
	assert.Equal(t, "renamed", second.After.Title)
}

func TestExecutor_ChangeNoChangeChange_KeepsBothEvents(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	dbRules := []*database.Rule{
		{Script: "rename", SortOrder: 1},
		{Script: "touch", SortOrder: 2},
		{Script: "notes", SortOrder: 3},
	}
	require.NoError(t, gormDB.Create(dbRules).Error)

	interpreter := NewMockInterpreter(gomock.NewController(t))
	srv := rules.NewExecutor(interpreter)

	tx := &database.Transaction{ID: 55, Title: "old", Notes: "no notes"}

	interpreter.EXPECT().Run(gomock.Any(), "rename", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, transaction *database.Transaction) (bool, error) {
			transaction.Title = "renamed"
			return true, nil
		})
	interpreter.EXPECT().Run(gomock.Any(), "touch", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ *database.Transaction) (bool, error) {
			return true, nil
		})
	interpreter.EXPECT().Run(gomock.Any(), "notes", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, transaction *database.Transaction) (bool, error) {
			transaction.Notes = "with notes"
			return true, nil
		})

	out, err := srv.ProcessTransactions(context.Background(), []*database.Transaction{tx})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Len(t, out[0].RuleAppliedEvents, 2)
	assert.Equal(t, dbRules[0].ID, out[0].RuleAppliedEvents[0].RuleID)
	assert.Equal(t, dbRules[2].ID, out[0].RuleAppliedEvents[1].RuleID)
}

func TestExecutor_RuleReturnedTrueButNoChange_NoEvent(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	dbRules := []*database.Rule{
		{
			Script:    "touch",
			SortOrder: 1,
		},
	}
	require.NoError(t, gormDB.Create(dbRules).Error)

	interpreter := NewMockInterpreter(gomock.NewController(t))
	srv := rules.NewExecutor(interpreter)

	tx := &database.Transaction{
		ID:    33,
		Title: "stable",
	}

	interpreter.EXPECT().Run(gomock.Any(), "touch", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ *database.Transaction) (bool, error) {
			return true, nil
		})

	out, err := srv.ProcessTransactions(context.Background(), []*database.Transaction{tx})
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Empty(t, out[0].RuleAppliedEvents)
}
