package accounts

import (
	"context"
	"fmt"
	"time"

	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	cfg *ServiceConfig
}

type ServiceConfig struct {
	MapperSvc       MapperSvc
	DefaultCurrency string
}

func NewService(
	cfg *ServiceConfig,
) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) GetAccountByID(ctx context.Context, id int32) (*database.Account, error) {
	var account database.Account

	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).Where("id = ?", id).
		First(&account).Error; err != nil {
		return nil, errors.Join(err, errors.New("failed to fetch account by id"))
	}

	return &account, nil
}

func (s *Service) GetAllAccounts(ctx context.Context) ([]*database.Account, error) {
	var accounts []*database.Account

	if err := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly)).
		Find(&accounts).Error; err != nil {
		return nil, errors.Join(err, errors.New("failed to fetch accounts"))
	}

	return accounts, nil
}

func (s *Service) GetDefaultAccount(ctx context.Context, accountType gomoneypbv1.AccountType) (*database.Account, error) {
	var account database.Account

	if err := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly)).
		Where("flags >= ? and flags & ? = ? and type = ? and deleted_at is null",
			database.AccountFlagIsDefault,
			database.AccountFlagIsDefault,
			database.AccountFlagIsDefault,
			accountType,
		).First(&account).Error; err != nil {
		return nil, err
	}

	return &account, nil
}

func (s *Service) List(ctx context.Context, req *accountsv1.ListAccountsRequest) (*accountsv1.ListAccountsResponse, error) {
	var accounts []*database.Account

	query := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly)).
		Order("display_order asc nulls last")

	if len(req.Ids) > 0 {
		query = query.Where("id in ?", req.Ids)
	}

	if req.IncludeDeleted {
		query = query.Unscoped()
	}

	if err := query.Find(&accounts).Error; err != nil {
		return nil, err
	}

	var mapped []*accountsv1.ListAccountsResponse_AccountItem

	for _, account := range accounts {
		mapped = append(mapped,
			&accountsv1.ListAccountsResponse_AccountItem{
				Account: s.cfg.MapperSvc.MapAccount(ctx, account),
			})
	}

	return &accountsv1.ListAccountsResponse{
		Accounts: mapped,
	}, nil
}

func (s *Service) Delete(ctx context.Context, req *accountsv1.DeleteAccountRequest) (*accountsv1.DeleteAccountResponse, error) {
	var account database.Account

	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	if err := tx.Where("id = ?", req.Id).First(&account).Error; err != nil {
		return nil, errors.Join(err, errors.New("account not found"))
	}

	if err := tx.Delete(&account).Error; err != nil {
		return nil, err
	}

	if err := s.EnsureDefaultExists(ctx, tx, &account); err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.Join(err, errors.New("failed to commit transaction"))
	}

	return &accountsv1.DeleteAccountResponse{
		Account: s.cfg.MapperSvc.MapAccount(ctx, &account),
	}, nil
}

func (s *Service) EnsureDefaultExists(
	_ context.Context,
	tx *gorm.DB,
	updatedAcc *database.Account,
) error {
	if updatedAcc.IsDefault() {
		if err := tx.Exec("update accounts set flags = flags & ~CAST(? AS bigint) where id != ? and type = ? and deleted_at is null",
			database.AccountFlagIsDefault,
			updatedAcc.ID,
			updatedAcc.Type,
		).Error; err != nil {
			return err
		}
	}

	// ensure that we have at least one default account
	var count int64
	if err := tx.Raw("select count(*) from accounts where flags & ? = ? and type = ? and deleted_at is null",
		database.AccountFlagIsDefault,
		database.AccountFlagIsDefault,
		updatedAcc.Type).Find(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		return errors.New("at least one default account is required")
	}

	return nil
}

func (s *Service) CreateBulk(
	ctx context.Context,
	req *accountsv1.CreateAccountsBulkRequest,
) (*accountsv1.CreateAccountsBulkResponse, error) {
	var existingAccounts []*database.Account

	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	if err := tx.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).Find(&existingAccounts).Error; err != nil {
		return nil, errors.Join(err, errors.New("failed to fetch existing accounts"))
	}

	accountMap := map[string]struct{}{}
	for _, account := range existingAccounts {
		key := fmt.Sprintf("%s-%s-%s", account.Name, account.Type, account.Currency)
		accountMap[key] = struct{}{}
	}

	var messages []string

	finalResp := &accountsv1.CreateAccountsBulkResponse{}

	for _, toCreate := range req.Accounts {
		key := fmt.Sprintf("%s-%s-%s", toCreate.Name, toCreate.Type, toCreate.Currency)

		if _, exists := accountMap[key]; exists {
			messages = append(messages,
				fmt.Sprintf("account with name '%s', type '%s', and currency '%s' already exists",
					toCreate.Name,
					toCreate.Type,
					toCreate.Currency,
				))

			finalResp.DuplicateCount += 1

			continue
		}

		if acc, err := s.Create(ctx, toCreate); err != nil {
			return nil, errors.Join(err, errors.New("failed to create account"))
		} else {
			messages = append(messages,
				fmt.Sprintf("account with id '%v' and and key '%s' created successfully",
					acc.Account.Id,
					key,
				))

			accountMap[key] = struct{}{}

			finalResp.CreatedCount += 1
		}
	}

	finalResp.Messages = messages

	return finalResp, nil
}

func (s *Service) Create(
	ctx context.Context,
	req *accountsv1.CreateAccountRequest,
) (*accountsv1.CreateAccountResponse, error) {
	account := &database.Account{
		Name:          req.Name,
		Currency:      req.Currency,
		Extra:         req.Extra,
		Flags:         req.Flags,
		LastUpdatedAt: time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
		DeletedAt:     gorm.DeletedAt{},
		Type:          req.Type,
		Note:          req.Note,
		Iban:          req.Iban,
		AccountNumber: req.AccountNumber,
		DisplayOrder:  req.DisplayOrder,
	}

	if account.Extra == nil {
		account.Extra = map[string]string{}
	}

	liabilityPercent, err := s.parseLiabilityPercent(req.LiabilityPercent)
	if err != nil {
		return nil, err
	}

	account.LiabilityPercent = liabilityPercent

	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	if err = tx.Create(account).Error; err != nil {
		return nil, err
	}

	if err = s.EnsureDefaultExists(ctx, tx, account); err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Join(err, errors.New("failed to commit transaction"))
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

	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
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
	account.DisplayOrder = req.DisplayOrder
	account.Flags = req.Flags

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

	if err = s.EnsureDefaultExists(ctx, tx, &account); err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, err
	}

	return &accountsv1.UpdateAccountResponse{
		Account: s.cfg.MapperSvc.MapAccount(ctx, &account),
	}, nil
}

func (s *Service) EnsureDefaultAccountsExist(
	ctx context.Context,
) error {
	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	for name, accType := range boilerplate.RequiredDefaultAccounts {
		acc, err := s.GetDefaultAccount(ctx, accType)

		if acc != nil {
			continue
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		account := &database.Account{
			Name:          name,
			Currency:      s.cfg.DefaultCurrency,
			Flags:         database.AccountFlagIsDefault,
			Type:          accType,
			Extra:         map[string]string{},
			CreatedAt:     time.Now().UTC(),
			LastUpdatedAt: time.Now().UTC(),
		}

		if err = tx.Create(account).Error; err != nil {
			return errors.Join(err, errors.New("failed to create default account"))
		}
	}

	return tx.Commit().Error
}
