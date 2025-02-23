package mappers

import (
	"context"
	"github.com/shopspring/decimal"
)

type DecimalSvc interface {
	ToString(ctx context.Context, amount decimal.Decimal, currency string) string
}
