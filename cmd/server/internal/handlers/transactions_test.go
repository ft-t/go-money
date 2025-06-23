package handlers_test

import (
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestTransactionApi_ListTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockTransactionsSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewTransactionApi(grpc, mockSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.ListTransactionsRequest{})
		respMsg := &transactionsv1.ListTransactionsResponse{}
		mockSvc.EXPECT().List(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ListTransactions(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.ListTransactionsRequest{})
		mockSvc.EXPECT().List(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ListTransactions(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&transactionsv1.ListTransactionsRequest{})
		resp, err := api.ListTransactions(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestTransactionApi_CreateTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockTransactionsSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewTransactionApi(grpc, mockSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.CreateTransactionRequest{})
		respMsg := &transactionsv1.CreateTransactionResponse{}
		mockSvc.EXPECT().Create(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.CreateTransaction(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.CreateTransactionRequest{})
		mockSvc.EXPECT().Create(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.CreateTransaction(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&transactionsv1.CreateTransactionRequest{})
		resp, err := api.CreateTransaction(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
