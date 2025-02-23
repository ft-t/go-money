package database

import (
	"github.com/shopspring/decimal"
	"time"
)

type DailyStat struct {
	AccountID int32     `gorm:"primaryKey"`
	Date      time.Time `gorm:"primaryKey"`

	Balance decimal.Decimal
}

type MonthlyStat struct {
	AccountID int32     `gorm:"primaryKey"`
	Date      time.Time `gorm:"primaryKey"`

	Balance decimal.Decimal
}
