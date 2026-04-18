package currency

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type Converter struct {
	cache        *expirable.LRU[string, decimal.Decimal]
	baseCurrency string
}

func NewConverter(
	baseCurrency string,
) *Converter {
	return &Converter{
		cache:        expirable.NewLRU[string, decimal.Decimal](100, nil, configuration.DefaultCacheTTL),
		baseCurrency: baseCurrency,
	}
}

func (c *Converter) Convert(
	ctx context.Context,
	fromCurrency string,
	toCurrency string,
	amount decimal.Decimal,
) (decimal.Decimal, error) {
	quote, err := c.Quote(ctx, fromCurrency, toCurrency, amount)
	if err != nil {
		return decimal.Zero, err
	}
	return quote.Converted, nil
}

func (c *Converter) Quote(
	ctx context.Context,
	fromCurrency string,
	toCurrency string,
	amount decimal.Decimal,
) (*Quote, error) {
	quote := &Quote{
		From:         fromCurrency,
		To:           toCurrency,
		Amount:       amount,
		BaseCurrency: c.baseCurrency,
	}

	if fromCurrency == toCurrency {
		quote.Converted = amount
		quote.FromRate = decimal.NewFromInt(1)
		quote.ToRate = decimal.NewFromInt(1)
		return quote, nil
	}

	rates, err := c.fetchRates(ctx, []string{fromCurrency, toCurrency, c.baseCurrency})
	if err != nil {
		return nil, err
	}

	fromRate, ok := rates[fromCurrency]
	if !ok {
		return nil, errors.Newf("rate for %s not found", fromCurrency)
	}

	toRate, ok := rates[toCurrency]
	if !ok {
		return nil, errors.Newf("rate for %s not found", toCurrency)
	}

	quote.FromRate = fromRate
	quote.ToRate = toRate
	quote.Converted = amount.Div(fromRate).Mul(toRate)
	return quote, nil
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

	var rates []*database.Currency
	if err := db.Where("id IN ?", missing).Find(&rates).Error; err != nil {
		return resp, nil
	}

	for _, rate := range rates {
		resp[rate.ID] = rate.Rate
		c.cache.Add(rate.ID, rate.Rate)
	}

	return resp, nil
}
