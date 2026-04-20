package mappers_test

import (
	"testing"
	"time"

	historyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/history/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapTransactionHistoryEvent_Success(t *testing.T) {
	occurred := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name       string
		row        *database.TransactionHistory
		wantSnap   bool
		wantDiff   bool
		wantUserID *int32
		wantRuleID *int32
		wantExtra  *string
	}{
		{
			name: "created no diff",
			row: &database.TransactionHistory{
				ID:            1,
				TransactionID: 100,
				EventType:     database.TransactionHistoryEventTypeCreated,
				ActorType:     database.TransactionHistoryActorTypeUser,
				ActorUserID:   lo.ToPtr(int32(7)),
				Snapshot:      map[string]any{"title": "first", "amount": float64(12.5)},
				OccurredAt:    occurred,
			},
			wantSnap:   true,
			wantDiff:   false,
			wantUserID: lo.ToPtr(int32(7)),
		},
		{
			name: "updated with diff",
			row: &database.TransactionHistory{
				ID:            2,
				TransactionID: 100,
				EventType:     database.TransactionHistoryEventTypeUpdated,
				ActorType:     database.TransactionHistoryActorTypeUser,
				ActorUserID:   lo.ToPtr(int32(7)),
				Snapshot:      map[string]any{"title": "second"},
				Diff: map[string]any{
					"ops": []any{
						map[string]any{"op": "replace", "path": "/title", "value": "second"},
					},
				},
				OccurredAt: occurred,
			},
			wantSnap:   true,
			wantDiff:   true,
			wantUserID: lo.ToPtr(int32(7)),
		},
		{
			name: "rule applied with actor rule id",
			row: &database.TransactionHistory{
				ID:            3,
				TransactionID: 101,
				EventType:     database.TransactionHistoryEventTypeRuleApplied,
				ActorType:     database.TransactionHistoryActorTypeRule,
				ActorRuleID:   lo.ToPtr(int32(55)),
				Snapshot:      map[string]any{"title": "ruled"},
				OccurredAt:    occurred,
			},
			wantSnap:   true,
			wantDiff:   false,
			wantRuleID: lo.ToPtr(int32(55)),
		},
		{
			name: "importer with actor extra",
			row: &database.TransactionHistory{
				ID:            4,
				TransactionID: 102,
				EventType:     database.TransactionHistoryEventTypeCreated,
				ActorType:     database.TransactionHistoryActorTypeImporter,
				ActorExtra:    lo.ToPtr("firefly"),
				Snapshot:      map[string]any{"title": "imp"},
				OccurredAt:    occurred,
			},
			wantSnap:  true,
			wantDiff:  false,
			wantExtra: lo.ToPtr("firefly"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decimalSvc := NewMockDecimalSvc(gomock.NewController(t))
			mapper := mappers.NewMapper(&mappers.MapperConfig{DecimalSvc: decimalSvc})

			got := mapper.MapTransactionHistoryEvent(tc.row)

			require.NotNil(t, got)
			assert.Equal(t, tc.row.ID, got.Id)
			assert.Equal(t, tc.row.TransactionID, got.TransactionId)
			assert.Equal(t, historyv1.TransactionHistoryEventType(tc.row.EventType), got.EventType)
			assert.Equal(t, historyv1.TransactionHistoryActorType(tc.row.ActorType), got.ActorType)
			assert.Equal(t, tc.wantUserID, got.ActorUserId)
			assert.Equal(t, tc.wantRuleID, got.ActorRuleId)
			assert.Equal(t, tc.wantExtra, got.ActorExtra)
			assert.Equal(t, occurred, got.OccurredAt.AsTime())
			assert.Equal(t, tc.wantSnap, got.Snapshot != nil)
			assert.Equal(t, tc.wantDiff, got.Diff != nil)
		})
	}
}

func TestMapTransactionHistoryEvent_SnapshotFields(t *testing.T) {
	decimalSvc := NewMockDecimalSvc(gomock.NewController(t))
	mapper := mappers.NewMapper(&mappers.MapperConfig{DecimalSvc: decimalSvc})

	row := &database.TransactionHistory{
		ID:            9,
		TransactionID: 200,
		EventType:     database.TransactionHistoryEventTypeUpdated,
		ActorType:     database.TransactionHistoryActorTypeUser,
		Snapshot:      map[string]any{"title": "t", "flag": true, "count": float64(3)},
		Diff: map[string]any{
			"ops": []any{map[string]any{"op": "replace", "path": "/title", "value": "t"}},
		},
		OccurredAt: time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC),
	}

	got := mapper.MapTransactionHistoryEvent(row)

	require.NotNil(t, got.Snapshot)
	snapMap := got.Snapshot.AsMap()
	assert.Equal(t, "t", snapMap["title"])
	assert.Equal(t, true, snapMap["flag"])
	assert.Equal(t, float64(3), snapMap["count"])

	require.NotNil(t, got.Diff)
	diffMap := got.Diff.AsMap()
	ops, ok := diffMap["ops"].([]any)
	require.True(t, ok)
	require.Len(t, ops, 1)
	op0, ok := ops[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "replace", op0["op"])
	assert.Equal(t, "/title", op0["path"])
	assert.Equal(t, "t", op0["value"])
}
