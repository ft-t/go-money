package handlers

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/currency/v1/currencyv1connect"
	"buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/currency/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
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
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

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
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.currencySvc.GetCurrencies(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *CurrencyApi) DeleteCurrency(ctx context.Context, c *connect.Request[currencyv1.DeleteCurrencyRequest]) (*connect.Response[currencyv1.DeleteCurrencyResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.currencySvc.DeleteCurrency(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *CurrencyApi) CreateCurrency(
	ctx context.Context,
	c *connect.Request[currencyv1.CreateCurrencyRequest],
) (*connect.Response[currencyv1.CreateCurrencyResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

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
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.currencySvc.UpdateCurrency(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}
