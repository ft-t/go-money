package transactions

import (
	"context"
	"github.com/cockroachdb/errors"
	transactionsv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/transactions/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"time"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Create(
	ctx context.Context,
	req *transactionsv1.CreateTransactionRequest,
) (*transactionsv1.CreateTransactionResponse, error) {
	newTx := &database.Transaction{
		ID:                   0,
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
	}

	switch v := req.GetTransaction().(type) {
	case *transactionsv1.CreateTransactionRequest_TransferBetweenAccounts:
		if err := s.fillTransferBetweenAccounts(ctx, v.TransferBetweenAccounts, newTx); err != nil {
			return nil, err
		}
	case *transactionsv1.CreateTransactionRequest_Withdrawal
		if err := s.fillWithdrawal(ctx, v.Withdrawal, newTx); err != nil {
			return nil, err
		}
	}
}

func (s *Service) fillWithdrawal(
	ctx context.Context,
	req *transactionsv1.Withdrawal,
	newTx *database.Transaction,
) error {

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

	if err := s.ensureAccountExists(ctx, []int32{req.SourceAccountId, req.DestinationAccountId}); err != nil {
		return err
	}

	sourceAmount, err := decimal.NewFromString(req.SourceAmount)
	if err != nil {
		return errors.Wrap(err, "invalid source amount")
	}

	if sourceAmount.IsNegative() || sourceAmount.IsZero() {
		return errors.New("source amount must be positive")
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
	newTx.SourceCurrency = req.SourceCurrency
	newTx.DestinationCurrency = req.DestinationCurrency

	return nil
}

func (s *Service) ensureAccountExists(
	ctx context.Context,
	ids []int32,
) error {
	ids = lo.Uniq(ids)

	var accounts []int32
	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).
		Model(&database.Account{}).Where("id IN ?", ids).Pluck("id", &accounts).Error; err != nil {
		return err
	}

	missing, _ := lo.Difference(ids, accounts)
	if len(missing) > 0 {
		return errors.Newf("accounts with ids %v not found", missing)
	}

	return nil
}
