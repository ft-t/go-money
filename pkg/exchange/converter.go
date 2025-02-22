package exchange

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"time"
)

type Converter struct {
	cache *expirable.LRU[string, decimal.Decimal]
}

func NewConverter() *Converter {
	return &Converter{
		cache: expirable.NewLRU[string, decimal.Decimal](100, nil, 30*time.Second),
	}
}

func (c *Converter) Convert(
	ctx context.Context,
	fromCurrency string,
	toCurrency string,
	amount decimal.Decimal,
) (decimal.Decimal, error) {
	if fromCurrency == toCurrency {
		return amount, nil
	}

	rates, err := c.fetchRates(ctx, []string{fromCurrency, toCurrency, configuration.BaseCurrency})
	if err != nil {
		return decimal.Zero, err
	}

	toBaseRate, toBaseRateOk := rates[fromCurrency]
	if !toBaseRateOk {
		return decimal.Zero, errors.Newf("rate for %s not found", fromCurrency)
	}

	amountInBase := amount.Div(toBaseRate)

	toRate, toRateOk := rates[toCurrency]
	if !toRateOk {
		return decimal.Zero, errors.Newf("rate for %s not found", toCurrency)
	}

	return amountInBase.Mul(toRate), nil
}

func (c *Converter) fetchRates(
	ctx context.Context,
	currencies []string,
) (map[string]decimal.Decimal, error) {
	currencies = lo.Uniq(currencies)

	var missing []string

	var resp = make(map[string]decimal.Decimal)
	for _, currency := range currencies {
		if rate, ok := c.cache.Get(currency); ok {
			resp[currency] = rate
		} else {
			missing = append(missing, currency)
		}
	}

	if len(missing) == 0 {
		return resp, nil
	}

	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly))

	var rates []*database.ExchangeRate
	if err := db.Where("id IN ?", missing).Find(&rates).Error; err != nil {
		return resp, nil
	}

	for _, rate := range rates {
		resp[rate.ID] = rate.Rate
		c.cache.Add(rate.ID, rate.Rate)
	}

	return resp, nil
}
