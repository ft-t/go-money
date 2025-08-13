package double_entry

import (
	"context"
	"time"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	roundPlaces = 18
)

type DoubleEntryConfig struct {
	BaseCurrency string
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
	dbTx *gorm.DB,
	txs []*database.Transaction,
	accounts map[int32]*database.Account,
) error {
	var entries []*database.DoubleEntry

	for _, tx := range txs {
		sourceAccount, ok := accounts[tx.SourceAccountID]
		if !ok {
			return errors.New("source account not found for double entry transaction")
		}

		et, err := s.Calculate(ctx, &RecordRequest{
			Transaction:   tx,
			SourceAccount: sourceAccount,
		})

		if err != nil {
			return errors.Wrapf(err, "failed to record double entry for transaction ID %d", tx.ID)
		}

		entries = append(entries, et...)
	}

	if len(entries) == 0 {
		return nil
	}

	if err := dbTx.CreateInBatches(entries, 5000).Error; err != nil {
		return errors.Wrap(err, "failed to create double entry records in database")
	}

	return nil
}

func (s *DoubleEntryService) Calculate(
	_ context.Context,
	req *RecordRequest,
) ([]*database.DoubleEntry, error) {
	tx := req.Transaction

	sourceAcc := req.SourceAccount
	if sourceAcc == nil {
		return nil, errors.New("source account is required for double entry transactions")
	}

	if tx.SourceAccountID == 0 {
		return nil, errors.New("source_account_id is required for double entry transactions")
	}

	if tx.DestinationAccountID == 0 {
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

	baseAmount := tx.SourceAmountInBaseCurrency.Decimal

	isDebit := s.isDebit(sourceAcc.Type, baseAmount)

	entries := []*database.DoubleEntry{
		{
			TransactionID:        tx.ID,
			IsDebit:              isDebit,
			AccountID:            tx.SourceAccountID,
			BaseCurrency:         s.cfg.BaseCurrency,
			AmountInBaseCurrency: baseAmount.Abs(),
			CreatedAt:            time.Now().UTC(),
		},
		{
			TransactionID:        tx.ID,
			IsDebit:              !isDebit,
			AccountID:            tx.DestinationAccountID,
			BaseCurrency:         s.cfg.BaseCurrency,
			AmountInBaseCurrency: baseAmount.Abs(),
			CreatedAt:            time.Now().UTC(),
		},
	}

	return entries, nil
}

func (s *DoubleEntryService) isDebitNormal(accountType gomoneypbv1.AccountType) bool {
	switch accountType {
	case gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
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
