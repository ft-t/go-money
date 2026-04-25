package database

import "time"

type TransactionHistoryEventType int16

const (
	TransactionHistoryEventTypeCreated     TransactionHistoryEventType = 1
	TransactionHistoryEventTypeUpdated     TransactionHistoryEventType = 2
	TransactionHistoryEventTypeDeleted     TransactionHistoryEventType = 3
	TransactionHistoryEventTypeRuleApplied TransactionHistoryEventType = 4
)

type TransactionHistoryActorType int16

const (
	TransactionHistoryActorTypeUser      TransactionHistoryActorType = 1
	TransactionHistoryActorTypeRule      TransactionHistoryActorType = 2
	TransactionHistoryActorTypeScheduler TransactionHistoryActorType = 3
	TransactionHistoryActorTypeImporter  TransactionHistoryActorType = 4
	TransactionHistoryActorTypeBulk      TransactionHistoryActorType = 5
)

type TransactionHistory struct {
	ID            int64
	TransactionID int64
	EventType     TransactionHistoryEventType `gorm:"type:smallint"`
	ActorType     TransactionHistoryActorType `gorm:"type:smallint"`
	ActorUserID   *int32
	ActorRuleID   *int32
	ActorExtra    *string
	Snapshot      map[string]any `gorm:"type:jsonb;serializer:json"`
	Diff          map[string]any `gorm:"type:jsonb;serializer:json"`
	OccurredAt    time.Time
}

func (TransactionHistory) TableName() string { return "transaction_history" }
