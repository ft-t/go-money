package handlers_test

import (
	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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
