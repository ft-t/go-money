package handlers

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/rules/v1/rulesv1connect"
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type RulesApi struct {
	cfg *RulesApiConfig
}

type RulesApiConfig struct {
	RulesScheduleSvc RulesScheduleSvc
	RuleSvc          RulesSvc
	DryRunSvc        DryRunSvc
	SchedulerSvc     SchedulerSvc
}

func NewRulesApi(
	mux *boilerplate.DefaultGrpcServer,
	cfg *RulesApiConfig,
) *RulesApi {
	res := &RulesApi{
		cfg: cfg,
	}

	mux.GetMux().Handle(
		rulesv1connect.NewRulesServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res
}

func (r *RulesApi) ListRules(ctx context.Context, c *connect.Request[rulesv1.ListRulesRequest]) (*connect.Response[rulesv1.ListRulesResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	rules, err := r.cfg.RuleSvc.ListRules(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(rules), nil
}

func (r *RulesApi) CreateRule(ctx context.Context, c *connect.Request[rulesv1.CreateRuleRequest]) (*connect.Response[rulesv1.CreateRuleResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	rule, err := r.cfg.RuleSvc.CreateRule(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(rule), nil
}

func (r *RulesApi) UpdateRule(ctx context.Context, c *connect.Request[rulesv1.UpdateRuleRequest]) (*connect.Response[rulesv1.UpdateRuleResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	rule, err := r.cfg.RuleSvc.UpdateRule(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(rule), nil
}

func (r *RulesApi) DeleteRule(ctx context.Context, c *connect.Request[rulesv1.DeleteRuleRequest]) (*connect.Response[rulesv1.DeleteRuleResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := r.cfg.RuleSvc.DeleteRule(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (r *RulesApi) DryRunRule(ctx context.Context, c *connect.Request[rulesv1.DryRunRuleRequest]) (*connect.Response[rulesv1.DryRunRuleResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := r.cfg.DryRunSvc.DryRunRule(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}
