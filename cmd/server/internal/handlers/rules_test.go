package handlers_test

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
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

func newRulesApiWithMock(t *testing.T) (*handlers.RulesApi, *MockRulesSvc) {
	ctrl := gomock.NewController(t)
	ruleSvc := NewMockRulesSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewRulesApi(grpc, ruleSvc)
	return api, ruleSvc
}

func TestRulesApi_ListRules(t *testing.T) {
	api, ruleSvc := newRulesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.ListRulesRequest{})
		respMsg := &rulesv1.ListRulesResponse{}
		ruleSvc.EXPECT().ListRules(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ListRules(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.ListRulesRequest{})
		ruleSvc.EXPECT().ListRules(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ListRules(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.ListRulesRequest{})
		resp, err := api.ListRules(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_CreateRule(t *testing.T) {
	api, ruleSvc := newRulesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.CreateRuleRequest{})
		respMsg := &rulesv1.CreateRuleResponse{}
		ruleSvc.EXPECT().CreateRule(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.CreateRule(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.CreateRuleRequest{})
		ruleSvc.EXPECT().CreateRule(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.CreateRule(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.CreateRuleRequest{})
		resp, err := api.CreateRule(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_UpdateRule(t *testing.T) {
	api, ruleSvc := newRulesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.UpdateRuleRequest{})
		respMsg := &rulesv1.UpdateRuleResponse{}
		ruleSvc.EXPECT().UpdateRule(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.UpdateRule(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.UpdateRuleRequest{})
		ruleSvc.EXPECT().UpdateRule(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.UpdateRule(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.UpdateRuleRequest{})
		resp, err := api.UpdateRule(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_DeleteRule(t *testing.T) {
	api, ruleSvc := newRulesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.DeleteRuleRequest{})
		respMsg := &rulesv1.DeleteRuleResponse{}
		ruleSvc.EXPECT().DeleteRule(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.DeleteRule(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.DeleteRuleRequest{})
		ruleSvc.EXPECT().DeleteRule(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.DeleteRule(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.DeleteRuleRequest{})
		resp, err := api.DeleteRule(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_DryRunRule(t *testing.T) {
	api, ruleSvc := newRulesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.DryRunRuleRequest{})
		respMsg := &rulesv1.DryRunRuleResponse{}
		ruleSvc.EXPECT().DryRunRule(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.DryRunRule(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.DryRunRuleRequest{})
		ruleSvc.EXPECT().DryRunRule(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.DryRunRule(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.DryRunRuleRequest{})
		resp, err := api.DryRunRule(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
