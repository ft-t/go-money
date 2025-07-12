package database_test

import (
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStruct(t *testing.T) {
	testCases := []struct {
		val      string
		expected string
	}{
		{
			val:      lo.ToPtr(database.Currency{}).TableName(),
			expected: "currencies",
		},
		{
			val:      lo.ToPtr(database.Tag{}).TableName(),
			expected: "tags",
		},
		{
			val:      lo.ToPtr(database.ImportDeduplication{}).TableName(),
			expected: "import_deduplication",
		},
		{
			val:      lo.ToPtr(database.DailyStat{}).TableName(),
			expected: "daily_stat",
		},
		{
			val:      lo.ToPtr(database.Category{}).TableName(),
			expected: "categories",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.val, func(t *testing.T) {
			assert.EqualValues(t, tc.expected, tc.val)
		})
	}
}
