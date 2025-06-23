package handlers_test

import (
	currencyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/currency/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCurrencyApi_Exchange(t *testing.T) {
	ctrl := gomock.NewController(t)
	currencySvc := NewMockCurrencySvc(ctrl)
	converterSvc := NewMockConverterSvc(ctrl)
	decimalSvc := NewMockDecimalSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewCurrencyApi(grpc, currencySvc, converterSvc, decimalSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.ExchangeRequest{
			FromCurrency: "USD",
			ToCurrency:   "EUR",
			Amount:       "10.50",
		})
		amount, _ := decimal.NewFromString("10.50")
		converted := decimal.NewFromFloat(9.99)
		converterSvc.EXPECT().Convert(gomock.Any(), "USD", "EUR", amount).Return(converted, nil)
		decimalSvc.EXPECT().ToString(gomock.Any(), converted, "EUR").Return("9.99")
		resp, err := api.Exchange(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "9.99", resp.Msg.Amount)
	})

	t.Run("invalid amount", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.ExchangeRequest{
			FromCurrency: "USD",
			ToCurrency:   "EUR",
			Amount:       "bad",
		})
		resp, err := api.Exchange(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("convert error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.ExchangeRequest{
			FromCurrency: "USD",
			ToCurrency:   "EUR",
			Amount:       "1.00",
		})
		amount, _ := decimal.NewFromString("1.00")
		converterSvc.EXPECT().Convert(gomock.Any(), "USD", "EUR", amount).Return(decimal.Decimal{}, assert.AnError)
		resp, err := api.Exchange(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&currencyv1.ExchangeRequest{
			FromCurrency: "USD",
			ToCurrency:   "EUR",
			Amount:       "1.00",
		})
		resp, err := api.Exchange(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestCurrencyApi_GetCurrencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	currencySvc := NewMockCurrencySvc(ctrl)
	converterSvc := NewMockConverterSvc(ctrl)
	decimalSvc := NewMockDecimalSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewCurrencyApi(grpc, currencySvc, converterSvc, decimalSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.GetCurrenciesRequest{})
		respMsg := &currencyv1.GetCurrenciesResponse{}
		currencySvc.EXPECT().GetCurrencies(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.GetCurrencies(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.GetCurrenciesRequest{})
		currencySvc.EXPECT().GetCurrencies(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.GetCurrencies(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&currencyv1.GetCurrenciesRequest{})
		resp, err := api.GetCurrencies(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestCurrencyApi_DeleteCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	currencySvc := NewMockCurrencySvc(ctrl)
	converterSvc := NewMockConverterSvc(ctrl)
	decimalSvc := NewMockDecimalSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewCurrencyApi(grpc, currencySvc, converterSvc, decimalSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.DeleteCurrencyRequest{})
		respMsg := &currencyv1.DeleteCurrencyResponse{}
		currencySvc.EXPECT().DeleteCurrency(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.DeleteCurrency(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.DeleteCurrencyRequest{})
		currencySvc.EXPECT().DeleteCurrency(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.DeleteCurrency(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&currencyv1.DeleteCurrencyRequest{})
		resp, err := api.DeleteCurrency(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestCurrencyApi_CreateCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	currencySvc := NewMockCurrencySvc(ctrl)
	converterSvc := NewMockConverterSvc(ctrl)
	decimalSvc := NewMockDecimalSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewCurrencyApi(grpc, currencySvc, converterSvc, decimalSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.CreateCurrencyRequest{})
		respMsg := &currencyv1.CreateCurrencyResponse{}
		currencySvc.EXPECT().CreateCurrency(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.CreateCurrency(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.CreateCurrencyRequest{})
		currencySvc.EXPECT().CreateCurrency(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.CreateCurrency(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&currencyv1.CreateCurrencyRequest{})
		resp, err := api.CreateCurrency(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestCurrencyApi_UpdateCurrency(t *testing.T) {
	ctrl := gomock.NewController(t)
	currencySvc := NewMockCurrencySvc(ctrl)
	converterSvc := NewMockConverterSvc(ctrl)
	decimalSvc := NewMockDecimalSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewCurrencyApi(grpc, currencySvc, converterSvc, decimalSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.UpdateCurrencyRequest{})
		respMsg := &currencyv1.UpdateCurrencyResponse{}
		currencySvc.EXPECT().UpdateCurrency(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.UpdateCurrency(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&currencyv1.UpdateCurrencyRequest{})
		currencySvc.EXPECT().UpdateCurrency(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.UpdateCurrency(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&currencyv1.UpdateCurrencyRequest{})
		resp, err := api.UpdateCurrency(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
