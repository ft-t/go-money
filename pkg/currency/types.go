package currency

import (
	"github.com/shopspring/decimal"
	"time"
)

type RemoteRates struct {
	Base      string                     `json:"b"`
	Rates     map[string]decimal.Decimal `json:"r"`
	UpdatedAt time.Time                  `json:"u"`
}
