package fetcher

import "github.com/shopspring/decimal"

type rateResponse struct {
	BaseCode        string                     `json:"base_code"`
	ConversionRates map[string]decimal.Decimal `json:"conversion_rates"`
}
