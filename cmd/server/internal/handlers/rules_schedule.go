package handlers

import (
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
)

func (r *RulesApi) ListScheduleRules(
	ctx context.Context,
	c *connect.Request[rulesv1.ListScheduleRulesRequest],
) (*connect.Response[rulesv1.ListScheduleRulesResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := r.cfg.RulesScheduleSvc.ListRules(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (r *RulesApi) CreateScheduleRule(
	ctx context.Context,
	c *connect.Request[rulesv1.CreateScheduleRuleRequest],
) (*connect.Response[rulesv1.CreateScheduleRuleResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := r.cfg.RulesScheduleSvc.CreateRule(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (r *RulesApi) UpdateScheduleRule(
	ctx context.Context,
	c *connect.Request[rulesv1.UpdateScheduleRuleRequest],
) (*connect.Response[rulesv1.UpdateScheduleRuleResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := r.cfg.RulesScheduleSvc.UpdateRule(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (r *RulesApi) DeleteScheduleRule(
	ctx context.Context,
	c *connect.Request[rulesv1.DeleteScheduleRuleRequest],
) (*connect.Response[rulesv1.DeleteScheduleRuleResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := r.cfg.RulesScheduleSvc.DeleteRule(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (r *RulesApi) ValidateCronExpression(
	ctx context.Context,
	c *connect.Request[rulesv1.ValidateCronExpressionRequest],
) (*connect.Response[rulesv1.ValidateCronExpressionResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	err := r.cfg.SchedulerSvc.ValidateCronExpression(c.Msg.CronExpression)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&rulesv1.ValidateCronExpressionResponse{
		Valid: true,
	}), nil
}
