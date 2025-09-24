package handlers_test

import (
	"context"
	"net/http"
	"testing"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions/applicable_accounts"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTransactionApi_ListTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockTransactionsSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewTransactionApi(grpc, mockSvc, nil, nil)

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
	api := handlers.NewTransactionApi(grpc, mockSvc, nil, nil)

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

func TestTransactionApi_UpdateTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockTransactionsSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewTransactionApi(grpc, mockSvc, nil, nil)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.UpdateTransactionRequest{})
		respMsg := &transactionsv1.UpdateTransactionResponse{}
		mockSvc.EXPECT().Update(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.UpdateTransaction(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.UpdateTransactionRequest{})
		mockSvc.EXPECT().Update(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.UpdateTransaction(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&transactionsv1.UpdateTransactionRequest{})
		resp, err := api.UpdateTransaction(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestTransactionApi_GetApplicableAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockApplicableSvc := NewMockApplicableAccountSvc(ctrl)
	mockMapper := NewMockMapperSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewTransactionApi(grpc, nil, mockApplicableSvc, mockMapper)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.GetApplicableAccountsRequest{})

		mockResp := map[gomoneypbv1.TransactionType]*applicable_accounts.PossibleAccount{
			gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME: {
				SourceAccounts: map[int32]*database.Account{
					1: {
						ID: 1,
					},
				},
				DestinationAccounts: map[int32]*database.Account{
					2: {
						ID: 2,
					},
				},
			},
		}
		mockApplicableSvc.EXPECT().GetAll(gomock.Any()).Return(mockResp, nil)
		mockMapper.EXPECT().MapAccount(gomock.Any(), gomock.Any()).Return(&gomoneypbv1.Account{Id: 1}).Times(2)

		resp, err := api.GetApplicableAccounts(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, resp.Msg.ApplicableRecords, 1)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.GetApplicableAccountsRequest{})
		mockApplicableSvc.EXPECT().GetAll(gomock.Any()).Return(nil, assert.AnError)
		resp, err := api.GetApplicableAccounts(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&transactionsv1.GetApplicableAccountsRequest{})
		resp, err := api.GetApplicableAccounts(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestTransactionApi_GetTitleSuggestions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockTransactionsSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewTransactionApi(grpc, mockSvc, nil, nil)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.GetTitleSuggestionsRequest{
			Query: "coffee",
			Limit: 10,
		})
		respMsg := &transactionsv1.GetTitleSuggestionsResponse{
			Titles: []string{"Coffee Shop", "Coffee Bean Store"},
		}
		mockSvc.EXPECT().GetTitleSuggestions(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.GetTitleSuggestions(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
		assert.Len(t, resp.Msg.Titles, 2)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&transactionsv1.GetTitleSuggestionsRequest{
			Query: "coffee",
		})
		mockSvc.EXPECT().GetTitleSuggestions(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.GetTitleSuggestions(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&transactionsv1.GetTitleSuggestionsRequest{
			Query: "coffee",
		})
		resp, err := api.GetTitleSuggestions(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
