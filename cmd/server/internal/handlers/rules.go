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
	ruleSvc RulesSvc
}

func NewRulesApi(
	mux *boilerplate.DefaultGrpcServer,
	ruleSvc RulesSvc,
) *RulesApi {
	res := &RulesApi{
		ruleSvc: ruleSvc,
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

	rules, err := r.ruleSvc.ListRules(ctx, c.Msg)
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

	rule, err := r.ruleSvc.CreateRule(ctx, c.Msg)
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

	rule, err := r.ruleSvc.UpdateRule(ctx, c.Msg)
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

	resp, err := r.ruleSvc.DeleteRule(ctx, c.Msg)
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

	resp, err := r.ruleSvc.DryRunRule(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}
