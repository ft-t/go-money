package handlers

import (
	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	currencyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/currency/v1"
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	usersv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/users/v1"
	"context"
	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package handlers_test -source=interfaces.go

type TransactionsSvc interface {
	Create(
		ctx context.Context,
		req *transactionsv1.CreateTransactionRequest,
	) (*transactionsv1.CreateTransactionResponse, error)

	List(
		ctx context.Context,
		req *transactionsv1.ListTransactionsRequest,
	) (*transactionsv1.ListTransactionsResponse, error)
}

type UserSvc interface {
	Login(
		ctx context.Context,
		req *usersv1.LoginRequest,
	) (*usersv1.LoginResponse, error)

	Create(
		ctx context.Context,
		req *usersv1.CreateRequest,
	) (*usersv1.CreateResponse, error)
}

type ImportSvc interface {
	Import(ctx context.Context, req *importv1.ImportTransactionsRequest) (*importv1.ImportTransactionsResponse, error)
}

type AccountSvc interface {
	Create(
		ctx context.Context,
		req *accountsv1.CreateAccountRequest,
	) (*accountsv1.CreateAccountResponse, error)

	Update(
		ctx context.Context,
		req *accountsv1.UpdateAccountRequest,
	) (*accountsv1.UpdateAccountResponse, error)

	List(
		ctx context.Context,
		req *accountsv1.ListAccountsRequest,
	) (*accountsv1.ListAccountsResponse, error)

	Delete(
		ctx context.Context,
		req *accountsv1.DeleteAccountRequest,
	) (*accountsv1.DeleteAccountResponse, error)

	CreateBulk(
		ctx context.Context,
		req *accountsv1.CreateAccountsBulkRequest,
	) (*accountsv1.CreateAccountsBulkResponse, error)
}

type ConfigSvc interface {
	GetConfiguration(
		ctx context.Context,
		_ *configurationv1.GetConfigurationRequest,
	) (*configurationv1.GetConfigurationResponse, error)
}

type CurrencySvc interface {
	GetCurrencies(
		ctx context.Context,
		_ *currencyv1.GetCurrenciesRequest,
	) (*currencyv1.GetCurrenciesResponse, error)

	CreateCurrency(
		ctx context.Context,
		req *currencyv1.CreateCurrencyRequest,
	) (*currencyv1.CreateCurrencyResponse, error)

	UpdateCurrency(
		ctx context.Context,
		req *currencyv1.UpdateCurrencyRequest,
	) (*currencyv1.UpdateCurrencyResponse, error)

	DeleteCurrency(
		ctx context.Context,
		req *currencyv1.DeleteCurrencyRequest,
	) (*currencyv1.DeleteCurrencyResponse, error)
}

type DecimalSvc interface {
	ToString(ctx context.Context, amount decimal.Decimal, currency string) string
}

type ConverterSvc interface {
	Convert(
		ctx context.Context,
		fromCurrency string,
		toCurrency string,
		amount decimal.Decimal,
	) (decimal.Decimal, error)
}
