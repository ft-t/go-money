package transactions

import (
	"context"
	"github.com/cockroachdb/errors"
	transactionsv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/transactions/v1"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"time"
)

type Service struct {
	accountCurrencyCache *expirable.LRU[int32, string]
}

func NewService() *Service {
	return &Service{
		accountCurrencyCache: expirable.NewLRU[int32, string](100, nil, configuration.DefaultCacheTTL),
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
		TransactionDate:      req.TransactionDate.AsTime(),
	}

	switch v := req.GetTransaction().(type) {
	case *transactionsv1.CreateTransactionRequest_TransferBetweenAccounts:
		if err := s.fillTransferBetweenAccounts(ctx, v.TransferBetweenAccounts, newTx); err != nil {
			return nil, err
		}
	case *transactionsv1.CreateTransactionRequest_Withdrawal:
		if err := s.fillWithdrawal(ctx, v.Withdrawal, newTx); err != nil {
			return nil, err
		}
	case *transactionsv1.CreateTransactionRequest_Deposit:
		if err := s.fillDeposit(ctx, v.Deposit, newTx); err != nil {
			return nil, err
		}
	}

	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()

	return &transactionsv1.CreateTransactionResponse{}, nil
}

func (s *Service) fillDeposit(
	ctx context.Context,
	req *transactionsv1.Deposit,
	newTx *database.Transaction,
) error {
	if req.DestinationAccountId <= 0 {
		return errors.New("destination account id is required")
	}

	destinationAmount, err := decimal.NewFromString(req.DestinationAmount)
	if err != nil {
		return errors.Wrap(err, "invalid destination amount")
	}

	if destinationAmount.IsNegative() || destinationAmount.IsZero() {
		return errors.New("destination amount must be positive")
	}

	if err = s.ensureAccountsExistAndCurrencyCorrect(ctx, map[int32]string{
		req.DestinationAccountId: req.DestinationCurrency,
	}); err != nil {
		return err
	}

	newTx.DestinationAmount = destinationAmount
	newTx.DestinationCurrency = req.DestinationCurrency
	newTx.DestinationAccountID = &req.DestinationAccountId

	return nil
}

func (s *Service) fillWithdrawal(
	ctx context.Context,
	req *transactionsv1.Withdrawal,
	newTx *database.Transaction,
) error {
	if req.SourceAccountId <= 0 {
		return errors.New("source account id is required")
	}

	sourceAmount, err := decimal.NewFromString(req.SourceAmount)
	if err != nil {
		return errors.Wrap(err, "invalid source amount")
	}

	if sourceAmount.IsPositive() || sourceAmount.IsZero() {
		return errors.New("source amount must be negative")
	}

	if err = s.ensureAccountsExistAndCurrencyCorrect(ctx, map[int32]string{
		req.SourceAccountId: req.SourceCurrency,
	}); err != nil {
		return err
	}

	newTx.SourceAmount = sourceAmount
	newTx.SourceCurrency = req.SourceCurrency
	newTx.SourceAccountID = &req.SourceAccountId

	return nil
}

func (s *Service) fillTransferBetweenAccounts(
	ctx context.Context,
	req *transactionsv1.TransferBetweenAccounts,
	newTx *database.Transaction,
) error {
	if req.SourceAccountId <= 0 {
		return errors.New("source account id is required")
	}

	if req.DestinationAccountId <= 0 {
		return errors.New("destination account id is required")
	}

	if err := s.ensureAccountsExistAndCurrencyCorrect(ctx, map[int32]string{
		req.SourceAccountId:      req.SourceCurrency,
		req.DestinationAccountId: req.DestinationCurrency,
	}); err != nil {
		return err
	}

	sourceAmount, err := decimal.NewFromString(req.SourceAmount)
	if err != nil {
		return errors.Wrap(err, "invalid source amount")
	}

	if sourceAmount.IsPositive() || sourceAmount.IsZero() {
		return errors.New("source amount must be negative")
	}

	destinationAmount, err := decimal.NewFromString(req.DestinationAmount)
	if err != nil {
		return errors.Wrap(err, "invalid destination amount")
	}

	if destinationAmount.IsNegative() || destinationAmount.IsZero() {
		return errors.New("destination amount must be positive")
	}

	newTx.SourceAmount = sourceAmount
	newTx.DestinationAmount = destinationAmount

	newTx.SourceAccountID = &req.SourceAccountId
	newTx.DestinationAccountID = &req.DestinationAccountId

	newTx.SourceCurrency = req.SourceCurrency
	newTx.DestinationCurrency = req.DestinationCurrency

	return nil
}

func (s *Service) ensureAccountsExistAndCurrencyCorrect(
	ctx context.Context,
	expectedAccounts map[int32]string,
) error {
	var accounts []*database.Account

	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).
		Model(&database.Account{}).Where("id IN ?", lo.Keys(expectedAccounts)).
		Select("id, currency").Find(&accounts).Error; err != nil {
		return errors.WithStack(err)
	}

	existing := map[int32]string{}
	for _, acc := range accounts {
		existing[acc.ID] = acc.Currency
	}

	for id, expectedCurrency := range expectedAccounts {
		existingCurrency, ok := existing[id]

		if !ok {
			return errors.Newf("account with id %d not found", id)
		}

		if existingCurrency != expectedCurrency {
			return errors.Newf("account with id %d has currency %s, expected %s", id, existingCurrency, expectedCurrency)
		}
	}

	return nil
}
