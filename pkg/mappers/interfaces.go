package mappers

import (
	"context"
	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package mappers_test -source=interfaces.go

type DecimalSvc interface {
	ToString(ctx context.Context, amount decimal.Decimal, currency string) string
}
