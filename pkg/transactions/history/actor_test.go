package history_test

import (
	"context"
	"testing"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/history"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestUserActor_Success(t *testing.T) {
	a := history.UserActor(7)
	assert.Equal(t, database.TransactionHistoryActorTypeUser, a.Type)
	assert.Equal(t, lo.ToPtr(int32(7)), a.UserID)
	assert.Nil(t, a.RuleID)
	assert.Empty(t, a.Detail)
}

func TestImporterActor_Success(t *testing.T) {
	a := history.ImporterActor("firefly")
	assert.Equal(t, database.TransactionHistoryActorTypeImporter, a.Type)
	assert.Nil(t, a.UserID)
	assert.Nil(t, a.RuleID)
	assert.Equal(t, "firefly", a.Detail)
}

func TestSchedulerActor_Success(t *testing.T) {
	a := history.SchedulerActor(11)
	assert.Equal(t, database.TransactionHistoryActorTypeScheduler, a.Type)
	assert.Nil(t, a.UserID)
	assert.Equal(t, lo.ToPtr(int32(11)), a.RuleID)
}

func TestBulkActor_Success(t *testing.T) {
	a := history.BulkActor(3, "set_category")
	assert.Equal(t, database.TransactionHistoryActorTypeBulk, a.Type)
	assert.Equal(t, lo.ToPtr(int32(3)), a.UserID)
	assert.Equal(t, "set_category", a.Detail)
}

func TestRuleActor_Success(t *testing.T) {
	a := history.RuleActor(42)
	assert.Equal(t, database.TransactionHistoryActorTypeRule, a.Type)
	assert.Nil(t, a.UserID)
	assert.Equal(t, lo.ToPtr(int32(42)), a.RuleID)
}

func TestWithActor_RoundTrip_Success(t *testing.T) {
	want := history.UserActor(99)
	ctx := history.WithActor(context.Background(), want)

	got, ok := history.ActorFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, want, got)
}

func TestActorFromContext_Failure_Empty(t *testing.T) {
	_, ok := history.ActorFromContext(context.Background())
	assert.False(t, ok)
}

func TestActorFromContext_Failure_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "not-an-actor")

	_, ok := history.ActorFromContext(ctx)
	assert.False(t, ok)
}
