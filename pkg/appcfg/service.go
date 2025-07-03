package appcfg

import (
	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
)

type Service struct {
	cfg *ServiceConfig
}

type ServiceConfig struct {
	UserSvc UserSvc
	AppCfg  *configuration.Configuration
}

func NewService(
	cfg *ServiceConfig,
) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) GetConfiguration(
	ctx context.Context,
	_ *configurationv1.GetConfigurationRequest,
) (*configurationv1.GetConfigurationResponse, error) {
	shouldCreatedAdmin, err := s.cfg.UserSvc.ShouldCreateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	return &configurationv1.GetConfigurationResponse{
		ShouldCreateAdmin: shouldCreatedAdmin,
		BaseCurrency:      configuration.BaseCurrency,
		GrafanaUrl:        s.cfg.AppCfg.GrafanaConfig.Url,
	}, nil
}
