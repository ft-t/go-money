package database

import (
	"github.com/shopspring/decimal"
	"time"
)

type DailyStat struct {
	AccountID int32     `gorm:"primaryKey"`
	Date      time.Time `gorm:"primaryKey"`

	Amount decimal.Decimal
}

func (*DailyStat) TableName() string {
	return "daily_stat"
}

type MonthlyStat struct {
	AccountID int32     `gorm:"primaryKey"`
	Date      time.Time `gorm:"primaryKey"`

	Balance decimal.Decimal
}
