package exchange

import (
	"context"
	"github.com/shopspring/decimal"
)

type Converter struct {
	cache expirable.Lru
}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) Convert(
	ctx context.Context,
	fromCurrency string,
	toCurrency string,
	amount decimal.Decimal,
) {

}

func (c *Converter) fetchRates(
	ctx context.Context,
	currencies []string,
) map[string]decimal.Decimal {

}
