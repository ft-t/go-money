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

type snapshotCase struct {
	name      string
	tx        *database.Transaction
	expectIn  []string
	expectOut []string
}

func TestSnapshot_Success(t *testing.T) {
	cases := []snapshotCase{
		{
			name: "all fields populated",
			tx: &database.Transaction{
				ID:                  42,
				Title:               "lunch",
				SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(1)),
				SourceCurrency:      "USD",
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
				TagIDs:              pq.Int32Array{1, 2},
				TransactionDateTime: time.Now(),
				TransactionDateOnly: time.Now(),
			},
			expectIn: []string{"title", "source_amount", "source_currency", "tag_ids"},
			expectOut: []string{"id", "created_at", "updated_at", "deleted_at",
				"source_amount_in_base_currency", "destination_amount_in_base_currency"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			snap, err := history.Snapshot(tc.tx)
			require.NoError(t, err)
			for _, k := range tc.expectIn {
				_, ok := snap[k]
				assert.True(t, ok, "expected key %q in snapshot", k)
			}
			for _, k := range tc.expectOut {
				_, ok := snap[k]
				assert.False(t, ok, "expected key %q NOT in snapshot", k)
			}
		})
	}
}

func TestDiff_Success(t *testing.T) {
	a := map[string]any{"title": "old", "notes": "same"}
	b := map[string]any{"title": "new", "notes": "same"}
	diff, err := history.Diff(a, b)
	require.NoError(t, err)
	require.NotNil(t, diff)
	ops, ok := diff["ops"].([]any)
	require.True(t, ok)
	assert.Len(t, ops, 1)
}

func TestDiff_Empty(t *testing.T) {
	a := map[string]any{"title": "same"}
	diff, err := history.Diff(a, a)
	require.NoError(t, err)
	assert.Nil(t, diff)
}
