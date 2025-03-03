package currency

import "github.com/shopspring/decimal"

type remoteRates struct {
	Base  string                     `json:"b"`
	Rates map[string]decimal.Decimal `json:"r"`
}
