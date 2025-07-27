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

func newRulesScheduleApiWithMock(t *testing.T) (*handlers.RulesApi, *MockRulesScheduleSvc, *MockSchedulerSvc) {
	ctrl := gomock.NewController(t)
	scheduleSvc := NewMockRulesScheduleSvc(ctrl)
	schedulerSvc := NewMockSchedulerSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewRulesApi(grpc, &handlers.RulesApiConfig{
		RulesScheduleSvc: scheduleSvc,
		SchedulerSvc:     schedulerSvc,
	})
	return api, scheduleSvc, schedulerSvc
}

func TestRulesApi_ListScheduleRules(t *testing.T) {
	api, scheduleSvc, _ := newRulesScheduleApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.ListScheduleRulesRequest{})
		respMsg := &rulesv1.ListScheduleRulesResponse{}
		scheduleSvc.EXPECT().ListRules(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ListScheduleRules(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.ListScheduleRulesRequest{})
		scheduleSvc.EXPECT().ListRules(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ListScheduleRules(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.ListScheduleRulesRequest{})
		resp, err := api.ListScheduleRules(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_CreateScheduleRule(t *testing.T) {
	api, scheduleSvc, _ := newRulesScheduleApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.CreateScheduleRuleRequest{})
		respMsg := &rulesv1.CreateScheduleRuleResponse{}
		scheduleSvc.EXPECT().CreateRule(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.CreateScheduleRule(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.CreateScheduleRuleRequest{})
		scheduleSvc.EXPECT().CreateRule(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.CreateScheduleRule(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.CreateScheduleRuleRequest{})
		resp, err := api.CreateScheduleRule(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_UpdateScheduleRule(t *testing.T) {
	api, scheduleSvc, _ := newRulesScheduleApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.UpdateScheduleRuleRequest{})
		respMsg := &rulesv1.UpdateScheduleRuleResponse{}
		scheduleSvc.EXPECT().UpdateRule(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.UpdateScheduleRule(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.UpdateScheduleRuleRequest{})
		scheduleSvc.EXPECT().UpdateRule(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.UpdateScheduleRule(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.UpdateScheduleRuleRequest{})
		resp, err := api.UpdateScheduleRule(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_DeleteScheduleRule(t *testing.T) {
	api, scheduleSvc, _ := newRulesScheduleApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.DeleteScheduleRuleRequest{})
		respMsg := &rulesv1.DeleteScheduleRuleResponse{}
		scheduleSvc.EXPECT().DeleteRule(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.DeleteScheduleRule(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.DeleteScheduleRuleRequest{})
		scheduleSvc.EXPECT().DeleteRule(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.DeleteScheduleRule(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.DeleteScheduleRuleRequest{})
		resp, err := api.DeleteScheduleRule(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestRulesApi_ValidateCronExpression(t *testing.T) {
	api, _, schedulerSvc := newRulesScheduleApiWithMock(t)

	t.Run("valid", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.ValidateCronExpressionRequest{CronExpression: "* * * * *"})
		schedulerSvc.EXPECT().ValidateCronExpression(req.Msg.CronExpression).Return(nil)
		resp, err := api.ValidateCronExpression(ctx, req)
		assert.NoError(t, err)
		assert.True(t, resp.Msg.Valid)
	})

	t.Run("invalid", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&rulesv1.ValidateCronExpressionRequest{CronExpression: "bad"})
		schedulerSvc.EXPECT().ValidateCronExpression(req.Msg.CronExpression).Return(assert.AnError)
		resp, err := api.ValidateCronExpression(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&rulesv1.ValidateCronExpressionRequest{CronExpression: "* * * * *"})
		resp, err := api.ValidateCronExpression(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
