package transactions

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func (s *Service) ValidateTransaction(
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
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL:
		return s.validateWithdrawal(ctx, dbTx, tx)
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT:
		return s.validateDeposit(ctx, dbTx, tx)
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_RECONCILIATION:
		return s.validateReconciliation(ctx, dbTx, tx)
	default:
		return errors.Newf(
			"unsupported transaction type: %s",
			tx.TransactionType,
		)
	}
}

func (s *Service) validateWithdrawal(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL

	if lo.FromPtr(tx.SourceAccountID) == 0 {
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

	accountsToCheck := map[int32]string{
		*tx.SourceAccountID: tx.SourceCurrency,
	}

	if tx.FxSourceAmount.Valid { // foreign amount is optional
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

	if tx.DestinationAmount.Valid { // destination amount is optional
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

		accountsToCheck[*tx.DestinationAccountID] = tx.DestinationCurrency
	}

	if _, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, accountsToCheck); err != nil {
		return err
	}

	return nil
}

func (s *Service) validateDeposit(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT

	if lo.FromPtr(tx.DestinationAccountID) == 0 {
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

	if _, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, map[int32]string{
		*tx.DestinationAccountID: tx.DestinationCurrency,
	}); err != nil {
		return err
	}

	return nil
}

func (s *Service) validateReconciliation(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_RECONCILIATION

	if lo.FromPtr(tx.DestinationAccountID) == 0 {
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

	if _, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, map[int32]string{
		*tx.DestinationAccountID: tx.DestinationCurrency,
	}); err != nil {
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

	if lo.FromPtr(tx.SourceAccountID) == 0 {
		return errors.Newf(
			"source_account_id is required for %s",
			txType,
		)
	}
	if lo.FromPtr(tx.DestinationAccountID) == 0 {
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

	_, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, map[int32]string{
		*tx.SourceAccountID:      tx.SourceCurrency,
		*tx.DestinationAccountID: tx.DestinationCurrency,
	})
	if err != nil {
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
