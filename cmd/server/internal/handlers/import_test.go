package handlers_test

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
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

func TestNewImportApi(t *testing.T) {
	mockSvc := NewMockImportSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, err := handlers.NewImportApi(grpc, mockSvc)
	assert.NoError(t, err)
	assert.NotNil(t, api)
}

func TestImportApi_ImportTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockImportSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewImportApi(grpc, mockSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.ImportTransactionsRequest{})
		respMsg := &importv1.ImportTransactionsResponse{}
		mockSvc.EXPECT().Import(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ImportTransactions(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.ImportTransactionsRequest{})
		mockSvc.EXPECT().Import(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ImportTransactions(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&importv1.ImportTransactionsRequest{})
		resp, err := api.ImportTransactions(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
