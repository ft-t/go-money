package importers_test

import (
	"context"
	"encoding/base64"
	_ "embed"
	"testing"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/blik.xlsx
var blik []byte

//go:embed testdata/blik_v2.xlsx
var blikV2 []byte

//go:embed testdata/two_expenses.xlsx
var twoExpenses []byte

//go:embed testdata/two_expenses_v2.xlsx
var twoExpensesV2 []byte

//go:embed testdata/to_phone.xlsx
var toPhone []byte

//go:embed testdata/to_phone_v2.xlsx
var toPhoneV2 []byte

//go:embed testdata/blokada_srodkow.xlsx
var blokada []byte

//go:embed testdata/blokada_srodkow_v2.xlsx
var blokadaV2 []byte

//go:embed testdata/income.xlsx
var income []byte

//go:embed testdata/income_v2.xlsx
var incomeV2 []byte

//go:embed testdata/transfer_to_private_acc.xlsx
var transferToPrivateAccount []byte

//go:embed testdata/transfer_to_private_acc_v2.xlsx
var transferToPrivateAccountV2 []byte

//go:embed testdata/credit_card.xlsx
var creditCardPayment []byte

//go:embed testdata/credit_card_v2.xlsx
var creditCardPaymentV2 []byte

//go:embed testdata/currency_exchange.xlsx
var currencyExchange []byte

//go:embed testdata/currency_exchange_v2.xlsx
var currencyExchangeV2 []byte

//go:embed testdata/transfer_betwee_accounts.xlsx
var betweenAccounts []byte

//go:embed testdata/transfer_between_accounts_v2.xlsx
var betweenAccountsV2 []byte

//go:embed testdata/currency_exchange2.xlsx
var currencyExchange2 []byte

//go:embed testdata/currency_exchange2_2.xlsx
var currencyExchange22 []byte

//go:embed testdata/outgoing_payment_multi_currency.xlsx
var outgoingPaymentMultiCurrency []byte

//go:embed testdata/pshelev_expense.xlsx
var pshelevExpense []byte

//go:embed testdata/account_commission.xlsx
var accountCommission []byte

//go:embed testdata/account_commission_v2.xlsx
var accountCommissionV2 []byte

//go:embed testdata/cash_withdrawal.xlsx
var cashWithdrawal []byte

//go:embed testdata/cash_withdrawal_v2.xlsx
var cashWithdrawalV2 []byte

//go:embed testdata/similar_transfers.xlsx
var similarTransfers []byte

//go:embed testdata/similar_transfers_v2.xlsx
var similarTransfersV2 []byte

//go:embed testdata/blik_refund.xlsx
var blikRefund []byte

//go:embed testdata/blik_refund_v2.xlsx
var blikRefundV2 []byte

//go:embed testdata/inne_withdrawal.xlsx
var inneWithdrawal []byte

//go:embed testdata/inne_withdrawal_v2.xlsx
var inneWithdrawalV2 []byte

//go:embed testdata/income_multicurrency.xlsx
var incomeMultiCurrency []byte

func TestInneWithdrawal_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 inne withdrawal",
			data:        inneWithdrawal,
		},
		{
			description: "v2 inne withdrawal",
			data:        inneWithdrawalV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
			assert.Equal(t, "8.63", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "111111111111111111111", resp[0].SourceAccount)

			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-04-04 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Name Surname", resp[0].Description)
		})
	}
}

func TestInneWithdrawal_Failure(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
		errMsg      string
	}{
		{
			description: "empty data",
			data:        []byte{},
			errMsg:      "failed to open excel",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			require.Error(t, resp[0].ParsingError)
			assert.Contains(t, resp[0].ParsingError.Error(), testCase.errMsg)
		})
	}
}

func TestBlikRefund_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 blik refund",
			data:        blikRefund,
		},
		{
			description: "v2 blik refund",
			data:        blikRefundV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)
			assert.Equal(t, "2699.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].DestinationCurrency)
			assert.Equal(t, "2222222222222222222222", resp[0].DestinationAccount)

			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-04-05 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Transakcja BLIK, Zwrot dla transakcji TR-U4J-BVPHTYX, Zwrot BLIK internet, Nr 1234566, TERG SPÓŁKA AKCYJNA, REF-12345", resp[0].Description)
		})
	}
}

func TestCashWithdrawal_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 cash withdrawal",
			data:        cashWithdrawal,
		},
		{
			description: "v2 cash withdrawal",
			data:        cashWithdrawalV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
			assert.Equal(t, "600.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "11111111111111111111111111", resp[0].SourceAccount)

			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-03-02 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "1111------222 XX YY zz 4321321 POL 600,00 PLN 2024-03-02", resp[0].Description)
		})
	}
}

func TestSimilarTransfers_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 similar transfer",
			data:        similarTransfers,
		},
		{
			description: "v2 similar transfer",
			data:        similarTransfersV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 2)

			assert.Equal(t, importers.TransactionTypeRemoteTransfer, resp[0].Type)
			assert.Equal(t, "11.68", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "22222222222222222222222222", resp[0].SourceAccount)

			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-03-01 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Common Desription", resp[0].Description)

			assert.Equal(t, "22.16", resp[1].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[1].SourceCurrency)
		})
	}
}

func TestExpenseCommission_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 account commission",
			data:        accountCommission,
		},
		{
			description: "v2 account commission",
			data:        accountCommissionV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
			assert.Equal(t, "2.31", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "EUR", resp[0].SourceCurrency)
			assert.Equal(t, "11111111111111111111111111111111111", resp[0].SourceAccount)

			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-02-24 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Prowizje i opłaty", resp[0].Description)
		})
	}
}

func TestPshelevExpense_Success(t *testing.T) {
	srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
		{
			Data: pshelevExpense,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp, 1)

	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
	assert.Equal(t, "200.00", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].SourceCurrency)
	assert.Equal(t, "111111111111111111111111", resp[0].SourceAccount)

	assert.Equal(t, "200.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].DestinationCurrency)
	assert.Equal(t, "2222222222222222222222", resp[0].DestinationAccount)

	assert.Equal(t, "00:00", resp[0].DateFromMessage)
	assert.Equal(t, "2024-02-08 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
	assert.Equal(t, "333333", resp[0].Description)
}

func TestToPhone_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 to phone",
			data:        toPhone,
		},
		{
			description: "v2 to phone",
			data:        toPhoneV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeRemoteTransfer, resp[0].Type)
			assert.Equal(t, "200.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "11111111111111111111111", resp[0].SourceAccount)

			assert.Equal(t, "200.00", resp[0].DestinationAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].DestinationCurrency)
			assert.Equal(t, "22222222222222222222222", resp[0].DestinationAccount)

			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-02-07 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Some Descriptions", resp[0].Description)
		})
	}
}

func TestParibasBlik_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 blik",
			data:        blik,
		},
		{
			description: "v2 blik",
			data:        blikV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
			assert.Equal(t, "119.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "11112222333344455556777", resp[0].SourceAccount)
			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-02-02 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Transakcja BLIK, Allegro xxxx-c21, Płatność BLIK w internecie, Nr 12324, ALLEGRO SP. Z O.O., allegro.pl", resp[0].Description)
		})
	}
}

func TestTwoExpenses_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 two expenses",
			data:        twoExpenses,
		},
		{
			description: "v2 two expenses",
			data:        twoExpensesV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 2)

			assert.NotEqual(t, resp[0].ID, resp[1].ID)
			assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
			assert.Equal(t, "500.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "USD", resp[0].SourceCurrency)
		})
	}
}

func TestParibasBlokadaSrodkow_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 blokada",
			data:        blokada,
		},
		{
			description: "v2 blokada",
			data:        blokadaV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {

			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)
			assert.Equal(t, "500.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "USD", resp[0].SourceCurrency)
			assert.Equal(t, "1234567", resp[0].SourceAccount)
			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-02-08 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "PAYPAL  XTB S 111 PL 111______111 500,00 USD ", resp[0].Description)
			assert.ErrorContains(t, resp[0].ParsingError, "transaction is still pending. will skip from firefly for now")
		})
	}
}

func TestMultiCurrencyPayment_Success(t *testing.T) {
	srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
		{
			Data: outgoingPaymentMultiCurrency,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp, 1)

	assert.Equal(t, importers.TransactionTypeExpense, resp[0].Type)

	assert.Equal(t, "837.89", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "PLN", resp[0].SourceCurrency)
	assert.Equal(t, "22222222222222222222222", resp[0].SourceAccount)

	assert.Equal(t, "199.00", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "USD", resp[0].DestinationCurrency)

	assert.Equal(t, "00:00", resp[0].DateFromMessage)
	assert.Equal(t, "2024-01-08 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
	assert.Equal(t, "xx-----yy 4,xx xx.yyy my.vmware.com DRI-VMware IRL 199,00 USD 2024-01-08", resp[0].Description)
}

func TestParibasIncome_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 income",
			data:        income,
		},
		{
			description: "v2 income",
			data:        incomeV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {

			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			require.NoError(t, resp[0].ParsingError)
			assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)
			assert.Equal(t, "11.48", resp[0].DestinationAmount.StringFixed(2))
			assert.Equal(t, "EUR", resp[0].DestinationCurrency)
			assert.Equal(t, "123443252341234214321331", resp[0].DestinationAccount)

			assert.Equal(t, "11.48", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "EUR", resp[0].SourceCurrency)
			assert.Equal(t, "/es123432523424213132", resp[0].SourceAccount)

			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-02-01 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "SOFTWARE DEVELOPMENT SERVICES, INVOICE NO 1-2 XXYY, 31.01.2024", resp[0].Description)
		})
	}
}

func TestParibasIncomeMultiCurrency_Success(t *testing.T) {
	srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
		{
			Data: incomeMultiCurrency,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp, 1)

	require.NoError(t, resp[0].ParsingError)
	assert.Equal(t, importers.TransactionTypeIncome, resp[0].Type)
	assert.Equal(t, "102.16", resp[0].DestinationAmount.StringFixed(2))
	assert.Equal(t, "EUR", resp[0].DestinationCurrency)
	assert.Equal(t, "22222222222222222222222222", resp[0].DestinationAccount)

	assert.Equal(t, "500.00", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "PLN", resp[0].SourceCurrency)
	assert.Equal(t, "11111111111111111111111111", resp[0].SourceAccount)

	assert.Equal(t, "00:00", resp[0].DateFromMessage)
	assert.Equal(t, "2024-07-09 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
	assert.Equal(t, "ID 4444444 5555555555555555555", resp[0].Description)
}

func TestParibasTransferToPrivateAccount_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 transfer to private account",
			data:        transferToPrivateAccount,
		},
		{
			description: "v2 transfer to private account",
			data:        transferToPrivateAccountV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {

			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeRemoteTransfer, resp[0].Type)
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "1200.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].DestinationCurrency)
			assert.Equal(t, "1200.00", resp[0].DestinationAmount.StringFixed(2))

			assert.Equal(t, "22222222222222222222222222222222", resp[0].SourceAccount)
			assert.Equal(t, "1111111111111111111111111111", resp[0].DestinationAccount)
			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-02-01 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Przelew środków", resp[0].Description)
		})
	}
}

func TestParibasCreditCardRepayment_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 credit card repayment",
			data:        creditCardPayment,
		},
		{
			description: "v2 credit card repayment",
			data:        creditCardPaymentV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "1.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].DestinationCurrency)
			assert.Equal(t, "1.00", resp[0].DestinationAmount.StringFixed(2))

			assert.Equal(t, "22222222222222222222222222", resp[0].SourceAccount)
			assert.Equal(t, "11111111111111111111111111", resp[0].DestinationAccount)
			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-10-20 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Spłata karty", resp[0].Description)
		})
	}
}

func TestParibasCurrencyExchange_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 currency exchange",
			data:        currencyExchange,
		},
		{
			description: "v2 currency exchange",
			data:        currencyExchangeV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {

			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)

			assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
			assert.Equal(t, "USD", resp[0].SourceCurrency)
			assert.Equal(t, "1500.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].DestinationCurrency)
			assert.Equal(t, "6000.90", resp[0].DestinationAmount.StringFixed(2))

			assert.Equal(t, "1111111111111", resp[0].SourceAccount)
			assert.Equal(t, "22222222222222222", resp[0].DestinationAccount)
			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-01-24 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "USD PLN 4.0006 TWM2131232132131", resp[0].Description)
		})
	}
}

func TestParibasCurrencyExchange2_Success(t *testing.T) {
	srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
		{
			Data: currencyExchange2,
		},
		{
			Data: currencyExchange22,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp, 1)

	assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
	assert.Equal(t, "EUR", resp[0].SourceCurrency)
	assert.Equal(t, "1390.56", resp[0].SourceAmount.StringFixed(2))
	assert.Equal(t, "PLN", resp[0].DestinationCurrency)
	assert.Equal(t, "6000.00", resp[0].DestinationAmount.StringFixed(2))

	assert.Equal(t, "22222222222222222222222222", resp[0].SourceAccount)
	assert.Equal(t, "11111111111111111111111111", resp[0].DestinationAccount)
	assert.Equal(t, "00:00", resp[0].DateFromMessage)
	assert.Equal(t, "2024-02-08 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
	assert.Equal(t, "EUR PLN 4.3148 TWM2402064671", resp[0].Description)
}

func TestParibasBetweenAccounts_Success(t *testing.T) {
	testCases := []struct {
		description string
		data        []byte
	}{
		{
			description: "v1 two expenses",
			data:        betweenAccounts,
		},
		{
			description: "v2 two expenses",
			data:        betweenAccountsV2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {

			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

			resp, err := srv.ParseMessages(context.Background(), []*importers.Record{
				{
					Data: testCase.data,
				},
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.Len(t, resp, 1)
			require.Len(t, resp[0].DuplicateTransactions, 1)

			assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].DuplicateTransactions[0].Type)
			assert.Equal(t, importers.TransactionTypeInternalTransfer, resp[0].Type)
			assert.Equal(t, "PLN", resp[0].SourceCurrency)
			assert.Equal(t, "1200.00", resp[0].SourceAmount.StringFixed(2))
			assert.Equal(t, "PLN", resp[0].DestinationCurrency)
			assert.Equal(t, "1200.00", resp[0].DestinationAmount.StringFixed(2))

			assert.Equal(t, "11111111111111111111111111", resp[0].DestinationAccount)
			assert.Equal(t, "22222222222222222222222222", resp[0].SourceAccount)
			assert.Equal(t, "00:00", resp[0].DateFromMessage)
			assert.Equal(t, "2024-02-01 00:00:00 +0000", resp[0].Date.Format("2006-01-02 15:04:05 -0700"))
			assert.Equal(t, "Przelew środków", resp[0].Description)
		})
	}
}

func TestParibasParseMessages_NoSheets(t *testing.T) {
	srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

	records := []*importers.Record{
		{
			Data: []byte("corrupted excel data"),
		},
	}

	parsedRecords, err := srv.ParseMessages(context.Background(), records)
	require.NoError(t, err)
	require.Len(t, parsedRecords, 1)
	assert.Error(t, parsedRecords[0].ParsingError)
	assert.Contains(t, parsedRecords[0].ParsingError.Error(), "failed to open excel")
}

func TestParibasType(t *testing.T) {
	srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

	assert.Equal(t, importv1.ImportSource_IMPORT_SOURCE_BNP_PARIBAS_POLSKA, srv.Type())
}

func TestParibasParse_InvalidBase64(t *testing.T) {
	srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{"invalid-base64!!!"},
			Accounts: nil,
		},
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode file content")
	assert.Nil(t, resp)
}

func TestParibasParse_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	base := importers.NewBaseParser(currencyConverter, txSvc, mapperSvc)
	srv := importers.NewParibas(base)

	account1 := &database.Account{
		ID:            1,
		Name:          "Test Account",
		Currency:      "PLN",
		AccountNumber: "11112222333344455556777",
		Type:          gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	}

	expenseAccount := &database.Account{
		ID:       2,
		Name:     "Default Expense",
		Currency: "EUR",
		Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
		Flags:    database.AccountFlagIsDefault,
	}

	accounts := []*database.Account{account1, expenseAccount}

	currencyConverter.EXPECT().
		Convert(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, from string, to string, amount decimal.Decimal) (decimal.Decimal, error) {
			assert.Equal(t, "PLN", from)
			assert.Equal(t, "EUR", to)
			return amount, nil
		}).
		Times(1)

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(blik)},
			Accounts: accounts,
		},
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.CreateRequests, 1)
}

func TestParibasParse_ParseMessagesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	base := importers.NewBaseParser(currencyConverter, txSvc, mapperSvc)
	srv := importers.NewParibas(base)

	invalidData := []byte("not excel data")

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(invalidData)},
			Accounts: []*database.Account{},
		},
	})

	// ParseMessages returns transactions with errors, not an error itself
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCanMatchAsInternalTransfer_Success(t *testing.T) {
	testCases := []struct {
		name string
		f    *importers.Transaction
		tx   *importers.Transaction
	}{
		{
			name: "income and remote transfer match",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "remote transfer and income match",
			f:    &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "empty source accounts allowed",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "empty destination accounts allowed",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: ""},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "both source accounts empty",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "", DestinationAccount: "B"},
		},
		{
			name: "both destination accounts empty",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: ""},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))
			result := srv.CanMatchAsInternalTransfer(tc.f, tc.tx)
			assert.True(t, result)
		})
	}
}

func TestCanMatchAsInternalTransfer_Failure(t *testing.T) {
	testCases := []struct {
		name string
		f    *importers.Transaction
		tx   *importers.Transaction
	}{
		{
			name: "both income type",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "both remote transfer type",
			f:    &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "both expense type",
			f:    &importers.Transaction{Type: importers.TransactionTypeExpense, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeExpense, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "expense and income",
			f:    &importers.Transaction{Type: importers.TransactionTypeExpense, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "internal transfer types",
			f:    &importers.Transaction{Type: importers.TransactionTypeInternalTransfer, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
		},
		{
			name: "different source accounts",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "C", DestinationAccount: "B"},
		},
		{
			name: "different destination accounts",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "A", DestinationAccount: "D"},
		},
		{
			name: "different source and destination accounts",
			f:    &importers.Transaction{Type: importers.TransactionTypeIncome, SourceAccount: "A", DestinationAccount: "B"},
			tx:   &importers.Transaction{Type: importers.TransactionTypeRemoteTransfer, SourceAccount: "C", DestinationAccount: "D"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srv := importers.NewParibas(importers.NewBaseParser(nil, nil, nil))
			result := srv.CanMatchAsInternalTransfer(tc.f, tc.tx)
			assert.False(t, result)
		})
	}
}

func TestParibasParse_GetAccountMapError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	currencyConverter := NewMockCurrencyConverterSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	base := importers.NewBaseParser(currencyConverter, txSvc, mapperSvc)
	srv := importers.NewParibas(base)

	account1 := &database.Account{
		ID:            1,
		AccountNumber: "duplicate",
	}
	account2 := &database.Account{
		ID:            2,
		AccountNumber: "duplicate",
	}

	resp, err := srv.Parse(context.TODO(), &importers.ParseRequest{
		ImportRequest: importers.ImportRequest{
			Data:     []string{base64.StdEncoding.EncodeToString(blik)},
			Accounts: []*database.Account{account1, account2},
		},
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get account map by numbers")
}
