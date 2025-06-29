package rules

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package rules_test -source=interfaces.go

type Interpreter interface {
	Run(
		_ context.Context,
		script string,
		clonedTx *database.Transaction,
	) (bool, error)
}

type MapperSvc interface {
	MapRule(rule *database.Rule) *gomoneypbv1.Rule
	MapTransaction(ctx context.Context, tx *database.Transaction) *gomoneypbv1.Transaction
}

type ExecutorSvc interface {
	ProcessSingleRule(
		ctx context.Context,
		inputTx *database.Transaction,
		rule *database.Rule,
	) (bool, *database.Transaction, error)
}

type TransactionSvc interface {
	GetTransactionByIDs(
		ctx context.Context,
		ids []int64,
	) ([]*database.Transaction, error)
}
