package importers_test

import (
	"context"
	"testing"
	"time"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestParseHeaderDate(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		wantTime    time.Time
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid header with AM",
			header:   "PrivatBank, [10/1/2025 9:50 AM]",
			wantTime: time.Date(2025, 10, 1, 9, 50, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid header with PM",
			header:   "PrivatBank, [10/1/2025 9:50 PM]",
			wantTime: time.Date(2025, 10, 1, 21, 50, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid header with single digit day",
			header:   "PrivatBank, [1/5/2025 3:15 PM]",
			wantTime: time.Date(2025, 1, 5, 15, 15, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid header with double digit day and month",
			header:   "PrivatBank, [12/31/2025 11:59 PM]",
			wantTime: time.Date(2025, 12, 31, 23, 59, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:        "missing opening bracket",
			header:      "PrivatBank, 10/1/2025 9:50 AM]",
			wantErr:     true,
			errContains: "invalid header format",
		},
		{
			name:        "missing closing bracket",
			header:      "PrivatBank, [10/1/2025 9:50 AM",
			wantErr:     true,
			errContains: "invalid header format",
		},
		{
			name:        "invalid date format",
			header:      "PrivatBank, [2025-10-01 09:50]",
			wantErr:     true,
			errContains: "failed to parse date",
		},
		{
			name:        "empty brackets",
			header:      "PrivatBank, []",
			wantErr:     true,
			errContains: "failed to parse date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p24 := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))
			gotTime, err := p24.ParseHeaderDate(tt.header)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.ErrorContains(t, err, tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTime, gotTime)
			}
		})
	}
}

func TestExpensesShouldNotMerged(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(`89.80PLN Ресторани, кафе, бари. Pyszne.pl, Wroclaw
4*67 16:17
Бал. 1.86USD
Курс 0.2547 USD/PLN`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
		{
			Data: []byte(`8.98PLN Ресторани, кафе, бари. Pyszne.pl, Wroclaw
4*67 16:18
Бал. 236.57USD
Курс 0.2550 USD/PLN`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 2)

	assert.Equal(t, "22.87", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].SourceCurrency)
	assert.Equal(t, "4*67", resp[0].SourceAccount)

	assert.Equal(t, "89.80", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "", resp[0].DestinationAccount)
	assert.Equal(t, "PLN", resp[0].DestinationCurrency)

	assert.Equal(t, "Ресторани, кафе, бари. Pyszne.pl, Wroclaw", resp[0].Description)
	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)

	assert.Equal(t, "2.29", resp[1].SourceAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[1].SourceCurrency)
	assert.Equal(t, "4*67", resp[1].SourceAccount)

	assert.Equal(t, "8.98", resp[1].DestinationAmount.StringFixed(2))
	assert.Equal(t, "", resp[1].DestinationAccount)
	assert.Equal(t, "PLN", resp[1].DestinationCurrency)

	assert.Equal(t, "Ресторани, кафе, бари. Pyszne.pl, Wroclaw", resp[1].Description)
	assert.Equal(t, importers.TransactionTypeExpense, resp[1].Type)
}

func TestTransferBetweenOwnCards(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("order 1", func(t *testing.T) {
		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(`1100.00USD Переказ зі своєї карти
4*59 22:40
Бал. 1.21USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(`1100.00USD Зарахування переказу на картку
4*71 22:40
Бал. 1.60USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "1100.00", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)
		assert.Equal(t, "4*59", resp[0].SourceAccount)

		assert.Equal(t, "1100.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "4*71", resp[0].DestinationAccount)
		assert.Equal(t, "USD", resp[0].DestinationCurrency)

		assert.Equal(t, "Переказ зі своєї карти", resp[0].Description)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
	t.Run("order 2", func(t *testing.T) {
		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(`1100.00USD Зарахування переказу на картку
4*71 22:40
Бал. 1.60USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(`1100.00USD Переказ зі своєї карти
4*59 22:40
Бал. 1.21USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "1100.00", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)
		assert.Equal(t, "4*59", resp[0].SourceAccount)

		assert.Equal(t, "1100.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "4*71", resp[0].DestinationAccount)
		assert.Equal(t, "USD", resp[0].DestinationCurrency)

		assert.Equal(t, "Зарахування переказу на картку", resp[0].Description)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
}

func TestTransferBetweenOwnCards3(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("order 1", func(t *testing.T) {
		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(`40.00USD Зарахування переказу. *1959
4*71 11:11
Бал. 1.37USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(`40.00USD Переказ на свою картку. *6471
4*59 11:11
Бал. 1446.74USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "40.00", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)
		assert.Equal(t, "4*59", resp[0].SourceAccount)

		assert.Equal(t, "40.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "4*71", resp[0].DestinationAccount)
		assert.Equal(t, "USD", resp[0].DestinationCurrency)

		assert.Equal(t, "Зарахування переказу. *1959", resp[0].Description)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
	t.Run("order 2", func(t *testing.T) {
		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(`40.00USD Переказ на свою картку. *6471
4*59 11:11
Бал. 1446.74USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(`40.00USD Зарахування переказу. *1959
4*71 11:11
Бал. 1.37USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "40.00", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)
		assert.Equal(t, "4*59", resp[0].SourceAccount)

		assert.Equal(t, "40.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "4*71", resp[0].DestinationAccount)
		assert.Equal(t, "USD", resp[0].DestinationCurrency)

		assert.Equal(t, "Переказ на свою картку. *6471", resp[0].Description)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
}

func TestTransferBetweenOwnCards2(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("order 1", func(t *testing.T) {
		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(`356.45USD Переказ на свою картку. *0320
4*59 05:06
Бал. 123.96USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(`14418.00UAH Переказ зі своєї картки *1959
5*20 05:06
Бал. 123.00UAH`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "356.45", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)
		assert.Equal(t, "4*59", resp[0].SourceAccount)

		assert.Equal(t, "14418.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "5*20", resp[0].DestinationAccount)
		assert.Equal(t, "UAH", resp[0].DestinationCurrency)

		assert.Equal(t, "Переказ на свою картку. *0320", resp[0].Description)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
	t.Run("order 2", func(t *testing.T) {
		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(`14418.00UAH Переказ зі своєї картки *1959
5*20 05:06
Бал. 123.00UAH`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(`356.45USD Переказ на свою картку. *0320
4*59 05:06
Бал. 123.96USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "356.45", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)
		assert.Equal(t, "4*59", resp[0].SourceAccount)

		assert.Equal(t, "14418.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "5*20", resp[0].DestinationAccount)
		assert.Equal(t, "UAH", resp[0].DestinationCurrency)

		assert.Equal(t, "Переказ зі своєї картки *1959", resp[0].Description)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
}

func TestConvertCurrency3(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("opt1", func(t *testing.T) {
		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(`5200.00UAH Переказ зі своєї карти 46**59 через додаток Приват24
5*20 18:51
Бал. 1.84UAH`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(`133.78USD Переказ на свою картку через додаток Приват24
4*59 18:51
Бал. 1.95USD`),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "133.78", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)
		assert.Equal(t, "4*59", resp[0].SourceAccount)

		assert.Equal(t, "5200.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "5*20", resp[0].DestinationAccount)
		assert.Equal(t, "UAH", resp[0].DestinationCurrency)

		assert.Equal(t, "Переказ зі своєї карти 46**59 через додаток Приват24", resp[0].Description)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
}

func TestParseCurrencyExchange4(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(`1723.64USD Переказ на свою карту через Приват24
4*67 15:16
Бал. 123.59USD`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
		{
			Data: []byte(`65550.00UAH Переказ зі своєї картки 46**67 через Приват24
5*20 15:16
Бал. 123.59UAH
Кред. лiмiт 300000.0UAH`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "1723.64", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].SourceCurrency)
	assert.Equal(t, "4*67", resp[0].SourceAccount)

	assert.Equal(t, "65550.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "5*20", resp[0].DestinationAccount)
	assert.Equal(t, "UAH", resp[0].DestinationCurrency)

	assert.Equal(t, "Переказ на свою карту через Приват24", resp[0].Description)
	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
}

func TestParseCurrencyExchange3(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(`65550.00UAH Переказ зі своєї картки 46**67 через Приват24
5*20 15:16
Бал. 123.59UAH
Кред. лiмiт 300000.0UAH`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
		{
			Data: []byte(`1723.64USD Переказ на свою карту через Приват24
4*67 15:16
Бал. 123.59USD`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "1723.64", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].SourceCurrency)
	assert.Equal(t, "4*67", resp[0].SourceAccount)

	assert.Equal(t, "65550.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "5*20", resp[0].DestinationAccount)
	assert.Equal(t, "UAH", resp[0].DestinationCurrency)

	assert.Equal(t, "Переказ зі своєї картки 46**67 через Приват24", resp[0].Description)
	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
}

func TestParseSimpleExpense(t *testing.T) {
	input := `1.33USD Розваги. Steam
4*71 16:27
Бал. 1.55USD`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "1.33", resp[0].SourceAmount.String())
	assert.Equal(t, "USD", resp[0].SourceCurrency)
	assert.Equal(t, "Розваги. Steam", resp[0].Description)
	assert.Equal(t, "4*71", resp[0].SourceAccount)
	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
}

func TestParseSimpleRefund(t *testing.T) {
	input := `120.52UAH Повернення. Транспорт. xx.yy
4*68 15:09
Бал. 329.89UAH`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "120.52", resp[0].DestinationAmount.String())
	assert.Equal(t, "UAH", resp[0].DestinationCurrency)
	assert.Equal(t, "Повернення. Транспорт. xx.yy", resp[0].Description)
	assert.Equal(t, "4*68", resp[0].DestinationAccount)
	assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)
}

func TestNewTransfer(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(`40.00USD Переказ на свою карту через Приват24
4*59 22:28
Бал. 1486.74USD`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
		{
			Data: []byte(`40.00USD Зарахування переказу через Приват24 зі своєї картки
4*71 22:28
Бал. 61.20USD`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "40", resp[0].DestinationAmount.String())
	assert.Equal(t, "USD", resp[0].DestinationCurrency)
	assert.Equal(t, "Переказ на свою карту через Приват24", resp[0].Description)
	assert.Equal(t, "4*71", resp[0].DestinationAccount)

	assert.Equal(t, "4*59", resp[0].SourceAccount)
	assert.Equal(t, "40", resp[0].SourceAmount.String())
	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
}

func TestNewTransfer2(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{

		{
			Data: []byte(`40.00USD Зарахування переказу через Приват24 зі своєї картки
4*71 22:28
Бал. 61.20USD`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
		{
			Data: []byte(`40.00USD Переказ на свою карту через Приват24
4*59 22:28
Бал. 1486.74USD`),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "40", resp[0].DestinationAmount.String())
	assert.Equal(t, "USD", resp[0].DestinationCurrency)
	assert.Equal(t, "Зарахування переказу через Приват24 зі своєї картки", resp[0].Description)
	assert.Equal(t, "4*71", resp[0].DestinationAccount)

	assert.Equal(t, "4*59", resp[0].SourceAccount)
	assert.Equal(t, "40", resp[0].SourceAmount.String())
	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
}

func TestIncomeTransferP2P(t *testing.T) {
	input := `1466.60USD Зарахування переказу з картки через Приват24
4*51 22:00
Бал. 1.89USD`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "1466.60", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].DestinationCurrency)
	assert.Equal(t, "Зарахування переказу з картки через Приват24", resp[0].Description)
	assert.Equal(t, "4*51", resp[0].DestinationAccount)
	assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)

	_, err = srv.Merge(context.TODO(), resp)
	assert.NoError(t, err)
}

func TestCreditExpense(t *testing.T) {
	input := `1466.60UAH Списання
4*40 20.05.24`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "1466.60", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].SourceCurrency)
	assert.Equal(t, "Списання", resp[0].Description)
	assert.Equal(t, "4*40", resp[0].SourceAccount)
	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)

	_, err = srv.Merge(context.TODO(), resp)
	assert.NoError(t, err)
}

func TestPartialRefund(t *testing.T) {
	input := `1.01USD Зарахування
4*71 09.03.24`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "1.01", resp[0].DestinationAmount.String())
	assert.Equal(t, "USD", resp[0].DestinationCurrency)
	assert.Equal(t, "Зарахування", resp[0].Description)
	assert.Equal(t, "4*71", resp[0].DestinationAccount)
	assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)
}

func TestParseSimpleExpense2(t *testing.T) {
	input := `83.69PLN Інтернет-магазини. AliExpress
4*67 12:19
Бал. 12.82USD
Курс 0.2558 USD/PLN`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "83.69", resp[0].DestinationAmount.String())
	assert.Equal(t, "PLN", resp[0].DestinationCurrency)

	assert.Equal(t, "21.41", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].SourceCurrency)
	assert.Equal(t, "Інтернет-магазини. AliExpress", resp[0].Description)
	assert.Equal(t, "4*67", resp[0].SourceAccount)
	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
}

func TestMultiCurrencyExpense(t *testing.T) {
	input := `249.00UAH Комуналка та Інтернет. Sweet TV
4*71 22:19
Бал. 45.45USD
Курс 37.3873 UAH/USD`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "249.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].DestinationCurrency)

	assert.Equal(t, "6.66", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].SourceCurrency)
	assert.Equal(t, "Комуналка та Інтернет. Sweet TV", resp[0].Description)
	assert.Equal(t, "4*71", resp[0].SourceAccount)
	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
}

func TestParseRemoteTransfer(t *testing.T) {
	input := `1.00UAH Переказ через Приват24 Одержувач: Імя Фамілія ПоБатькові
4*68 16:41
Бал. 17.81UAH`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "1.00", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].SourceCurrency)
	assert.Equal(t, "Переказ через Приват24 Одержувач: Імя Фамілія ПоБатькові", resp[0].Description)
	assert.Equal(t, "4*68", resp[0].SourceAccount)
	assert.Equal(t, importers.TransactionTypeRemoteTransfer, resp[0].Type)
}

func TestParseRemoteTransfer3(t *testing.T) {
	input := `715.06UAH Переказ зі своєї карти
5*20 19:29
Бал. 111.29UAH
Кред. лiмiт 111.0UAH`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "715.06", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].SourceCurrency)
	assert.Equal(t, "Переказ зі своєї карти", resp[0].Description)
	assert.Equal(t, "5*20", resp[0].SourceAccount)
	assert.Equal(t, importers.TransactionTypeRemoteTransfer, resp[0].Type)
}

func TestParseRemoteTransfer2(t *testing.T) {
	input := `4000.00UAH Переказ через додаток Приват24. Одержувач: ХХ УУ ММ
5*20 12:58
Бал. 123.32UAH
Кред. лiмiт 3000000.0UAH`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "4000.00", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].SourceCurrency)
	assert.Equal(t, "Переказ через додаток Приват24. Одержувач: ХХ УУ ММ", resp[0].Description)
	assert.Equal(t, "5*20", resp[0].SourceAccount)
	assert.Equal(t, importers.TransactionTypeRemoteTransfer, resp[0].Type)
}

func TestParseInternalTransferTo(t *testing.T) {
	input := `1.00UAH Переказ на свою карту 51**20 через додаток Приват24
4*68 16:13
Бал. 18.81UAH`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "1.00", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].SourceCurrency)
	assert.Equal(t, "Переказ на свою карту 51**20 через додаток Приват24", resp[0].Description)
	assert.Equal(t, "4*68", resp[0].SourceAccount)
	assert.Equal(t, "5*20", resp[0].DestinationAccount)
	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	assert.True(t, resp[0].InternalTransferDirectionTo)
}

func TestParseInternalTransferFrom(t *testing.T) {
	input := `1.00UAH Переказ зі своєї карти 47**68 через додаток Приват24
5*20 16:13
Бал. 123.32UAH`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp, 1)

	assert.Equal(t, "0.00", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "", resp[0].SourceCurrency)

	assert.Equal(t, "1.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].DestinationCurrency)

	assert.Equal(t, "Переказ зі своєї карти 47**68 через додаток Приват24", resp[0].Description)
	assert.Equal(t, "4*68", resp[0].SourceAccount)
	assert.Equal(t, "5*20", resp[0].DestinationAccount)
	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	assert.False(t, resp[0].InternalTransferDirectionTo)
}

func TestParseInternalTransferFromUSD(t *testing.T) {
	input := `1.00USD Переказ зі своєї карти 52**20 через додаток Приват24
4*71 23:50
Бал. 89.19USD`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "1.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].DestinationCurrency)
	assert.Equal(t, "Переказ зі своєї карти 52**20 через додаток Приват24", resp[0].Description)
	assert.Equal(t, "5*20", resp[0].SourceAccount)
	assert.Equal(t, "4*71", resp[0].DestinationAccount)
	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	assert.False(t, resp[0].InternalTransferDirectionTo)
}

func TestParseIncomeTransfer(t *testing.T) {
	input := `123.11UAH Переказ через Приват24 Відправник: Імя Фамілія ПоБатькові
5*20 20:11
Бал. 11111.22UAH`

	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
		{
			Data: []byte(input),
			Message: &importers.Message{
				CreatedAt: time.Now(),
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.Equal(t, "123.11", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "UAH", resp[0].DestinationCurrency)
	assert.Equal(t, "Переказ через Приват24 Відправник: Імя Фамілія ПоБатькові", resp[0].Description)
	assert.Equal(t, "5*20", resp[0].DestinationAccount)
	assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)
}

func TestMerger(t *testing.T) {
	t.Run("firstIsTo", func(t *testing.T) {
		pr := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

		txList := []*importers.Transaction{
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				SourceCurrency:              "UAH",
				SourceAccount:               "4*68",
				SourceAmount:                decimal.RequireFromString("1.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: true,
			},
			{
				ID:             uuid.NewString(),
				Type:           importers.TransactionTypeExpense,
				SourceCurrency: "USD",
				SourceAccount:  "4*71",
				SourceAmount:   decimal.RequireFromString("1.33"),
			},
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				SourceCurrency:              "UAH",
				SourceAccount:               "4*68",
				SourceAmount:                decimal.RequireFromString("1.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: false,
			},
		}
		resp, err := pr.Merge(context.TODO(), txList)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)

		assert.Equal(t, txList[0].ID, resp[0].ID)
		assert.Len(t, resp[0].DuplicateTransactions, 1)
		assert.Equal(t, txList[2].ID, resp[0].DuplicateTransactions[0].ID)

		assert.Equal(t, txList[1].ID, resp[1].ID)
	})

	t.Run("firstIsFrom", func(t *testing.T) {
		pr := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

		txList := []*importers.Transaction{
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				SourceCurrency:              "UAH",
				SourceAccount:               "4*68",
				SourceAmount:                decimal.RequireFromString("1.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: false,
			},
			{
				ID:             uuid.NewString(),
				Type:           importers.TransactionTypeExpense,
				SourceCurrency: "USD",
				SourceAccount:  "4*71",
				SourceAmount:   decimal.RequireFromString("1.33"),
			},
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				SourceCurrency:              "UAH",
				SourceAccount:               "4*68",
				SourceAmount:                decimal.RequireFromString("1.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: true,
			},
		}
		resp, err := pr.Merge(context.TODO(), txList)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)

		assert.Equal(t, txList[0].ID, resp[0].ID)
		assert.Len(t, resp[0].DuplicateTransactions, 1)
		assert.Equal(t, txList[2].ID, resp[0].DuplicateTransactions[0].ID)

		assert.Equal(t, txList[1].ID, resp[1].ID)
	})

	t.Run("multi currency", func(t *testing.T) {
		pr := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

		txList := []*importers.Transaction{
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				DestinationCurrency:         "USD",
				SourceAccount:               "4*68",
				DestinationAmount:           decimal.RequireFromString("1.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: false,
			},
			{
				ID:             uuid.NewString(),
				Type:           importers.TransactionTypeExpense,
				SourceCurrency: "USD",
				SourceAccount:  "4*71",
				SourceAmount:   decimal.RequireFromString("1.33"),
			},
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				SourceCurrency:              "UAH",
				SourceAccount:               "4*68",
				SourceAmount:                decimal.RequireFromString("38.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: true,
			},
		}
		resp, err := pr.Merge(context.TODO(), txList)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)

		assert.Equal(t, txList[0].ID, resp[0].ID)
		assert.Len(t, resp[0].DuplicateTransactions, 1)
		assert.Equal(t, txList[2].ID, resp[0].DuplicateTransactions[0].ID)

		assert.Equal(t, txList[1].ID, resp[1].ID)

		assert.Equal(t, "USD", resp[0].DestinationCurrency)
		assert.Equal(t, "UAH", resp[0].SourceCurrency)

		assert.Equal(t, "1.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "38.00", resp[0].SourceAmount.StringFixed(2))
	})

	t.Run("multi currency 2", func(t *testing.T) {
		pr := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

		txList := []*importers.Transaction{
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				SourceCurrency:              "UAH",
				SourceAccount:               "4*68",
				SourceAmount:                decimal.RequireFromString("38.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: true,
			},
			{
				ID:                          uuid.NewString(),
				Type:                        importers.TransactionTypeInternalTransfer,
				DestinationCurrency:         "USD",
				SourceAccount:               "4*68",
				DestinationAmount:           decimal.RequireFromString("1.00"),
				DestinationAccount:          "5*20",
				DateFromMessage:             "16:13",
				InternalTransferDirectionTo: false,
			},
			{
				ID:             uuid.NewString(),
				Type:           importers.TransactionTypeExpense,
				SourceCurrency: "USD",
				SourceAccount:  "4*71",
				SourceAmount:   decimal.RequireFromString("1.33"),
			},
		}
		resp, err := pr.Merge(context.TODO(), txList)

		assert.NoError(t, err)
		assert.Len(t, resp, 2)

		assert.Equal(t, txList[0].ID, resp[0].ID)
		assert.Len(t, resp[0].DuplicateTransactions, 1)
		assert.Equal(t, txList[1].ID, resp[0].DuplicateTransactions[0].ID)

		assert.Equal(t, txList[2].ID, resp[1].ID)

		assert.Equal(t, "USD", resp[0].DestinationCurrency)
		assert.Equal(t, "UAH", resp[0].SourceCurrency)

		assert.Equal(t, "1.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "38.00", resp[0].SourceAmount.StringFixed(2))
	})

	t.Run("test currency exchange", func(t *testing.T) {
		input1 := `123.54EUR Переказ на свою карту 52**20 через Приват24
5*60 22:02
Бал. 0.00EUR`

		input2 := `5555.99UAH Переказ зі своєї картки 52**60 через Приват24
5*20 22:02
Бал. 123.28UAH
Кред. лiмiт 300000.0UAH`

		srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(input1),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(input2),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "5555.99", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "UAH", resp[0].DestinationCurrency)

		assert.Equal(t, "123.54", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "EUR", resp[0].SourceCurrency)

		assert.Equal(t, "Переказ на свою карту 52**20 через Приват24", resp[0].Description)
		assert.Equal(t, "5*60", resp[0].SourceAccount)
		assert.Equal(t, "5*20", resp[0].DestinationAccount)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})

	t.Run("test transfer between account same currency", func(t *testing.T) {
		input1 := `100.00USD Переказ на свою карту 47**71 через Приват24
4*67 11:19
Бал. 123.00USD`

		input2 := `100.00USD Переказ зі своєї картки 46**67 через Приват24
4*71 11:20
Бал. 178.65USD`

		srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

		resp, err := srv.ParseMessages(context.TODO(), []*importers.Record{
			{
				Data: []byte(input1),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
			{
				Data: []byte(input2),
				Message: &importers.Message{
					CreatedAt: time.Now(),
				},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp, 1)

		assert.Equal(t, "100.00", resp[0].DestinationAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].DestinationCurrency)

		assert.Equal(t, "100.00", resp[0].SourceAmount.StringFixed(2))
		assert.Equal(t, "USD", resp[0].SourceCurrency)

		assert.Equal(t, "Переказ на свою карту 47**71 через Приват24", resp[0].Description)
		assert.Equal(t, "4*67", resp[0].SourceAccount)
		assert.Equal(t, "4*71", resp[0].DestinationAccount)
		assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	})
}

func TestImportExpense(t *testing.T) {
	p := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	usdAccount := &database.Account{
		ID:            1,
		Currency:      "USD",
		AccountNumber: "4*67",
		Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
	}
	plnAccount := &database.Account{
		ID:            2,
		Currency:      "PLN",
		Type:          v1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Name:          "_default_expense",
		Flags:         database.AccountFlagIsDefault,
		AccountNumber: "_default_expense_pln",
	}

	data := []byte(`PrivatBank, [10/1/2025 9:50 AM]
89.80PLN Ресторани, кафе, бари. Pyszne.pl, Wroclaw
4*67 16:17
Бал. 1.86USD
Курс 0.2547 USD/PLN`)

	_, err := p.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{string(data)},
			Accounts: []*database.Account{usdAccount, plnAccount},
		},
	})

	assert.NoError(t, err)
}

func TestImportInternalTransfer(t *testing.T) {
	p := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	account1 := &database.Account{
		ID:            1,
		Currency:      "USD",
		AccountNumber: "4*59",
	}
	account2 := &database.Account{
		ID:            2,
		Currency:      "USD",
		AccountNumber: "4*71",
	}

	data := []byte(`PrivatBank, [10/1/2025 9:50 AM]
1100.00USD Переказ зі своєї карти
4*59 22:40
Бал. 1.21USD

PrivatBank, [10/1/2025 9:50 AM]
1100.00USD Зарахування переказу на картку
4*71 22:40
Бал. 1.60USD`)

	_, err := p.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{string(data)},
			Accounts: []*database.Account{account1, account2},
		},
	})

	assert.NoError(t, err)
}

func TestImportRemoteTransfer(t *testing.T) {
	p := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	usdAccount := &database.Account{
		ID:            1,
		Currency:      "USD",
		AccountNumber: "4*67",
	}
	uahAccount := &database.Account{
		ID:            2,
		Currency:      "UAH",
		Type:          v1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Flags:         database.AccountFlagIsDefault,
		AccountNumber: "_default_expense_uah",
	}

	data := []byte(`PrivatBank, [10/1/2025 9:50 AM]
1000.00UAH Переказ через Приват24 зі своєї карти
Отримувач: Іван Іваненко
4*67 12:30
Бал. 100.50USD
Курс 0.0263 USD/UAH`)

	_, err := p.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{string(data)},
			Accounts: []*database.Account{usdAccount, uahAccount},
		},
	})

	assert.NoError(t, err)
}

func TestImportIncomeTransfer(t *testing.T) {
	p := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	usdAccount := &database.Account{
		ID:            1,
		Currency:      "USD",
		AccountNumber: "4*67",
	}
	uahAccount := &database.Account{
		ID:            2,
		Currency:      "UAH",
		Type:          v1.AccountType_ACCOUNT_TYPE_INCOME,
		Flags:         database.AccountFlagIsDefault,
		AccountNumber: "_default_income_uah",
	}

	data := []byte(`PrivatBank, [10/1/2025 9:50 AM]
5000.00UAH Зарахування переказу
Відправник: ТОВ "Компанія"
4*67 14:22
Бал. 1050.00USD`)

	_, err := p.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{string(data)},
			Accounts: []*database.Account{usdAccount, uahAccount},
		},
	})

	assert.NoError(t, err)
}

func TestToDbTransactionsIncome(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mockConverter.EXPECT().Convert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(decimal.NewFromInt(1), nil).AnyTimes()

	p := importers.NewPrivat24(importers.NewBaseParser(mockConverter, nil, nil))

	uahAccount := &database.Account{
		ID:            1,
		Currency:      "UAH",
		AccountNumber: "5*20",
		Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
	}
	incomeAccount := &database.Account{
		ID:            2,
		Currency:      "UAH",
		Type:          v1.AccountType_ACCOUNT_TYPE_INCOME,
		Flags:         database.AccountFlagIsDefault,
		AccountNumber: "_default_income_uah",
	}

	data := []byte(`PrivatBank, [10/1/2025 9:50 AM]
123.11UAH Переказ через Приват24 Відправник: Імя Фамілія ПоБатькові
5*20 20:11
Бал. 11111.22UAH`)

	result, err := p.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{string(data)},
			Accounts: []*database.Account{uahAccount, incomeAccount},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.CreateRequests, 1)

	req := result.CreateRequests[0]
	assert.Equal(t, "Переказ через Приват24 Відправник: Імя Фамілія ПоБатькові", req.Title)
	income := req.Transaction.(*transactionsv1.CreateTransactionRequest_Income)
	assert.EqualValues(t, "123.11", income.Income.DestinationAmount)
	assert.EqualValues(t, "UAH", income.Income.DestinationCurrency)
	assert.EqualValues(t, uahAccount.ID, income.Income.DestinationAccountId)
	assert.EqualValues(t, incomeAccount.ID, income.Income.SourceAccountId)
}

func TestToDbTransactionsRemoteTransfer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConverter := NewMockCurrencyConverterSvc(ctrl)
	mockConverter.EXPECT().Convert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(decimal.NewFromInt(1), nil).AnyTimes()

	p := importers.NewPrivat24(importers.NewBaseParser(mockConverter, nil, nil))

	uahAccount := &database.Account{
		ID:            1,
		Currency:      "UAH",
		AccountNumber: "4*68",
		Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
	}
	expenseAccount := &database.Account{
		ID:            2,
		Currency:      "UAH",
		Type:          v1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Flags:         database.AccountFlagIsDefault,
		AccountNumber: "_default_expense_uah",
	}

	data := []byte(`PrivatBank, [10/1/2025 9:50 AM]
1.00UAH Переказ через Приват24 Одержувач: Імя Фамілія ПоБатькові
4*68 16:41
Бал. 17.81UAH`)

	result, err := p.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{string(data)},
			Accounts: []*database.Account{uahAccount, expenseAccount},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.CreateRequests, 1)

	req := result.CreateRequests[0]
	assert.Equal(t, "Переказ через Приват24 Одержувач: Імя Фамілія ПоБатькові", req.Title)
	expense := req.Transaction.(*transactionsv1.CreateTransactionRequest_Expense)
	assert.EqualValues(t, "-1", expense.Expense.SourceAmount)
	assert.EqualValues(t, "UAH", expense.Expense.SourceCurrency)
	assert.EqualValues(t, uahAccount.ID, expense.Expense.SourceAccountId)
	assert.EqualValues(t, expenseAccount.ID, expense.Expense.DestinationAccountId)
}
