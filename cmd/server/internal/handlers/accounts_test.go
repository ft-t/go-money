package handlers_test

import (
	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
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

func TestReorder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		app, err := handlers.NewAccountsApi(grpc, accountSvc)
		assert.NoError(t, err)
		assert.NotNil(t, app)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{
			UserID: 123,
		})

		req := connect.NewRequest(&accountsv1.ReorderAccountsRequest{})
		assert.Panics(t, func() {
			_, _ = app.ReorderAccounts(ctx, req)
		})
	})

	t.Run("no auth", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		app, err := handlers.NewAccountsApi(grpc, accountSvc)
		assert.NoError(t, err)
		assert.NotNil(t, app)

		req := connect.NewRequest(&accountsv1.ReorderAccountsRequest{})

		resp, err := app.ReorderAccounts(context.TODO(), req)

		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestAccountsCreateBulk(t *testing.T) {
	t.Run("Create account bulk", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			accountSvc := NewMockAccountSvc(gomock.NewController(t))
			grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

			app, err := handlers.NewAccountsApi(grpc, accountSvc)
			assert.NoError(t, err)
			assert.NotNil(t, app)

			ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{
				UserID: 123,
			})

			req := connect.NewRequest(&accountsv1.CreateAccountsBulkRequest{})
			accResp := &accountsv1.CreateAccountsBulkResponse{}

			accountSvc.EXPECT().CreateBulk(gomock.Any(), req.Msg).
				Return(accResp, nil)

			resp, err := app.CreateAccountsBulk(ctx, req)

			assert.NoError(t, err)
			assert.NotNil(t, resp)

			assert.EqualValues(t, accResp, resp.Msg)
		})

		t.Run("service error", func(t *testing.T) {
			accountSvc := NewMockAccountSvc(gomock.NewController(t))
			grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

			app, err := handlers.NewAccountsApi(grpc, accountSvc)
			assert.NoError(t, err)
			assert.NotNil(t, app)

			ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{
				UserID: 123,
			})

			req := connect.NewRequest(&accountsv1.CreateAccountsBulkRequest{})

			accountSvc.EXPECT().CreateBulk(gomock.Any(), req.Msg).
				Return(nil, auth.ErrInvalidToken)

			resp, err := app.CreateAccountsBulk(ctx, req)

			assert.ErrorIs(t, err, auth.ErrInvalidToken)
			assert.Nil(t, resp)
		})

		t.Run("no auth", func(t *testing.T) {
			accountSvc := NewMockAccountSvc(gomock.NewController(t))
			grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

			app, err := handlers.NewAccountsApi(grpc, accountSvc)
			assert.NoError(t, err)
			assert.NotNil(t, app)

			req := connect.NewRequest(&accountsv1.CreateAccountsBulkRequest{})

			resp, err := app.CreateAccountsBulk(context.TODO(), req)

			assert.ErrorIs(t, err, auth.ErrInvalidToken)
			assert.Nil(t, resp)
		})
	})
}

func TestAccountsCreateAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.CreateAccountRequest{})
		respMsg := &accountsv1.CreateAccountResponse{}
		accountSvc.EXPECT().Create(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := app.CreateAccount(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})
	t.Run("service error", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.CreateAccountRequest{})
		accountSvc.EXPECT().Create(gomock.Any(), req.Msg).Return(nil, auth.ErrInvalidToken)
		resp, err := app.CreateAccount(ctx, req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
	t.Run("no auth", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		req := connect.NewRequest(&accountsv1.CreateAccountRequest{})
		resp, err := app.CreateAccount(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestAccountsUpdateAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.UpdateAccountRequest{})
		respMsg := &accountsv1.UpdateAccountResponse{}
		accountSvc.EXPECT().Update(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := app.UpdateAccount(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})
	t.Run("service error", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.UpdateAccountRequest{})
		accountSvc.EXPECT().Update(gomock.Any(), req.Msg).Return(nil, auth.ErrInvalidToken)
		resp, err := app.UpdateAccount(ctx, req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
	t.Run("no auth", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		req := connect.NewRequest(&accountsv1.UpdateAccountRequest{})
		resp, err := app.UpdateAccount(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestAccountsDeleteAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.DeleteAccountRequest{})
		respMsg := &accountsv1.DeleteAccountResponse{}
		accountSvc.EXPECT().Delete(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := app.DeleteAccount(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})
	t.Run("service error", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.DeleteAccountRequest{})
		accountSvc.EXPECT().Delete(gomock.Any(), req.Msg).Return(nil, auth.ErrInvalidToken)
		resp, err := app.DeleteAccount(ctx, req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
	t.Run("no auth", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		req := connect.NewRequest(&accountsv1.DeleteAccountRequest{})
		resp, err := app.DeleteAccount(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestAccountsListAccounts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.ListAccountsRequest{})
		respMsg := &accountsv1.ListAccountsResponse{}
		accountSvc.EXPECT().List(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := app.ListAccounts(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})
	t.Run("service error", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&accountsv1.ListAccountsRequest{})
		accountSvc.EXPECT().List(gomock.Any(), req.Msg).Return(nil, auth.ErrInvalidToken)
		resp, err := app.ListAccounts(ctx, req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
	t.Run("no auth", func(t *testing.T) {
		accountSvc := NewMockAccountSvc(gomock.NewController(t))
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		app, _ := handlers.NewAccountsApi(grpc, accountSvc)
		req := connect.NewRequest(&accountsv1.ListAccountsRequest{})
		resp, err := app.ListAccounts(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
