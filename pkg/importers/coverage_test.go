package importers_test

import (
	"context"
	"testing"
	"time"

	"github.com/ft-t/go-money/pkg/importers"
	"github.com/stretchr/testify/assert"
)

func TestPrivat24ParseCreditPaymentErrors(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("insufficient lines", func(t *testing.T) {
		input := `100.00UAH Test`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.Contains(t, err.Error(), "expected")
	})

	t.Run("invalid amount format", func(t *testing.T) {
		input := `invalidUAH Списання
4*40 20.05.24`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("insufficient source parts", func(t *testing.T) {
		input := `100.00UAH Списання
invalid`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("invalid exchange rate parts", func(t *testing.T) {
		input := `100.00UAH Списання
4*40 20.05.24
Курс invalid`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("invalid currency pair", func(t *testing.T) {
		input := `100.00UAH Списання
4*40 20.05.24
Курс 38.50 USD`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("invalid exchange rate value", func(t *testing.T) {
		input := `100.00UAH Списання
4*40 20.05.24
Курс invalid UAH/USD`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("currency mismatch in exchange", func(t *testing.T) {
		input := `100.00UAH Списання
4*40 20.05.24
Курс 38.50 EUR/USD`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.Contains(t, err.Error(), "currency mismatch")
	})

	t.Run("valid with exchange rate UAH/USD", func(t *testing.T) {
		input := `100.00UAH Списання
4*40 20.05.24
Курс 38.50 UAH/USD`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, "UAH", tx.DestinationCurrency)
		assert.Equal(t, "USD", tx.SourceCurrency)
	})

	t.Run("valid with exchange rate USD/UAH", func(t *testing.T) {
		input := `100.00UAH Списання
4*40 20.05.24
Курс 38.50 USD/UAH`
		tx, err := srv.ParseCreditPayment(context.TODO(), input, time.Now())
		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, "UAH", tx.DestinationCurrency)
		assert.Equal(t, "USD", tx.SourceCurrency)
	})
}

func TestPrivat24ParseIncomingCardTransferErrors(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("invalid amount", func(t *testing.T) {
		input := `invalidUAH Переказ з картки
Від: John Doe, картка *1234
*5678 15.05.24 12:00`
		tx, err := srv.ParseIncomingCardTransfer(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("invalid time", func(t *testing.T) {
		input := `100.00UAH Переказ з картки
Від: John Doe, картка *1234
*5678 invalid-time`
		tx, err := srv.ParseIncomingCardTransfer(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}

func TestPrivat24ParsePartialRefundErrors(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("invalid amount", func(t *testing.T) {
		input := `invalidUAH Зарахування
4*71 09.03.24`
		tx, err := srv.ParsePartialRefund(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("insufficient lines", func(t *testing.T) {
		input := `100.00UAH Зарахування`
		tx, err := srv.ParsePartialRefund(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}

func TestPrivat24ParseIncomeTransferErrors(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("invalid amount", func(t *testing.T) {
		input := `invalidUAH Переказ
Від: John Doe
*5678 15.05.24 12:00`
		tx, err := srv.ParseIncomeTransfer(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("invalid time", func(t *testing.T) {
		input := `100.00UAH Переказ
Від: John Doe
*5678 invalid-time`
		tx, err := srv.ParseIncomeTransfer(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}

func TestPrivat24ParseRemoteTransferErrors(t *testing.T) {
	srv := importers.NewPrivat24(importers.NewBaseParser(nil, nil, nil))

	t.Run("invalid amount", func(t *testing.T) {
		input := `invalidUAH Переказ на картку
Одержувач: John Doe *1234
*5678 15.05.24 12:00`
		tx, err := srv.ParseRemoteTransfer(context.TODO(), input, time.Now())
		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}
