package transactions

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"time"
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

func (s *DoubleEntryService) Record(
	ctx context.Context,
	tx *database.Transaction,
) error {
	if tx.SourceAccountID == nil {
		return errors.New("source_account_id is required for double entry transactions")
	}

	if tx.DestinationAccountID == nil {
		return errors.New("destination_account_id is required for double entry transactions")
	}

	if tx.SourceAmountInBaseCurrency.Decimal.Abs().String() != tx.DestinationAmountInBaseCurrency.Decimal.Abs().String() {
		return errors.New("source and destination amounts in base currency must be equal for double entry transactions")
	}

	sourceAcc, err := s.cfg.AccountSvc.GetAccount(ctx, *tx.SourceAccountID)
	if err != nil {
		return errors.Wrapf(err, "failed to get source account")
	}

	baseAmount := tx.SourceAmountInBaseCurrency.Decimal

	isDebitNormal := func(t gomoneypbv1.AccountType) bool {
		switch t {
		case gomoneypbv1.AccountType_ACCOUNT_TYPE_REGULAR,
			gomoneypbv1.AccountType_ACCOUNT_TYPE_SAVINGS,
			gomoneypbv1.AccountType_ACCOUNT_TYPE_BROKERAGE,
			gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE:
			return true
		default:
			return false
		}
	} // true

	switch {
	case isDebitNormal(sourceAcc.Type):
		// Debit-normal credit => DECREASE => expect negative input
		if tx.SourceAmountInBaseCurrency.Decimal.IsPositive() {
			return errors.New("source (debit-normal) must be negative for a credit")
		}
	default:
		// Credit-normal credit => INCREASE => expect positive input
		if !tx.SourceAmountInBaseCurrency.Decimal.IsPositive() {
			return errors.New("source (credit-normal) must be positive for a credit")
		}
	}

	isDebitFn := func(t gomoneypbv1.AccountType, amt decimal.Decimal) bool {
		if isDebitNormal(t) {
			return amt.IsPositive() // debit-normal: + => debit, - => credit
		}
		return amt.IsNegative() // credit-normal: - => debit, + => credit
	}

	isDebit := isDebitFn(sourceAcc.Type, baseAmount)

	entiries := []*database.DoubleEntry{
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

	return nil
}
