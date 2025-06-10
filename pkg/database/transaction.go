package database

import (
	gomoneypbv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"time"
)

type Transaction struct {
	ID int64

	SourceAmount   decimal.Decimal
	SourceCurrency string

	DestinationAmount   decimal.Decimal
	DestinationCurrency string

	SourceAccountID      *int32
	DestinationAccountID *int32

	LabelIDs pq.Int32Array `gorm:"type:integer[]"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Notes string
	Extra map[string]string `gorm:"serializer:json"`

	TransactionDateTime time.Time `gorm:"type:timestamp"`
	TransactionDateOnly time.Time `gorm:"type:date"`

	TransactionType gomoneypbv1.TransactionType `gorm:"type:int"`
	Flags           TransactionFlags            `gorm:"type:bigint"`
	
	VoidedByTransactionID *int64
}

type TransactionFlags int64
