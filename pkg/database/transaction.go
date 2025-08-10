package database

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"time"
)

type Transaction struct {
	ID int64

	SourceAmount               decimal.NullDecimal
	SourceCurrency             string
	SourceAmountInBaseCurrency decimal.NullDecimal

	FxSourceAmount   decimal.NullDecimal // withdrawals only
	FxSourceCurrency string              // withdrawals only

	DestinationAmount               decimal.NullDecimal
	DestinationCurrency             string
	DestinationAmountInBaseCurrency decimal.NullDecimal

	SourceAccountID      *int32
	DestinationAccountID *int32

	TagIDs pq.Int32Array `gorm:"type:integer[]"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Notes string
	Extra map[string]string `gorm:"serializer:json"`

	TransactionDateTime time.Time `gorm:"type:timestamp"`
	TransactionDateOnly time.Time `gorm:"type:date"`

	TransactionType gomoneypbv1.TransactionType `gorm:"type:int"`
	Flags           TransactionFlags            `gorm:"type:bigint"`

	VoidedByTransactionID *int64
	Title                 string

	ReferenceNumber         *string
	InternalReferenceNumber *string
	CategoryID              *int32
}

type TransactionFlags int64
