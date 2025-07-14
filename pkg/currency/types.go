package currency

import "github.com/shopspring/decimal"

type RemoteRates struct {
	Base  string                     `json:"b"`
	Rates map[string]decimal.Decimal `json:"r"`
}
