package main

import (
	"context"
	accountsv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/accounts/v1"
	configurationv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/configuration/v1"
	currencyv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/currency/v1"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
	"github.com/shopspring/decimal"
)

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
