package handlers

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/configuration/v1/configurationv1connect"
	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"connectrpc.com/connect"
	"context"
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
