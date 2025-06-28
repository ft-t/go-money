package rules

import (
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
