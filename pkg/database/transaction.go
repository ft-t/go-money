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

	LabelIDs pq.Int32Array

	CreatedAt time.Time
	Notes     string
	Extra     map[string]string

	TransactionDate time.Time
	TransactionType gomoneypbv1.TransactionType
}
