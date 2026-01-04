package handlers_test

import (
	"context"
	"net/http"
	"testing"

	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"connectrpc.com/connect"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

func TestNewConfigApi(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, err := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)
	assert.NoError(t, err)
	assert.NotNil(t, api)
}

func TestConfigApi_GetConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	t.Run("success", func(t *testing.T) {
		req := connect.NewRequest(&configurationv1.GetConfigurationRequest{})
		respMsg := &configurationv1.GetConfigurationResponse{}
		mockSvc.EXPECT().GetConfiguration(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.GetConfiguration(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		req := connect.NewRequest(&configurationv1.GetConfigurationRequest{})
		mockSvc.EXPECT().GetConfiguration(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.GetConfiguration(context.TODO(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestConfigApi_GetConfigsByKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&configurationv1.GetConfigsByKeysRequest{
			Keys: []string{"key1"},
		})
		respMsg := &configurationv1.GetConfigsByKeysResponse{
			Configs: map[string]string{"key1": "value1"},
		}
		mockSvc.EXPECT().GetConfigsByKeys(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.GetConfigsByKeys(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&configurationv1.GetConfigsByKeysRequest{})
		mockSvc.EXPECT().GetConfigsByKeys(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.GetConfigsByKeys(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&configurationv1.GetConfigsByKeysRequest{})
		resp, err := api.GetConfigsByKeys(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestConfigApi_SetConfigByKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&configurationv1.SetConfigByKeyRequest{
			Key:   "key1",
			Value: "value1",
		})
		respMsg := &configurationv1.SetConfigByKeyResponse{Key: "key1"}
		mockSvc.EXPECT().SetConfigByKey(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.SetConfigByKey(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&configurationv1.SetConfigByKeyRequest{})
		mockSvc.EXPECT().SetConfigByKey(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.SetConfigByKey(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&configurationv1.SetConfigByKeyRequest{})
		resp, err := api.SetConfigByKey(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestConfigApi_GetServiceTokens_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
	req := connect.NewRequest(&configurationv1.GetServiceTokensRequest{
		Ids: []string{"token1"},
	})
	respMsg := &configurationv1.GetServiceTokensResponse{
		ServiceTokens: []*gomoneypbv1.ServiceToken{
			{Id: "token1", Name: "Test Token"},
		},
	}
	mockServiceTokenSvc.EXPECT().GetServiceTokens(gomock.Any(), req.Msg).Return(respMsg, nil)
	resp, err := api.GetServiceTokens(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, respMsg, resp.Msg)
}

func TestConfigApi_GetServiceTokens_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&configurationv1.GetServiceTokensRequest{})
		mockServiceTokenSvc.EXPECT().GetServiceTokens(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.GetServiceTokens(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&configurationv1.GetServiceTokensRequest{})
		resp, err := api.GetServiceTokens(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestConfigApi_CreateServiceToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
	req := connect.NewRequest(&configurationv1.CreateServiceTokenRequest{
		Name: "Test Token",
	})
	respMsg := &configurationv1.CreateServiceTokenResponse{
		ServiceToken: &gomoneypbv1.ServiceToken{Id: "token1", Name: "Test Token"},
		Token:        "jwt.token.here",
	}
	mockServiceTokenSvc.EXPECT().CreateServiceToken(gomock.Any(), &auth.CreateServiceTokenRequest{
		Req:           req.Msg,
		CurrentUserID: 1,
	}).Return(respMsg, nil)
	resp, err := api.CreateServiceToken(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, respMsg, resp.Msg)
}

func TestConfigApi_CreateServiceToken_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&configurationv1.CreateServiceTokenRequest{Name: "Test"})
		mockServiceTokenSvc.EXPECT().CreateServiceToken(gomock.Any(), &auth.CreateServiceTokenRequest{
			Req:           req.Msg,
			CurrentUserID: 1,
		}).Return(nil, assert.AnError)
		resp, err := api.CreateServiceToken(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&configurationv1.CreateServiceTokenRequest{})
		resp, err := api.CreateServiceToken(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestConfigApi_RevokeServiceToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
	req := connect.NewRequest(&configurationv1.RevokeServiceTokenRequest{
		Id: "token1",
	})
	respMsg := &configurationv1.RevokeServiceTokenResponse{
		Id: "token1",
	}
	mockServiceTokenSvc.EXPECT().RevokeServiceToken(gomock.Any(), req.Msg).Return(respMsg, nil)
	resp, err := api.RevokeServiceToken(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, respMsg, resp.Msg)
}

func TestConfigApi_RevokeServiceToken_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockConfigSvc(ctrl)
	mockServiceTokenSvc := NewMockServiceTokenSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc, mockServiceTokenSvc)

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&configurationv1.RevokeServiceTokenRequest{Id: "token1"})
		mockServiceTokenSvc.EXPECT().RevokeServiceToken(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.RevokeServiceToken(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&configurationv1.RevokeServiceTokenRequest{})
		resp, err := api.RevokeServiceToken(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
