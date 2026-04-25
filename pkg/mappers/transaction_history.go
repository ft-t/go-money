package mappers

import (
	historyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/history/v1"
	"github.com/ft-t/go-money/pkg/database"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *Mapper) MapTransactionHistoryEvent(row *database.TransactionHistory) *historyv1.TransactionHistoryEvent {
	out := &historyv1.TransactionHistoryEvent{
		Id:            row.ID,
		TransactionId: row.TransactionID,
		EventType:     historyv1.TransactionHistoryEventType(row.EventType),
		ActorType:     historyv1.TransactionHistoryActorType(row.ActorType),
		OccurredAt:    timestamppb.New(row.OccurredAt),
		ActorUserId:   row.ActorUserID,
		ActorRuleId:   row.ActorRuleID,
		ActorExtra:    row.ActorExtra,
	}

	if row.Snapshot != nil {
		snap, err := structpb.NewStruct(row.Snapshot)
		if err == nil {
			out.Snapshot = snap
		}
	}

	if row.Diff != nil {
		diff, err := structpb.NewStruct(row.Diff)
		if err == nil {
			out.Diff = diff
		}
	}

	return out
}
