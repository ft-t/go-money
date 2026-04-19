package appcfg

import (
	"context"
	"time"

	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm/clause"
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

	var configs []*database.AppConfig

	db := database.FromContext(ctx, database.GetDb(database.DbTypeReadonly))
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
	db := database.FromContext(ctx, database.GetDb(database.DbTypeMaster))

	now := time.Now().UTC()
	cfg := database.AppConfig{
		ID:        req.Key,
		Value:     req.Value,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		TargetWhere: clause.Where{
			Exprs: []clause.Expression{clause.Expr{SQL: "deleted_at IS NULL"}},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"value":      req.Value,
			"updated_at": now,
		}),
	}).Create(&cfg).Error; err != nil {
		return nil, err
	}

	return &configurationv1.SetConfigByKeyResponse{
		Key: req.Key,
	}, nil
}
