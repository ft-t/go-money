package database

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"time"
)

type WalletFlags = int64

type Account struct {
	ID int32

	Name     string
	Currency string // code

	CurrentBalance decimal.Decimal
	Extra          map[string]string `gorm:"serializer:json"`
	Flags          WalletFlags

	LastUpdatedAt time.Time
	CreatedAt     time.Time

	DeletedAt        gorm.DeletedAt
	Type             gomoneypbv1.AccountType
	Note             string
	AccountNumber    string
	Iban             string
	LiabilityPercent decimal.NullDecimal
	Position         int

	FirstTransactionAt *time.Time
}
