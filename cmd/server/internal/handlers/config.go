package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/configuration/v1/configurationv1connect"
	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"connectrpc.com/connect"

	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type ConfigApi struct {
	configSvc       ConfigSvc
	serviceTokenSvc ServiceTokenSvc
}

func NewConfigApi(
	mux *boilerplate.DefaultGrpcServer,
	userSvc ConfigSvc,
	serviceTokenSvc ServiceTokenSvc,
) (*ConfigApi, error) {
	res := &ConfigApi{
		configSvc:       userSvc,
		serviceTokenSvc: serviceTokenSvc,
	}

	mux.GetMux().Handle(
		configurationv1connect.NewConfigurationServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res, nil
}

func (a *ConfigApi) GetConfiguration(
	ctx context.Context,
	c *connect.Request[configurationv1.GetConfigurationRequest],
) (*connect.Response[configurationv1.GetConfigurationResponse], error) {
	res, err := a.configSvc.GetConfiguration(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (a *ConfigApi) GetConfigsByKeys(
	ctx context.Context,
	c *connect.Request[configurationv1.GetConfigsByKeysRequest],
) (*connect.Response[configurationv1.GetConfigsByKeysResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	res, err := a.configSvc.GetConfigsByKeys(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (a *ConfigApi) SetConfigByKey(
	ctx context.Context,
	c *connect.Request[configurationv1.SetConfigByKeyRequest],
) (*connect.Response[configurationv1.SetConfigByKeyResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	res, err := a.configSvc.SetConfigByKey(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (a *ConfigApi) GetServiceTokens(
	ctx context.Context,
	c *connect.Request[configurationv1.GetServiceTokensRequest],
) (*connect.Response[configurationv1.GetServiceTokensResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	res, err := a.serviceTokenSvc.GetServiceTokens(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (a *ConfigApi) CreateServiceToken(
	ctx context.Context,
	c *connect.Request[configurationv1.CreateServiceTokenRequest],
) (*connect.Response[configurationv1.CreateServiceTokenResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	res, err := a.serviceTokenSvc.CreateServiceToken(ctx, &auth.CreateServiceTokenRequest{
		Req:           c.Msg,
		CurrentUserID: jwtData.UserID,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (a *ConfigApi) RevokeServiceToken(
	ctx context.Context,
	c *connect.Request[configurationv1.RevokeServiceTokenRequest],
) (*connect.Response[configurationv1.RevokeServiceTokenResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	res, err := a.serviceTokenSvc.RevokeServiceToken(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}
