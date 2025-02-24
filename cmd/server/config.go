package main

import (
	"connectrpc.com/connect"
	"context"
	configurationv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/configuration/v1"
	"github.com/ft-t/go-money-pb/gen/gomoneypb/configuration/v1/accountsv1connect"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type ConfigApi struct {
	configSvc ConfigSvc
}

func NewConfigApi(
	mux *boilerplate.DefaultGrpcServer,
	userSvc ConfigSvc,
) (*ConfigApi, error) {
	res := &ConfigApi{
		configSvc: userSvc,
	}

	mux.GetMux().Handle(
		accountsv1connect.NewConfigurationServiceHandler(res, mux.GetDefaultHandlerOptions()...),
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
