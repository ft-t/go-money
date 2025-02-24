package accounts

import (
	"context"
	accountsv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/accounts/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Service struct {
	cfg *ServiceConfig
}

type ServiceConfig struct {
	MapperSvc MapperSvc
}

func NewService(
	cfg *ServiceConfig,
) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) Create(
	ctx context.Context,
	req *accountsv1.CreateAccountRequest,
) (*accountsv1.CreateAccountResponse, error) {
	account := &database.Account{
		Name:           req.Account.Name,
		Currency:       req.Account.Currency,
		CurrentBalance: decimal.Decimal{},
		Extra:          req.Account.Extra,
		Flags:          0,
		LastUpdatedAt:  time.Now().UTC(),
		CreatedAt:      time.Now().UTC(),
		DeletedAt:      gorm.DeletedAt{},
		Type:           req.Account.Type,
	}

	if err := database.GetDbWithContext(ctx, database.DbTypeMaster).Create(account).Error; err != nil {
		return nil, err
	}

	return &accountsv1.CreateAccountResponse{
		Account: s.cfg.MapperSvc.MapAccount(ctx, account),
	}, nil
}

func (s *Service) Update(
	ctx context.Context,
	req *accountsv1.UpdateAccountRequest,
) (*accountsv1.UpdateAccountResponse, error) {
	var account database.Account

	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()

	if err := tx.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).First(&account, req.Id).Error; err != nil {
		return nil, err
	}

	account.Name = req.Name
	account.Type = req.Type
	account.Extra = req.Extra
	account.LastUpdatedAt = time.Now().UTC()

	if err := tx.Save(&account).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &accountsv1.UpdateAccountResponse{
		Account: s.cfg.MapperSvc.MapAccount(ctx, &account),
	}, nil
}
