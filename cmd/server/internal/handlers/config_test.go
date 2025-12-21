package handlers_test

import (
	"context"
	"net/http"
	"testing"

	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"connectrpc.com/connect"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

func TestNewConfigApi(t *testing.T) {
	mockSvc := NewMockConfigSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, err := handlers.NewConfigApi(grpc, mockSvc)
	assert.NoError(t, err)
	assert.NotNil(t, api)
}

func TestConfigApi_GetConfiguration(t *testing.T) {
	mockSvc := NewMockConfigSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc)

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
	mockSvc := NewMockConfigSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc)

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
	mockSvc := NewMockConfigSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewConfigApi(grpc, mockSvc)

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
