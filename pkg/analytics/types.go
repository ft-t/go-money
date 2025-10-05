package analytics

import "github.com/shopspring/decimal"

type AccountSummary struct {
	TotalDebits  decimal.Decimal
	TotalCredits decimal.Decimal
	DebitsCount  int32
	CreditsCount int32
}
