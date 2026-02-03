package importers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/twmb/murmur3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BaseParser struct {
	currencyConverterSvc CurrencyConverterSvc
	transactionSvc       TransactionSvc
	mapperSvc            MapperSvc
}

func NewBaseParser(
	svc CurrencyConverterSvc,
	txSvc TransactionSvc,
	mapper MapperSvc,
) *BaseParser {
	return &BaseParser{
		currencyConverterSvc: svc,
		transactionSvc:       txSvc,
		mapperSvc:            mapper,
	}
}

func (b *BaseParser) toKey(tx *Transaction, importSource importv1.ImportSource) string {
	return fmt.Sprintf("%v_%x", importSource.String(), b.GenerateHash(tx.Raw))
}

func (b *BaseParser) ToCreateRequests(
	ctx context.Context,
	transactions []*Transaction,
	skipRules bool,
	accountMap map[string]*database.Account,
	importSource importv1.ImportSource,
) ([]*transactionsv1.CreateTransactionRequest, error) {
	var requests []*transactionsv1.CreateTransactionRequest

	for _, tx := range transactions {
		key := b.toKey(tx, importSource)

		newTx := &transactionsv1.CreateTransactionRequest{
			Notes:                    tx.Raw,
			Extra:                    make(map[string]string),
			TransactionDate:          timestamppb.New(tx.Date),
			Title:                    tx.Description,
			InternalReferenceNumbers: []string{key},
			SkipRules:                skipRules,
			CategoryId:               nil,
			Transaction:              nil,
		}
		for _, dup := range tx.DuplicateTransactions {
			newTx.InternalReferenceNumbers = append(newTx.InternalReferenceNumbers, b.toKey(dup, importSource))
		}

		if tx.ParsingError != nil {
			newTx.Extra["parsing_error"] = tx.ParsingError.Error()
			requests = append(requests, newTx)
			continue
		}

		switch tx.Type {
		case TransactionTypeIncome:
			sourceAccount, err := b.GetDefaultAccountAndAmount(
				ctx,
				&GetAccountRequest{
					InitialAmount:   tx.SourceAmount.Abs().Neg(),
					InitialCurrency: tx.SourceCurrency,
					Accounts:        accountMap,
					TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
				},
			)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get source account for income")
			}

			destinationAccount, err := b.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   tx.DestinationAmount.Abs(),
				InitialCurrency: tx.DestinationCurrency,
				Accounts:        accountMap,
				AccountName:     tx.DestinationAccount,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to get destination account for income")
			}

			newTx.Transaction = &transactionsv1.CreateTransactionRequest_Income{
				Income: &transactionsv1.Income{
					SourceAccountId: sourceAccount.Account.ID,
					SourceAmount:    sourceAccount.AmountInAccountCurrency.Abs().Neg().String(),
					SourceCurrency:  sourceAccount.Account.Currency,

					DestinationAccountId: destinationAccount.Account.ID,
					DestinationAmount:    destinationAccount.AmountInAccountCurrency.Abs().String(),
					DestinationCurrency:  destinationAccount.Account.Currency,
				},
			}
		case TransactionTypeExpense:
			// destination here is usually FX currency
			sourceAccount, err := b.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   tx.SourceAmount.Abs().Neg(),
				InitialCurrency: tx.SourceCurrency,
				Accounts:        accountMap,
				AccountName:     tx.SourceAccount,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to get source account for expense")
			}

			destinationAccount, err := b.GetDefaultAccountAndAmount(
				ctx,
				&GetAccountRequest{
					InitialAmount:   tx.DestinationAmount.Abs(),
					InitialCurrency: tx.DestinationCurrency,
					Accounts:        accountMap,
					TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				},
			)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get destination account for expense")
			}

			newTx.Transaction = &transactionsv1.CreateTransactionRequest_Expense{
				Expense: &transactionsv1.Expense{
					SourceAmount:         sourceAccount.AmountInAccountCurrency.Abs().Neg().String(),
					SourceCurrency:       sourceAccount.Account.Currency,
					SourceAccountId:      sourceAccount.Account.ID,
					FxSourceAmount:       lo.ToPtr(tx.DestinationAmount.Abs().Neg().String()),
					FxSourceCurrency:     &tx.DestinationCurrency,
					DestinationAccountId: destinationAccount.Account.ID,
					DestinationAmount:    destinationAccount.AmountInAccountCurrency.Abs().String(),
					DestinationCurrency:  destinationAccount.Account.Currency,
				},
			}
		case TransactionTypeInternalTransfer:
			sourceAccount, err := b.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   tx.SourceAmount.Abs().Neg(),
				InitialCurrency: tx.SourceCurrency,
				Accounts:        accountMap,
				AccountName:     tx.SourceAccount,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			})
			if err != nil {
				return nil, errors.Wrapf(err,
					"failed to get source account for internal transfer: description=%q, source_account=%q, dest_account=%q, raw=%q",
					tx.Description, tx.SourceAccount, tx.DestinationAccount, tx.Raw)
			}

			destinationAccount, err := b.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   tx.DestinationAmount.Abs(),
				InitialCurrency: tx.DestinationCurrency,
				Accounts:        accountMap,
				AccountName:     tx.DestinationAccount,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
			})
			if err != nil {
				return nil, errors.Wrapf(err,
					"failed to get destination account for internal transfer: description=%q, source_account=%q, dest_account=%q, raw=%q",
					tx.Description, tx.SourceAccount, tx.DestinationAccount, tx.Raw)
			}

			newTx.Transaction = &transactionsv1.CreateTransactionRequest_TransferBetweenAccounts{
				TransferBetweenAccounts: &transactionsv1.TransferBetweenAccounts{
					SourceAccountId:      sourceAccount.Account.ID,
					SourceAmount:         sourceAccount.AmountInAccountCurrency.Abs().Neg().String(),
					SourceCurrency:       sourceAccount.Account.Currency,
					DestinationAccountId: destinationAccount.Account.ID,
					DestinationAmount:    destinationAccount.AmountInAccountCurrency.Abs().String(),
					DestinationCurrency:  destinationAccount.Account.Currency,
				},
			}
		case TransactionTypeRemoteTransfer:
			// Remote transfer to external party - treat as expense
			sourceAccount, err := b.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   tx.SourceAmount.Abs().Neg(),
				InitialCurrency: tx.SourceCurrency,
				Accounts:        accountMap,
				AccountName:     tx.SourceAccount,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to get source account for remote transfer")
			}

			destinationAccount, err := b.GetDefaultAccountAndAmount(
				ctx,
				&GetAccountRequest{
					InitialAmount:   tx.DestinationAmount.Abs(),
					InitialCurrency: tx.DestinationCurrency,
					Accounts:        accountMap,
					TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				},
			)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get destination account for remote transfer")
			}

			newTx.Transaction = &transactionsv1.CreateTransactionRequest_Expense{
				Expense: &transactionsv1.Expense{
					SourceAccountId:      sourceAccount.Account.ID,
					SourceAmount:         sourceAccount.AmountInAccountCurrency.Abs().Neg().String(),
					SourceCurrency:       sourceAccount.Account.Currency,
					FxSourceAmount:       lo.ToPtr(tx.DestinationAmount.Abs().Neg().String()),
					FxSourceCurrency:     &tx.DestinationCurrency,
					DestinationAccountId: destinationAccount.Account.ID,
					DestinationAmount:    destinationAccount.AmountInAccountCurrency.Abs().String(),
					DestinationCurrency:  destinationAccount.Account.Currency,
				},
			}
		default:
			if newTx.Extra == nil {
				newTx.Extra = make(map[string]string)
			}
			newTx.Extra["unknown_type"] = fmt.Sprintf("%d", tx.Type)
		}

		requests = append(requests, newTx)
	}

	return requests, nil
}

func (b *BaseParser) GetAccountAndAmount(
	ctx context.Context,
	req *GetAccountRequest,
) (*GetSecondaryAccountResponse, error) {
	account, ok := req.Accounts[req.AccountName]
	if !ok {
		if req.TransactionType == gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS {
			return nil, errors.Errorf(
				"account not found for internal transfer: account_name=%q, available_accounts=%v",
				req.AccountName,
				b.getAccountNames(req.Accounts),
			)
		}

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
		converted, convertErr := b.currencyConverterSvc.Convert(
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
		converted, convertErr := b.currencyConverterSvc.Convert(
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

func (b *BaseParser) getAccountNames(accounts map[string]*database.Account) []string {
	names := make([]string, 0, len(accounts))
	for name := range accounts {
		names = append(names, name)
	}
	return names
}

func (b *BaseParser) GenerateHash(input string) string {
	return strconv.FormatUint(murmur3.Sum64([]byte(input)), 10)
}

func (b *BaseParser) GetAccountMapByNumbers(
	accounts []*database.Account,
) (map[string]*database.Account, error) {
	accountNumberToAccountMap := map[string]*database.Account{}

	for _, acc := range accounts {
		for _, num := range strings.Split(acc.AccountNumber, ",") {
			num = strings.TrimSpace(num)
			if num == "" {
				num = uuid.NewString() // fallback to ensure all accounts are passed
			}

			if _, exists := accountNumberToAccountMap[num]; exists {
				return nil, errors.Newf("duplicate account number: %s", num)
			}

			accountNumberToAccountMap[num] = acc
		}
	}

	return accountNumberToAccountMap, nil
}

func toLines(input string) []string {
	input = strings.ReplaceAll(input, "\r\n", "\n")

	return strings.Split(input, "\n")
}

func (b *BaseParser) DecodeFiles(
	records []string,
) ([][]byte, error) {
	var results [][]byte

	for _, record := range records {
		decoded, err := base64.StdEncoding.DecodeString(record)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode file content")
		}

		results = append(results, decoded)
	}

	return results, nil
}

func stripAccountPrefix(account string) string {
	account = strings.ToLower(account)
	var accountStriped strings.Builder

	for idx, l := range account {
		if !unicode.IsLetter(l) && idx == 0 {
			return account
		}

		if unicode.IsLetter(l) {
			continue
		}
		accountStriped.WriteRune(l)
	}

	return accountStriped.String()
}
