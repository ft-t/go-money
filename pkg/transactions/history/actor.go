package history

import (
	"context"

	"github.com/ft-t/go-money/pkg/database"
)

type Actor struct {
	Type   database.TransactionHistoryActorType
	UserID *int32
	RuleID *int32
	Extra  string
}

type actorCtxKey struct{}

func WithActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, actorCtxKey{}, a)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(actorCtxKey{}).(Actor)
	return a, ok
}

func UserActor(userID int32) Actor {
	return Actor{Type: database.TransactionHistoryActorTypeUser, UserID: &userID}
}

func ImporterActor(name string) Actor {
	return Actor{Type: database.TransactionHistoryActorTypeImporter, Extra: name}
}

func SchedulerActor(ruleID int32) Actor {
	return Actor{Type: database.TransactionHistoryActorTypeScheduler, RuleID: &ruleID}
}

func BulkActor(userID int32, op string) Actor {
	return Actor{Type: database.TransactionHistoryActorTypeBulk, UserID: &userID, Extra: op}
}

func RuleActor(ruleID int32) Actor {
	return Actor{Type: database.TransactionHistoryActorTypeRule, RuleID: &ruleID}
}
