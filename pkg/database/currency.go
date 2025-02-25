package database

import (
	"github.com/shopspring/decimal"
	"time"
)

type Currency struct {
	ID   string // Currency ID
	Rate decimal.Decimal

	IsActive bool

	DecimalPlaces int32
	UpdatedAt     time.Time
}
