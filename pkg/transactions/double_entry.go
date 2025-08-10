package transactions

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"time"
)

const (
	roundPlaces = 18
)

type DoubleEntryConfig struct {
	BaseCurrency string
	AccountSvc   AccountSvc
}

type DoubleEntryService struct {
	cfg *DoubleEntryConfig
}

func NewDoubleEntryService(
	cfg *DoubleEntryConfig,
) *DoubleEntryService {
	return &DoubleEntryService{
		cfg: cfg,
	}
}

func (s *DoubleEntryService) Record(ctx context.Context, tx *database.Transaction) ([]*database.DoubleEntry, error) {
	if tx.SourceAccountID == nil {
		return nil, errors.New("source_account_id is required for double entry transactions")
	}

	if tx.DestinationAccountID == nil {
		return nil, errors.New("destination_account_id is required for double entry transactions")
	}

	if !tx.SourceAmountInBaseCurrency.Decimal.Abs().Round(roundPlaces).Equal(
		tx.DestinationAmountInBaseCurrency.Decimal.Abs().Round(roundPlaces),
	) {
		return nil, errors.New("source and destination amounts in base currency must be equal for double entry transactions")
	}

	if tx.SourceAmountInBaseCurrency.Decimal.Sign() == tx.DestinationAmountInBaseCurrency.Decimal.Sign() {
		return nil, errors.New("source and destination amounts must have opposite signs for double entry transactions")
	}

	sourceAcc, err := s.cfg.AccountSvc.GetAccount(ctx, *tx.SourceAccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get source account")
	}

	baseAmount := tx.SourceAmountInBaseCurrency.Decimal

	switch {
	case s.isDebitNormal(sourceAcc.Type):
		// Debit-normal credit => DECREASE => expect negative input
		if tx.SourceAmountInBaseCurrency.Decimal.IsPositive() {
			return nil, errors.New("source (debit-normal) must be negative for a credit")
		}
	default:
		// Credit-normal credit => INCREASE => expect positive input
		//if !tx.SourceAmountInBaseCurrency.Decimal.IsPositive() {
		//	return nil, errors.New("source (credit-normal) must be positive for a credit")
		//}
	}

	isDebit := s.isDebit(sourceAcc.Type, baseAmount)

	entries := []*database.DoubleEntry{
		{
			TransactionID:        tx.ID,
			IsDebit:              isDebit,
			AccountID:            *tx.SourceAccountID,
			BaseCurrency:         s.cfg.BaseCurrency,
			AmountInBaseCurrency: baseAmount.Abs(),
			CreatedAt:            time.Now().UTC(),
		},
		{
			TransactionID:        tx.ID,
			IsDebit:              !isDebit,
			AccountID:            *tx.DestinationAccountID,
			BaseCurrency:         s.cfg.BaseCurrency,
			AmountInBaseCurrency: baseAmount.Abs(),
			CreatedAt:            time.Now().UTC(),
		},
	}

	return entries, nil
}

func (s *DoubleEntryService) isDebitNormal(accountType gomoneypbv1.AccountType) bool {
	switch accountType {
	case gomoneypbv1.AccountType_ACCOUNT_TYPE_REGULAR,
		gomoneypbv1.AccountType_ACCOUNT_TYPE_SAVINGS,
		gomoneypbv1.AccountType_ACCOUNT_TYPE_BROKERAGE,
		gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE:
		return true
	default:
		return false
	}
}

func (s *DoubleEntryService) isDebit(accountType gomoneypbv1.AccountType, amount decimal.Decimal) bool {
	if s.isDebitNormal(accountType) {
		return amount.IsPositive() // debit-normal: + => debit, - => credit
	}
	return amount.IsNegative() // credit-normal: - => debit, + => credit
}
