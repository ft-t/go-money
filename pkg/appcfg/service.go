package appcfg

import (
	"context"
	"errors"
	"time"

	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"gorm.io/gorm"

	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
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
		BaseCurrency:      s.cfg.AppCfg.CurrencyConfig.BaseCurrency,
		GrafanaUrl:        s.cfg.AppCfg.GrafanaConfig.Url,
		BackendVersion:    boilerplate.GetVersion(),
		CommitSha:         boilerplate.GetCommit(),
	}, nil
}

func (s *Service) GetConfigsByKeys(
	ctx context.Context,
	req *configurationv1.GetConfigsByKeysRequest,
) (*configurationv1.GetConfigsByKeysResponse, error) {
	if len(req.Keys) == 0 {
		return &configurationv1.GetConfigsByKeysResponse{
			Configs: make(map[string]string),
		}, nil
	}

	var configs []database.AppConfig

	db := database.GetDbWithContext(ctx, database.DbTypeReadonly)
	if err := db.Where("id IN ?", req.Keys).Find(&configs).Error; err != nil {
		return nil, err
	}

	result := make(map[string]string, len(configs))
	for _, cfg := range configs {
		result[cfg.ID] = cfg.Value
	}

	return &configurationv1.GetConfigsByKeysResponse{
		Configs: result,
	}, nil
}

func (s *Service) SetConfigByKey(
	ctx context.Context,
	req *configurationv1.SetConfigByKeyRequest,
) (*configurationv1.SetConfigByKeyResponse, error) {
	db := database.GetDbWithContext(ctx, database.DbTypeMaster)

	var existing database.AppConfig
	if err := db.Unscoped().Where("id = ?", req.Key).First(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	now := time.Now().UTC()

	if existing.ID == "" {
		existing = database.AppConfig{
			ID:        req.Key,
			Value:     req.Value,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Create(&existing).Error; err != nil {
			return nil, err
		}
	} else {
		existing.Value = req.Value
		existing.UpdatedAt = now
		existing.DeletedAt.Valid = false

		if err := db.Unscoped().Save(&existing).Error; err != nil {
			return nil, err
		}
	}

	return &configurationv1.SetConfigByKeyResponse{
		Key: req.Key,
	}, nil
}
