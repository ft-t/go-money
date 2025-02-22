package database

import (
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

	SourceAccountID      *int
	DestinationAccountID *int

	LabelIDs []pq.Int32Array

	CreatedAt time.Time
}
