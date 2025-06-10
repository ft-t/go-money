package transactions

import (
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type Service struct {
	accountCurrencyCache *expirable.LRU[int32, string]
	cfg                  *ServiceConfig
}

type ServiceConfig struct {
	StatsSvc StatsSvc
}

func NewService(
	cfg *ServiceConfig,
) *Service {
	return &Service{
		accountCurrencyCache: expirable.NewLRU[int32, string](100, nil, configuration.DefaultCacheTTL),
		cfg:                  cfg,
	}
}

func (s *Service) Create(
	ctx context.Context,
	req *transactionsv1.CreateTransactionRequest,
) (*transactionsv1.CreateTransactionResponse, error) {
	if req.TransactionDate == nil {
		return nil, errors.New("transaction date is required")
	}

	newTx := &database.Transaction{
		SourceAmount:         decimal.Decimal{},
		SourceCurrency:       "",
		DestinationAmount:    decimal.Decimal{},
		DestinationCurrency:  "",
		SourceAccountID:      nil,
		DestinationAccountID: nil,
		LabelIDs:             req.LabelIds,
		CreatedAt:            time.Now().UTC(),
		Notes:                req.Notes,
		Extra:                req.Extra,
		TransactionDateTime:  req.TransactionDate.AsTime(),
		TransactionDateOnly:  req.TransactionDate.AsTime(),
	}

	if newTx.Extra == nil {
		newTx.Extra = map[string]string{}
	}

	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()

	var fillRes *fillResponse
	var err error

	switch v := req.GetTransaction().(type) {
	case *transactionsv1.CreateTransactionRequest_TransferBetweenAccounts:
		if fillRes, err = s.fillTransferBetweenAccounts(ctx, tx, v.TransferBetweenAccounts, newTx); err != nil {
			return nil, err
		}
	case *transactionsv1.CreateTransactionRequest_Withdrawal:
		if fillRes, err = s.fillWithdrawal(ctx, tx, v.Withdrawal, newTx); err != nil {
			return nil, err
		}
	case *transactionsv1.CreateTransactionRequest_Deposit:
		if fillRes, err = s.fillDeposit(ctx, tx, v.Deposit, newTx); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid transaction type")
	}

	for _, acc := range fillRes.Accounts {
		if acc.FirstTransactionAt == nil || acc.FirstTransactionAt.After(newTx.TransactionDateTime) {
			acc.FirstTransactionAt = &newTx.TransactionDateTime
		}
	}

	// validate wallet transaction date

	if err = tx.Create(newTx).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err = s.cfg.StatsSvc.HandleTransactions(ctx, tx, []*database.Transaction{newTx}); err != nil { // CALL BEFORE CREATE
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &transactionsv1.CreateTransactionResponse{}, nil
}

func (s *Service) fillDeposit(
	ctx context.Context,
	dbTx *gorm.DB,
	req *transactionsv1.Deposit,
	newTx *database.Transaction,
) (*fillResponse, error) {
	if req.DestinationAccountId <= 0 {
		return nil, errors.New("destination account id is required")
	}

	destinationAmount, err := decimal.NewFromString(req.DestinationAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid destination amount")
	}

	if destinationAmount.IsNegative() || destinationAmount.IsZero() {
		return nil, errors.New("destination amount must be positive")
	}

	accounts, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, map[int32]string{
		req.DestinationAccountId: req.DestinationCurrency,
	})
	if err != nil {
		return nil, err
	}

	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT
	newTx.DestinationAmount = destinationAmount
	newTx.DestinationCurrency = req.DestinationCurrency
	newTx.DestinationAccountID = &req.DestinationAccountId

	return &fillResponse{
		Accounts: accounts,
	}, nil
}

func (s *Service) fillWithdrawal(
	ctx context.Context,
	dbTx *gorm.DB,
	req *transactionsv1.Withdrawal,
	newTx *database.Transaction,
) (*fillResponse, error) {
	if req.SourceAccountId <= 0 {
		return nil, errors.New("source account id is required")
	}

	sourceAmount, err := decimal.NewFromString(req.SourceAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid source amount")
	}

	if sourceAmount.IsPositive() || sourceAmount.IsZero() {
		return nil, errors.New("source amount must be negative")
	}

	accounts, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, map[int32]string{
		req.SourceAccountId: req.SourceCurrency,
	})
	if err != nil {
		return nil, err
	}

	newTx.SourceAmount = sourceAmount
	newTx.SourceCurrency = req.SourceCurrency
	newTx.SourceAccountID = &req.SourceAccountId
	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL

	return &fillResponse{
		Accounts: accounts,
	}, nil
}

func (s *Service) fillTransferBetweenAccounts(
	ctx context.Context,
	dbTx *gorm.DB,
	req *transactionsv1.TransferBetweenAccounts,
	newTx *database.Transaction,
) (*fillResponse, error) {
	if req.SourceAccountId <= 0 {
		return nil, errors.New("source account id is required")
	}

	if req.DestinationAccountId <= 0 {
		return nil, errors.New("destination account id is required")
	}

	accounts, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, map[int32]string{
		req.SourceAccountId:      req.SourceCurrency,
		req.DestinationAccountId: req.DestinationCurrency,
	})
	if err != nil {
		return nil, err
	}

	sourceAmount, err := decimal.NewFromString(req.SourceAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid source amount")
	}

	if sourceAmount.IsPositive() || sourceAmount.IsZero() {
		return nil, errors.New("source amount must be negative")
	}

	destinationAmount, err := decimal.NewFromString(req.DestinationAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid destination amount")
	}

	if destinationAmount.IsNegative() || destinationAmount.IsZero() {
		return nil, errors.New("destination amount must be positive")
	}

	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS

	newTx.SourceAmount = sourceAmount
	newTx.DestinationAmount = destinationAmount

	newTx.SourceAccountID = &req.SourceAccountId
	newTx.DestinationAccountID = &req.DestinationAccountId

	newTx.SourceCurrency = req.SourceCurrency
	newTx.DestinationCurrency = req.DestinationCurrency

	return &fillResponse{
		Accounts: accounts,
	}, nil
}

func (s *Service) ensureAccountsExistAndCurrencyCorrect(
	_ context.Context,
	dbTx *gorm.DB,
	expectedAccounts map[int32]string,
) (map[int32]*database.Account, error) {
	var accounts []*database.Account

	if err := dbTx.
		Where("id IN ?", lo.Keys(expectedAccounts)).
		Clauses(&clause.Locking{Strength: "UPDATE"}).
		Find(&accounts).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	accCurrencies := map[int32]string{}
	accMap := map[int32]*database.Account{}
	for _, acc := range accounts {
		accMap[acc.ID] = acc
		accCurrencies[acc.ID] = acc.Currency
	}

	for id, expectedCurrency := range expectedAccounts {
		existingCurrency, ok := accCurrencies[id]

		if !ok {
			return nil, errors.Newf("account with id %d not found", id)
		}

		if existingCurrency != expectedCurrency {
			return nil, errors.Newf("account with id %d has currency %s, expected %s", id, existingCurrency, expectedCurrency)
		}
	}

	return accMap, nil
}
