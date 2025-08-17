package validation

import (
	"context"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Service struct {
	cfg *ServiceConfig
}

type ServiceConfig struct {
	ApplicableAccountSvc ApplicableAccountSvc
}

func NewValidationService(cfg *ServiceConfig) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) Validate(
	ctx context.Context,
	dbTx *gorm.DB,
	txs []*database.Transaction,
	accounts map[int32]*database.Account,
// req *Request,
) error {
	//accounts := req.Accounts
	//txs := req.Txs

	applicableAccounts := s.cfg.ApplicableAccountSvc.GetApplicableAccounts(ctx, lo.Values(accounts))

	for _, createdTx := range txs {
		if err := s.ValidateTransactionData(ctx, dbTx, createdTx); err != nil {
			return errors.Wrapf(err, "failed to validate transaction")
		}

		//if req.SkipAccountsValidation {
		if err := s.ValidateTransactionAccounts(ctx, applicableAccounts, createdTx); err != nil {
			return errors.Wrapf(err, "failed to validate transaction accounts")
		}
		//}

		if err := s.ensureCurrencyExists(ctx, createdTx.SourceCurrency); err != nil {
			return errors.Wrapf(err, "failed to ensure source currency exists: %s",
				createdTx.SourceCurrency)
		}

		if err := s.ensureCurrencyExists(ctx, createdTx.DestinationCurrency); err != nil {
			return errors.Wrapf(err, "failed to ensure destination currency exists: %s",
				createdTx.DestinationCurrency)
		}

		if err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, accounts, map[int32]string{
			createdTx.SourceAccountID:      createdTx.SourceCurrency,
			createdTx.DestinationAccountID: createdTx.DestinationCurrency,
		}); err != nil {
			return errors.Wrapf(err,
				"failed to ensure accounts exist and have correct currency for transaction id: %d",
				createdTx.ID,
			)
		}
	}

	return nil
}

func (s *Service) ensureCategoryExists(
	_ context.Context,
	tx *database.Transaction,
) error {
	if tx.CategoryID == nil {
		return nil
	}

	return nil // todo
}

func (s *Service) ValidateTransactionData(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	if err := s.ensureCategoryExists(ctx, tx); err != nil {
		return err
	}
	switch tx.TransactionType {
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS:
		return s.validateTransferBetweenAccounts(ctx, dbTx, tx)
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE:
		return s.validateWithdrawal(ctx, dbTx, tx)
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME:
		return s.validateDeposit(ctx, dbTx, tx)
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT:
		return s.validateReconciliation(ctx, dbTx, tx)
	default:
		return errors.Newf(
			"unsupported transaction type: %s",
			tx.TransactionType,
		)
	}
}

func (s *Service) ValidateTransactionAccounts(
	_ context.Context,
	possible map[gomoneypbv1.TransactionType]*transactions.PossibleAccount,
	tx *database.Transaction,
) error {
	possibleAccounts, ok := possible[tx.TransactionType]
	if !ok {
		return errors.Newf(
			"no possible accounts found for transaction type: %s",
			tx.TransactionType,
		)
	}

	if _, ok = possibleAccounts.SourceAccounts[tx.SourceAccountID]; !ok {
		return errors.Newf(
			"source account %d is not applicable for transaction type: %s",
			tx.SourceAccountID,
			tx.TransactionType,
		)
	}

	if _, ok = possibleAccounts.DestinationAccounts[tx.DestinationAccountID]; !ok {
		return errors.Newf(
			"destination account %d is not applicable for transaction type: %s",
			tx.DestinationAccountID,
			tx.TransactionType,
		)
	}

	return nil
}

func (s *Service) validateWithdrawal(
	_ context.Context,
	_ *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE

	if tx.SourceAccountID == 0 {
		return errors.Newf(
			"source_account_id is required for %s",
			txType,
		)
	}

	if err := s.validateCurrency(
		tx.SourceCurrency,
		txType,
		"source_currency",
	); err != nil {
		return err
	}

	if err := s.validateAmount(
		tx.SourceAmount,
		false,
		txType,
		"source_amount",
	); err != nil {
		return err
	}

	if tx.FxSourceAmount.Valid { // fx source amount is optional
		if err := s.validateAmount(
			tx.FxSourceAmount,
			false,
			txType,
			"fx_source_amount",
		); err != nil {
			return err
		}

		if err := s.validateCurrency(
			tx.FxSourceCurrency,
			txType,
			"fx_source_currency",
		); err != nil {
			return err
		}
	}

	if err := s.validateAmount(
		tx.DestinationAmount,
		true,
		txType,
		"destination_amount",
	); err != nil {
		return err
	}

	if err := s.validateCurrency(
		tx.DestinationCurrency,
		txType,
		"destination_currency",
	); err != nil {
		return err
	}

	if tx.DestinationAccountID == 0 {
		return errors.Newf(
			"destination_account_id is required for %s when destination_amount is provided",
			txType,
		)
	}

	return nil
}

func (s *Service) validateDeposit(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME

	if tx.DestinationAccountID == 0 {
		return errors.Newf(
			"destination_account_id is required for %s",
			txType,
		)
	}

	if err := s.validateCurrency(
		tx.DestinationCurrency,
		txType,
		"destination_currency",
	); err != nil {
		return err
	}

	if err := s.validateAmount(
		tx.DestinationAmount,
		true,
		txType,
		"destination_amount",
	); err != nil {
		return err
	}

	return nil
}

func (s *Service) validateReconciliation(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT

	if tx.DestinationAccountID == 0 {
		return errors.Newf(
			"destination_account_id is required for %s",
			txType,
		)
	}

	if !tx.DestinationAmount.Valid {
		return errors.Newf(
			"destination_amount is required for %s",
			txType,
		)
	}

	if err := s.validateCurrency(
		tx.DestinationCurrency,
		txType,
		"destination_currency",
	); err != nil {
		return err
	}

	return nil
}

func (s *Service) validateTransferBetweenAccounts(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS

	if tx.SourceAccountID == 0 {
		return errors.Newf(
			"source_account_id is required for %s",
			txType,
		)
	}
	if tx.DestinationAccountID == 0 {
		return errors.Newf(
			"destination_account_id is required for %s",
			txType,
		)
	}

	if err := s.validateAmount(
		tx.SourceAmount,
		false,
		txType,
		"source_amount",
	); err != nil {
		return err
	}

	if err := s.validateAmount(
		tx.DestinationAmount,
		true,
		txType,
		"destination_amount",
	); err != nil {
		return err
	}

	if err := s.validateCurrency(
		tx.SourceCurrency,
		txType,
		"source_currency",
	); err != nil {
		return err
	}

	if err := s.validateCurrency(
		tx.DestinationCurrency,
		txType,
		"destination_currency",
	); err != nil {
		return err
	}

	return nil
}

func (s *Service) validateCurrency(
	currency string,
	txType gomoneypbv1.TransactionType,
	fieldName string,
) error {
	if currency == "" {
		return errors.Newf(
			"%s is required for %s",
			fieldName,
			txType,
		)
	}

	return nil
}

func (s *Service) validateAmount(
	amount decimal.NullDecimal,
	shouldBePositive bool,
	txType gomoneypbv1.TransactionType,
	fieldName string,
) error {
	if !amount.Valid {
		return errors.Newf(
			"%s is required for %s",
			fieldName,
			txType,
		)
	}

	if shouldBePositive && amount.Decimal.IsNegative() {
		return errors.Newf(
			"%s must be positive for %s",
			fieldName,
			txType,
		)
	}

	if !shouldBePositive && amount.Decimal.IsPositive() {
		return errors.Newf(
			"%s must be negative for %s",
			fieldName,
			txType,
		)
	}

	return nil
}

func (s *Service) ensureCurrencyExists(
	ctx context.Context,
	currency string,
) error {
	return nil // todo
}

func (s *Service) ensureAccountsExistAndCurrencyCorrect(
	_ context.Context,
	_ *gorm.DB,
	accMap map[int32]*database.Account,
	expectedAccounts map[int32]string,
) error {
	accCurrencies := map[int32]string{}
	for _, acc := range accMap {
		accCurrencies[acc.ID] = acc.Currency
	}

	for id, expectedCurrency := range expectedAccounts {
		existingCurrency, ok := accCurrencies[id]

		if !ok {
			return errors.Newf("account with id %d not found", id)
		}

		if existingCurrency != expectedCurrency {
			return errors.Newf("account with id %d has currency %s, expected %s", id, existingCurrency, expectedCurrency)
		}
	}

	return nil
}
