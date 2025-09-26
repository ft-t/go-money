package importers

import (
	"context"
	"strconv"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/twmb/murmur3"
)

type BaseParser struct {
	currencyConverter  CurrencyConverterSvc
	transactionService TransactionSvc
}

func NewBaseParser(
	svc CurrencyConverterSvc,
	txSvc TransactionSvc,
) *BaseParser {
	return &BaseParser{
		currencyConverter:  svc,
		transactionService: txSvc,
	}
}

func (b *BaseParser) GenerateHash(input string) string {
	return strconv.FormatUint(murmur3.Sum64([]byte(input)), 10)
}

func (b *BaseParser) GetAccountAndAmount(
	ctx context.Context,
	req *GetAccountRequest,
) (*GetSecondaryAccountResponse, error) {
	account, ok := req.Accounts[req.AccountName]
	if !ok {
		dest, err := b.getDefaultAccountForTransactionType(
			req.TransactionType,
			req.Accounts,
		)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to get default account for transaction type: %s",
				req.TransactionType)
		}

		account = dest
	}

	finalAmount := req.InitialAmount

	if account.Currency != req.InitialCurrency {
		converted, convertErr := b.currencyConverter.Convert(
			ctx,
			req.InitialCurrency,
			account.Currency,
			finalAmount,
		)
		if convertErr != nil {
			return nil, errors.Wrapf(convertErr,
				"failed to convert amount from %s to %s",
				req.InitialCurrency,
				account.Currency,
			)
		}

		finalAmount = converted
	}

	return &GetSecondaryAccountResponse{
		Account:                 account,
		AmountInAccountCurrency: finalAmount,
	}, nil
}

func (b *BaseParser) GetDefaultAccountAndAmount(
	ctx context.Context,
	req *GetAccountRequest,
) (*GetSecondaryAccountResponse, error) {
	account, err := b.getDefaultAccountForTransactionType(
		req.TransactionType,
		req.Accounts,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get default account for transaction type: %s",
			req.TransactionType)
	}

	finalAmount := req.InitialAmount

	if account.Currency != req.InitialCurrency {
		converted, convertErr := b.currencyConverter.Convert(
			ctx,
			req.InitialCurrency,
			account.Currency,
			finalAmount,
		)
		if convertErr != nil {
			return nil, errors.Wrapf(convertErr,
				"failed to convert amount from %s to %s",
				req.InitialCurrency,
				account.Currency,
			)
		}

		finalAmount = converted
	}

	return &GetSecondaryAccountResponse{
		Account:                 account,
		AmountInAccountCurrency: finalAmount,
	}, nil
}

func (b *BaseParser) getDefaultAccountForTransactionType(
	transactionType gomoneypbv1.TransactionType,
	accounts map[string]*database.Account,
) (*database.Account, error) {
	switch transactionType {
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE:
		for _, acc := range accounts {
			if acc.Type == gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE && acc.IsDefault() {
				return acc, nil
			}
		}
	case gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME:
		for _, acc := range accounts {
			if acc.Type == gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME && acc.IsDefault() {
				return acc, nil
			}
		}
	}

	return nil, errors.Errorf("unsupported transaction type for default account: %s", transactionType)
}
