package mappers_test

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTransactionMapper(t *testing.T) {
	decimalSvc := NewMockDecimalSvc(gomock.NewController(t))

	mapper := mappers.NewMapper(&mappers.MapperConfig{
		DecimalSvc: decimalSvc,
	})

	decimalSvc.EXPECT().ToString(gomock.Any(), gomock.Any(), "PLN").Return("11.123")
	decimalSvc.EXPECT().ToString(gomock.Any(), gomock.Any(), "USD").Return("22.456")

	txDate := time.Now().Add(22 * time.Hour).UTC()

	mapped := mapper.MapTransaction(context.TODO(), &database.Transaction{
		ID:                    555,
		SourceAmount:          decimal.NewNullDecimal(decimal.NewFromInt(11)),
		SourceCurrency:        "PLN",
		DestinationAmount:     decimal.NewNullDecimal(decimal.NewFromInt(22)),
		DestinationCurrency:   "USD",
		SourceAccountID:       lo.ToPtr(int32(1)),
		DestinationAccountID:  lo.ToPtr(int32(2)),
		TagIDs:                []int32{1, 2, 3},
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		Notes:                 "Test transaction",
		Extra:                 map[string]string{"key": "value"},
		TransactionDateTime:   txDate,
		TransactionDateOnly:   txDate,
		TransactionType:       gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT,
		Flags:                 12,
		VoidedByTransactionID: lo.ToPtr(int64(123)),
		Title:                 "xxx",
	})

	assert.EqualValues(t, "xxx", mapped.Title)
	assert.EqualValues(t, 555, mapped.Id)
	assert.EqualValues(t, "11.123", *mapped.SourceAmount)
	assert.EqualValues(t, "PLN", *mapped.SourceCurrency)

	assert.EqualValues(t, "22.456", *mapped.DestinationAmount)
	assert.EqualValues(t, "USD", *mapped.DestinationCurrency)
	assert.EqualValues(t, 1, *mapped.SourceAccountId)
	assert.EqualValues(t, 2, *mapped.DestinationAccountId)
	assert.EqualValues(t, []int32{1, 2, 3}, mapped.LabelIds)
	assert.EqualValues(t, "Test transaction", mapped.Notes)
	assert.EqualValues(t, map[string]string{"key": "value"}, mapped.Extra)
	assert.EqualValues(t, gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT, mapped.Type)
	assert.NotNil(t, mapped.CreatedAt)
	assert.EqualValues(t, txDate, mapped.TransactionDate.AsTime())
}
