package database

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type Currency struct {
	ID   string // Currency ID
	Rate decimal.Decimal

	IsActive bool

	DecimalPlaces int32
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt
}

func (c *Currency) TableName() string {
	return "currencies"
}
