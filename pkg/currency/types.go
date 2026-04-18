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

type Quote struct {
	From         string
	To           string
	Amount       decimal.Decimal
	Converted    decimal.Decimal
	FromRate     decimal.Decimal // rate of From vs BaseCurrency
	ToRate       decimal.Decimal // rate of To vs BaseCurrency
	BaseCurrency string
}
