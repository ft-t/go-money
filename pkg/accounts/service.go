package accounts

import (
	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Service struct {
	cfg *ServiceConfig
}

func (s *Service) List(ctx context.Context, req *accountsv1.ListAccountsRequest) (*accountsv1.ListAccountsResponse, error) {
	var accounts []database.Account

	db := database.GetDbWithContext(ctx, database.DbTypeReadonly)

	if err := db.Unscoped().Order("position desc").Find(&accounts).Error; err != nil {
		return nil, err
	}

	var mapped []*accountsv1.ListAccountsResponse_AccountItem

	for _, account := range accounts {
		mapped = append(mapped,
			&accountsv1.ListAccountsResponse_AccountItem{
				Account: s.cfg.MapperSvc.MapAccount(ctx, &account),
			})
	}

	return &accountsv1.ListAccountsResponse{
		Accounts: mapped,
	}, nil
}

func (s *Service) Delete(ctx context.Context, req *accountsv1.DeleteAccountRequest) (*accountsv1.DeleteAccountResponse, error) {
	var account database.Account

	db := database.GetDbWithContext(ctx, database.DbTypeMaster)

	if err := db.Where("id = ?", req.Id).First(&account).Error; err != nil {
		return nil, errors.Join(err, errors.New("account not found"))
	}

	if err := db.Delete(&account).Error; err != nil {
		return nil, err
	}

	return &accountsv1.DeleteAccountResponse{
		Account: s.cfg.MapperSvc.MapAccount(ctx, &account),
	}, nil
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
		Name:          req.Name,
		Currency:      req.Currency,
		Extra:         req.Extra,
		Flags:         0,
		LastUpdatedAt: time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
		DeletedAt:     gorm.DeletedAt{},
		Type:          req.Type,
		Note:          req.Note,
		Iban:          req.Iban,
		AccountNumber: req.AccountNumber,
	}

	liabilityPercent, err := s.parseLiabilityPercent(req.LiabilityPercent)
	if err != nil {
		return nil, err
	}

	account.LiabilityPercent = liabilityPercent

	if err = database.GetDbWithContext(ctx, database.DbTypeMaster).Create(account).Error; err != nil {
		return nil, err
	}

	return &accountsv1.CreateAccountResponse{
		Account: s.cfg.MapperSvc.MapAccount(ctx, account),
	}, nil
}

func (s *Service) parseLiabilityPercent(input *string) (decimal.NullDecimal, error) {
	if input == nil {
		return decimal.NullDecimal{}, nil
	}

	parsed, err := decimal.NewFromString(*input)
	if err != nil {
		return decimal.NullDecimal{}, errors.Join(err, errors.New("failed to parse liability percent"))
	}

	return decimal.NewNullDecimal(parsed), nil
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
	account.Note = req.Note
	account.AccountNumber = req.AccountNumber
	account.Iban = req.Iban

	liabilityPercent, err := s.parseLiabilityPercent(req.LiabilityPercent)
	if err != nil {
		return nil, err
	}

	account.LiabilityPercent = liabilityPercent

	if account.Extra == nil {
		account.Extra = map[string]string{}
	}

	if err = tx.Save(&account).Error; err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return &accountsv1.UpdateAccountResponse{
		Account: s.cfg.MapperSvc.MapAccount(ctx, &account),
	}, nil
}
