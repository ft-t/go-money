package database

import (
	"github.com/shopspring/decimal"
	"time"
)

type ExchangeRate struct {
	ID   string // Currency ID
	Rate decimal.Decimal

	UpdatedAt time.Time
}
