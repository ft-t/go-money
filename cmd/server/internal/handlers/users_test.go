package handlers_test

import (
	usersv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/users/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewUserApi(t *testing.T) {
	mockSvc := NewMockUserSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, err := handlers.NewUserApi(grpc, mockSvc)
	assert.NoError(t, err)
	assert.NotNil(t, api)
}

func TestUserApi_Create(t *testing.T) {
	mockSvc := NewMockUserSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewUserApi(grpc, mockSvc)

	t.Run("success", func(t *testing.T) {
		req := connect.NewRequest(&usersv1.CreateRequest{})
		respMsg := &usersv1.CreateResponse{}
		mockSvc.EXPECT().Create(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.Create(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		req := connect.NewRequest(&usersv1.CreateRequest{})
		mockSvc.EXPECT().Create(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.Create(context.TODO(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestUserApi_Login(t *testing.T) {
	mockSvc := NewMockUserSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewUserApi(grpc, mockSvc)

	t.Run("success", func(t *testing.T) {
		req := connect.NewRequest(&usersv1.LoginRequest{})
		respMsg := &usersv1.LoginResponse{}
		mockSvc.EXPECT().Login(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.Login(context.TODO(), req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		req := connect.NewRequest(&usersv1.LoginRequest{})
		mockSvc.EXPECT().Login(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.Login(context.TODO(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
