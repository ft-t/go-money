package main

import (
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money-pb/gen/gomoneypb/currency/v1"
	"github.com/ft-t/go-money-pb/gen/gomoneypb/currency/v1/currencyv1connect"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/shopspring/decimal"
)

type CurrencyApi struct {
	currencySvc  CurrencySvc
	converterSvc ConverterSvc
	decimalSvc   DecimalSvc
}

func NewCurrencyApi(
	mux *boilerplate.DefaultGrpcServer,
	currencySvc CurrencySvc,
	converterSvc ConverterSvc,
	decimalSvc DecimalSvc,
) (*CurrencyApi, error) {
	res := &CurrencyApi{
		currencySvc:  currencySvc,
		converterSvc: converterSvc,
		decimalSvc:   decimalSvc,
	}

	mux.GetMux().Handle(
		currencyv1connect.NewCurrencyServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res, nil
}

func (a *CurrencyApi) Exchange(
	ctx context.Context,
	c *connect.Request[currencyv1.ExchangeRequest],
) (*connect.Response[currencyv1.ExchangeResponse], error) {
	amount, err := decimal.NewFromString(c.Msg.Amount)
	if err != nil {
		return nil, err
	}

	resp, err := a.converterSvc.Convert(ctx, c.Msg.FromCurrency, c.Msg.ToCurrency, amount)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&currencyv1.ExchangeResponse{
		Amount: a.decimalSvc.ToString(ctx, resp, c.Msg.ToCurrency),
	}), nil
}

func (a *CurrencyApi) GetCurrencies(
	ctx context.Context,
	c *connect.Request[currencyv1.GetCurrenciesRequest],
) (*connect.Response[currencyv1.GetCurrenciesResponse], error) {
	resp, err := a.currencySvc.GetCurrencies(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *CurrencyApi) DeleteCurrency(ctx context.Context, c *connect.Request[currencyv1.DeleteCurrencyRequest]) (*connect.Response[currencyv1.DeleteCurrencyResponse], error) {
	// todo auth
	panic("implement me")
}

func (a *CurrencyApi) CreateCurrency(
	ctx context.Context,
	c *connect.Request[currencyv1.CreateCurrencyRequest],
) (*connect.Response[currencyv1.CreateCurrencyResponse], error) {
	// todo auth

	resp, err := a.currencySvc.CreateCurrency(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *CurrencyApi) UpdateCurrency(
	ctx context.Context,
	c *connect.Request[currencyv1.UpdateCurrencyRequest],
) (*connect.Response[currencyv1.UpdateCurrencyResponse], error) {
	// todo auth

	resp, err := a.currencySvc.UpdateCurrency(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}
