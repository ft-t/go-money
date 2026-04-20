package history_test

import (
	"testing"
	"time"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/history"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshot_Success_DropsExcludedKeys(t *testing.T) {
	tx := &database.Transaction{
		ID:                  42,
		Title:               "lunch",
		SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(1)),
		SourceCurrency:      "USD",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		TagIDs:              pq.Int32Array{1, 2},
		TransactionDateTime: time.Now(),
		TransactionDateOnly: time.Now(),
	}
	snap, err := history.Snapshot(tx)
	require.NoError(t, err)

	for _, k := range []string{"title", "source_amount", "source_currency", "tag_ids"} {
		_, ok := snap[k]
		assert.True(t, ok, "expected key %q in snapshot", k)
	}
	for _, k := range []string{"id", "created_at", "updated_at", "deleted_at",
		"source_amount_in_base_currency", "destination_amount_in_base_currency"} {
		_, ok := snap[k]
		assert.False(t, ok, "expected key %q NOT in snapshot", k)
	}

	assert.Equal(t, "1", snap["source_amount"])
	assert.Equal(t, "USD", snap["source_currency"])
}

func TestSnapshot_Success_NullableShape(t *testing.T) {
	tx := &database.Transaction{
		ID:    7,
		Title: "no fx, no tags",
	}
	snap, err := history.Snapshot(tx)
	require.NoError(t, err)

	assert.Nil(t, snap["fx_source_amount"])
	assert.Nil(t, snap["destination_amount"])
	assert.Nil(t, snap["category_id"])
	assert.Nil(t, snap["tag_ids"])
	assert.Nil(t, snap["reference_number"])
	assert.Equal(t, "", snap["fx_source_currency"])
}

func TestDiff_Success(t *testing.T) {
	a := map[string]any{"title": "old", "notes": "same"}
	b := map[string]any{"title": "new", "notes": "same"}
	diff, err := history.Diff(a, b)
	require.NoError(t, err)
	require.NotNil(t, diff)

	ops, ok := diff["ops"].([]any)
	require.True(t, ok)
	require.Len(t, ops, 1)

	op := ops[0].(map[string]any)
	assert.Equal(t, "replace", op["op"])
	assert.Equal(t, "/title", op["path"])
	assert.Equal(t, "new", op["value"])
}

func TestDiff_Empty(t *testing.T) {
	a := map[string]any{"title": "same"}
	diff, err := history.Diff(a, a)
	require.NoError(t, err)
	assert.Nil(t, diff)
}
