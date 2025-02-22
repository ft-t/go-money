package exchange

import "github.com/shopspring/decimal"

type remoteRates struct {
	Base  string
	Rates map[string]decimal.Decimal
}
