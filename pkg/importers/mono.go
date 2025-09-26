package importers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Mono struct {
	*BaseParser
}

func NewMono(base *BaseParser) *Mono {
	return &Mono{
		BaseParser: base,
	}
}

func (m *Mono) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_UNSPECIFIED
}

func (m *Mono) Import(
	ctx context.Context,
	req *ImportRequest,
) (*importv1.ImportTransactionsResponse, error) {
	records, err := m.splitCsv(ctx, req.Data)
	if err != nil {
		return nil, err
	}

	parsed, err := m.parseMessages(ctx, records)
	if err != nil {
		return nil, err
	}

	converted, err := m.toDbTransactions(ctx, req, parsed)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fmt.Println(converted)

	return nil, nil
}

func (m *Mono) splitCsv(
	_ context.Context,
	data []byte,
) ([]*Record, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1

	linesData, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(linesData) <= 1 {
		return nil, errors.New("empty file")
	}

	headerIndex := 1

	var records []*Record
	for i := headerIndex; i < len(linesData); i++ {
		if len(linesData[i]) == 0 || linesData[i][0] == "" {
			break
		}

		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		if err = writer.WriteAll(linesData[i : i+1]); err != nil {
			return nil, err
		}

		writer.Flush()

		records = append(records, &Record{
			Data:    []byte(hex.EncodeToString(buf.Bytes())),
			Message: &Message{},
		})
	}

	return records, nil
}

func (m *Mono) parseMessages(
	_ context.Context,
	rawArr []*Record,
) ([]*Transaction, error) {
	var transactions []*Transaction

	for _, raw := range rawArr {
		rawCsv, err := hex.DecodeString(string(raw.Data))
		if err != nil {
			return nil, err
		}

		reader := csv.NewReader(bytes.NewReader(rawCsv))
		reader.FieldsPerRecord = -1

		linesData, err := reader.ReadAll()
		if err != nil {
			transactions = append(transactions, &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    err,
			})
			continue
		}

		if len(linesData) == 0 {
			transactions = append(transactions, &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    errors.New("empty file"),
			})
			continue
		}

		tx, parsingErr := m.parseTransaction(linesData[0], raw.Message)
		if parsingErr != nil {
			tx.ParsingError = parsingErr
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (m *Mono) parseTransaction(
	data []string,
	message *Message,
) (*Transaction, error) {
	if len(data) < 8 {
		return &Transaction{
			ID:              uuid.NewString(),
			Raw:             strings.Join(data, ","),
			OriginalMessage: message,
		}, errors.Newf("expected len >= 8, got %d", len(data))
	}

	operationTime, timeErr := time.Parse("02.01.2006 15:04:05", strings.TrimSpace(data[0]))
	if timeErr != nil {
		return &Transaction{
			ID:              uuid.NewString(),
			Raw:             strings.Join(data, ","),
			OriginalMessage: message,
		}, errors.Wrapf(timeErr, "failed to parse operation time %s", data[0])
	}

	sourceAmount, err := decimal.NewFromString(data[3])
	if err != nil {
		return &Transaction{
			ID:              uuid.NewString(),
			Raw:             strings.Join(data, ","),
			OriginalMessage: message,
		}, errors.Wrapf(err, "failed to parse source amount %s", data[3])
	}

	if sourceAmount.GreaterThan(decimal.Zero) {
		return &Transaction{
			ID:              uuid.NewString(),
			Raw:             strings.Join(data, ","),
			OriginalMessage: message,
		}, errors.New("income transactions not supported")
	}

	destAmount, err := decimal.NewFromString(data[4])
	if err != nil {
		return &Transaction{
			ID:              uuid.NewString(),
			Raw:             strings.Join(data, ","),
			OriginalMessage: message,
		}, errors.Wrapf(err, "failed to parse dest amount %s", data[4])
	}

	tx := &Transaction{
		ID:                  uuid.NewString(),
		Type:                TransactionTypeExpense,
		Date:                operationTime,
		SourceAmount:        sourceAmount.Abs(),
		SourceCurrency:      "UAH",
		SourceAccount:       "UAH",
		DestinationAmount:   destAmount.Abs(),
		DestinationCurrency: data[5],
		Description:         data[1],
		Raw:                 strings.Join(data, ","),
		OriginalMessage:     message,
		DeduplicationKeys:   []string{strings.Join(data, "_")},
	}

	return tx, nil
}

func (m *Mono) toDbTransactions(
	ctx context.Context,
	req *ImportRequest,
	transactions []*Transaction,
) ([]*transactionsv1.CreateTransactionRequest, error) {
	var requests []*transactionsv1.CreateTransactionRequest

	for _, tx := range transactions {
		if tx.ParsingError != nil {
			continue
		}

		key := fmt.Sprintf("mono_%x", m.GenerateHash(tx.Raw))

		newTx := &transactionsv1.CreateTransactionRequest{
			Notes:                   tx.Raw,
			Extra:                   make(map[string]string),
			TransactionDate:         timestamppb.New(tx.Date),
			Title:                   tx.Description,
			ReferenceNumber:         nil,
			InternalReferenceNumber: &key,
			SkipRules:               req.SkipRules,
			CategoryId:              nil,
			Transaction:             nil,
		}

		accountNumberToAccountMap := map[string]*database.Account{}
		for _, acc := range req.Accounts {
			for _, num := range strings.Split(acc.AccountNumber, ",") {
				accountNumberToAccountMap[strings.TrimSpace(num)] = acc
			}
		}

		switch tx.Type {
		case TransactionTypeExpense:
			sourceAccount, err := m.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   tx.SourceAmount.Abs().Neg(),
				InitialCurrency: tx.SourceCurrency,
				Accounts:        accountNumberToAccountMap,
				AccountName:     tx.SourceAccount,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to get source account for expense")
			}

			destinationAccount, err := m.GetDefaultAccountAndAmount(
				ctx,
				&GetAccountRequest{
					InitialAmount:   tx.DestinationAmount.Abs(),
					InitialCurrency: tx.DestinationCurrency,
					Accounts:        accountNumberToAccountMap,
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
					FxSourceAmount:       lo.ToPtr(tx.DestinationAmount.Abs().String()),
					FxSourceCurrency:     &tx.DestinationCurrency,
					DestinationAccountId: destinationAccount.Account.ID,
					DestinationAmount:    destinationAccount.AmountInAccountCurrency.Abs().String(),
					DestinationCurrency:  destinationAccount.Account.Currency,
				},
			}
		default:
			return nil, errors.Newf("unsupported transaction type: %d", tx.Type)
		}

		requests = append(requests, newTx)
	}

	return requests, nil
}
