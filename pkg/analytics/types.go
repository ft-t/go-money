package analytics

import (
	"context"
	"github.com/shopspring/decimal"
)

type AccountSummary struct {
	TotalDebits  decimal.Decimal
	TotalCredits decimal.Decimal
	DebitsCount  int32
	CreditsCount int32
}

type DecimalSvc interface {
	ToString(ctx context.Context, amount decimal.Decimal, currency string) string
}
