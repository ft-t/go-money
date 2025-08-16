package database

import (
	"time"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AccountFlag = int64

const (
	AccountFlagIsDefault AccountFlag = 1 << 0
)

type Account struct {
	ID int32

	Name     string
	Currency string // code

	CurrentBalance decimal.Decimal
	Extra          map[string]string `gorm:"serializer:json"`
	Flags          AccountFlag

	LastUpdatedAt time.Time
	CreatedAt     time.Time

	DeletedAt        gorm.DeletedAt
	Type             gomoneypbv1.AccountType
	Note             string
	AccountNumber    string
	Iban             string
	LiabilityPercent decimal.NullDecimal
	DisplayOrder     *int32

	FirstTransactionAt *time.Time
}

func (a *Account) IsDefault() bool {
	return a.Flags&AccountFlagIsDefault == AccountFlagIsDefault
}
