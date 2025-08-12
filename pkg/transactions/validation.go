package transactions

import (
	"context"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ValidationService struct {
	cfg *ValidationServiceConfig
}

type ValidationServiceConfig struct {
	AccountSvc           AccountSvc
	ApplicableAccountSvc ApplicableAccountSvc
}

func NewValidationService(cfg *ValidationServiceConfig) *ValidationService {
	return &ValidationService{
		cfg: cfg,
	}
}

func (s *ValidationService) ValidateTransactions(
	ctx context.Context,
	dbTx *gorm.DB,
	txs []*database.Transaction,
) error {
	accounts, err := s.cfg.AccountSvc.GetAllAccounts(ctx) // todo filter by ids
	if err != nil {
		return errors.Wrap(err, "failed to get accounts")
	}

	applicableAccounts := s.cfg.ApplicableAccountSvc.GetApplicableAccounts(ctx, accounts)

	for _, createdTx := range txs {
		if err = s.ValidateTransactionData(ctx, dbTx, createdTx); err != nil {
			return errors.Wrapf(err, "failed to validate transaction")
		}

		if err = s.ValidateTransactionAccounts(ctx, applicableAccounts, createdTx); err != nil {
			return errors.Wrapf(err, "failed to validate transaction accounts")
		}
	}

	return nil
}

func (s *ValidationService) ensureCategoryExists(
	ctx context.Context,
	tx *database.Transaction,
) error {
	if tx.CategoryID == nil {
		return nil
	}

	return nil // todo
}

func (s *ValidationService) ValidateTransactionData(
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

func (s *ValidationService) ValidateTransactionAccounts(
	_ context.Context,
	possible map[gomoneypbv1.TransactionType]*PossibleAccount,
	tx *database.Transaction,
) error {
	possibleAccounts, ok := possible[tx.TransactionType]
	if !ok {
		return errors.Newf(
			"no possible accounts found for transaction type: %s",
			tx.TransactionType,
		)
	}

	if _, ok = possibleAccounts.SourceAccounts[*tx.SourceAccountID]; !ok {
		return errors.Newf(
			"source account %d is not applicable for transaction type: %s",
			*tx.SourceAccountID,
			tx.TransactionType,
		)
	}

	if _, ok = possibleAccounts.DestinationAccounts[*tx.DestinationAccountID]; !ok {
		return errors.Newf(
			"destination account %d is not applicable for transaction type: %s",
			*tx.DestinationAccountID,
			tx.TransactionType,
		)
	}

	return nil
}

func (s *ValidationService) validateWithdrawal(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE

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

		if tx.DestinationAccountID == nil {
			return errors.Newf(
				"destination_account_id is required for %s when destination_amount is provided",
				txType,
			)
		}

		accountsToCheck[*tx.DestinationAccountID] = tx.DestinationCurrency
	}

	if _, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, accountsToCheck); err != nil {
		return err
	}

	return nil
}

func (s *ValidationService) validateDeposit(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME

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

func (s *ValidationService) validateReconciliation(
	ctx context.Context,
	dbTx *gorm.DB,
	tx *database.Transaction,
) error {
	txType := gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT

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

func (s *ValidationService) validateTransferBetweenAccounts(
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

func (s *ValidationService) validateCurrency(
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

func (s *ValidationService) validateAmount(
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

func (s *ValidationService) ensureCurrencyExists(
	ctx context.Context,
	currency string,
) error {
	return nil // todo
}

func (s *ValidationService) ensureAccountsExistAndCurrencyCorrect(
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
