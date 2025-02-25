package mappers_test

import (
	"context"
	gomoneypbv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestMapper(t *testing.T) {
	decimalSvc := NewMockDecimalSvc(gomock.NewController(t))

	mapper := mappers.NewMapper(&mappers.MapperConfig{
		DecimalSvc: decimalSvc,
	})

	updatedTime := time.Now().UTC()
	createdTime := time.Now().UTC().Add(-time.Hour)
	deletedTime := time.Now().UTC().Add(-time.Minute)

	currentBalance := decimal.RequireFromString("100.1230")
	decimalSvc.EXPECT().ToString(gomock.Any(), currentBalance, "EUR").
		Return("100.1")

	mapped := mapper.MapAccount(context.TODO(), &database.Account{
		ID:             11,
		Name:           "name",
		Currency:       "EUR",
		CurrentBalance: currentBalance,
		Extra: map[string]string{
			"key": "value",
		},
		Flags:         22,
		LastUpdatedAt: updatedTime,
		CreatedAt:     createdTime,
		DeletedAt: gorm.DeletedAt{
			Time:  deletedTime,
			Valid: true,
		},
		Type:          gomoneypbv1.AccountType_ACCOUNT_TYPE_BROKERAGE,
		Note:          "note",
		AccountNumber: "number",
		Iban:          "iban",
		LiabilityPercent: decimal.NullDecimal{
			Decimal: decimal.RequireFromString("0.123"),
			Valid:   true,
		},
	})

	assert.EqualValues(t, 11, mapped.Id)
	assert.EqualValues(t, "name", mapped.Name)
	assert.EqualValues(t, "EUR", mapped.Currency)
	assert.EqualValues(t, "100.1", mapped.CurrentBalance)
	assert.EqualValues(t, map[string]string{
		"key": "value",
	}, mapped.Extra)
	assert.EqualValues(t, updatedTime, mapped.UpdatedAt.AsTime())
	assert.EqualValues(t, deletedTime, mapped.DeletedAt.AsTime())
	assert.EqualValues(t, gomoneypbv1.AccountType_ACCOUNT_TYPE_BROKERAGE, mapped.Type)
	assert.EqualValues(t, "note", mapped.Note)
	assert.EqualValues(t, "number", mapped.AccountNumber)
	assert.EqualValues(t, "iban", mapped.Iban)
	assert.EqualValues(t, "0.123", *mapped.LiabilityPercent)
}
