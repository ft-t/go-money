package database

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type DoubleEntry struct {
	ID            int64
	TransactionID int64
	IsDebit       bool

	AmountInBaseCurrency decimal.Decimal
	BaseCurrency         string

	AccountID int32

	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}
