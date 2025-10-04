package importers_test

import (
	"context"
	_ "embed"
	"testing"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/mono/chargeoff.csv
var monoChargeOff []byte

func TestMonoParseSimpleExpense(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mockConverter.EXPECT().Convert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(decimal.NewFromInt(1), nil).AnyTimes()

	mono := importers.NewMono(importers.NewBaseParser(mockConverter, nil, nil))

	sourceAccount := &database.Account{
		ID:            1,
		AccountNumber: "UAH",
		Currency:      "UAH",
		Type:          gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	}

	expenseAccount := &database.Account{
		ID:       2,
		Currency: "UAH",
		Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Flags:    database.AccountFlagIsDefault,
	}

	_, err := mono.Import(context.TODO(), &importers.ImportRequest{
		Data:     []string{string(monoChargeOff)},
		Accounts: []*database.Account{sourceAccount, expenseAccount},
	})

	assert.NoError(t, err)
}

func TestMonoParseEmptyFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mono := importers.NewMono(importers.NewBaseParser(mockConverter, nil, nil))

	resp, err := mono.Import(context.TODO(), &importers.ImportRequest{
		Data:     []string{"Header\n"},
		Accounts: []*database.Account{},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "empty file")
}

func TestMonoParseInvalidDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mono := importers.NewMono(importers.NewBaseParser(mockConverter, nil, nil))

	csvData := []byte(`Дата і час операції,Опис,MCC,Сума у валюті картки,Сума у валюті операції,Валюта операції,Курс,Баланс після операції
invalid-date,Test,5262,-100.00,10.00,USD,10.00,1000.00`)

	_, err := mono.Import(context.TODO(), &importers.ImportRequest{
		Data:     []string{string(csvData)},
		Accounts: []*database.Account{},
	})

	assert.NoError(t, err)
}

func TestMonoParseInvalidAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mono := importers.NewMono(importers.NewBaseParser(mockConverter, nil, nil))

	csvData := []byte(`Дата і час операції,Опис,MCC,Сума у валюті картки,Сума у валюті операції,Валюта операції,Курс,Баланс після операції
11.08.2024 12:19:14,Test,5262,invalid,10.00,USD,10.00,1000.00`)

	_, err := mono.Import(context.TODO(), &importers.ImportRequest{
		Data:     []string{string(csvData)},
		Accounts: []*database.Account{},
	})

	assert.NoError(t, err)
}

func TestMonoParseIncomeNotSupported(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mono := importers.NewMono(importers.NewBaseParser(mockConverter, nil, nil))

	csvData := []byte(`Дата і час операції,Опис,MCC,Сума у валюті картки,Сума у валюті операції,Валюта операції,Курс,Баланс після операції
11.08.2024 12:19:14,Test Income,5262,100.00,10.00,USD,10.00,1000.00`)

	_, err := mono.Import(context.TODO(), &importers.ImportRequest{
		Data:     []string{string(csvData)},
		Accounts: []*database.Account{},
	})

	assert.NoError(t, err)
}

func TestMonoParseMultipleTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mockConverter.EXPECT().Convert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(decimal.NewFromInt(1), nil).AnyTimes()

	mono := importers.NewMono(importers.NewBaseParser(mockConverter, nil, nil))

	sourceAccount := &database.Account{
		ID:            1,
		AccountNumber: "UAH",
		Currency:      "UAH",
		Type:          gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	}

	expenseAccount := &database.Account{
		ID:       2,
		Currency: "UAH",
		Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Flags:    database.AccountFlagIsDefault,
	}

	csvData := []byte(`Дата і час операції,Опис,MCC,Сума у валюті картки,Сума у валюті операції,Валюта операції,Курс,Баланс після операції
11.08.2024 12:19:14,Transaction 1,5262,-1231.79,128.71,PLN,9.57,1000.00
12.08.2024 13:30:00,Transaction 2,5411,-500.00,50.00,EUR,10.00,500.00`)

	_, err := mono.Import(context.TODO(), &importers.ImportRequest{
		Data:     []string{string(csvData)},
		Accounts: []*database.Account{sourceAccount, expenseAccount},
	})

	assert.NoError(t, err)
}

func TestMonoType(t *testing.T) {
	mono := importers.NewMono(nil)

	assert.Equal(t, importv1.ImportSource_IMPORT_SOURCE_UNSPECIFIED, mono.Type())
}
