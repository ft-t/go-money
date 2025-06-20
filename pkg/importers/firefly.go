package importers

import (
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type FireflyImporter struct {
}

func NewFireflyImporter() *FireflyImporter {
	return &FireflyImporter{}
}

type ImportRequest struct {
	Data     []byte
	Accounts map[string]int32
}

func (f *FireflyImporter) Import(
	ctx context.Context,
	req *ImportRequest,
) error {
	reader := csv.NewReader(bytes.NewBuffer(req.Data))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return errors.Wrap(err, "failed to read CSV data")
	}

	if len(records) == 0 {
		return errors.New("no records found in CSV data")
	}

	records = records[1:] // Skip header row
	mutable.Reverse(records)

	for _, record := range records {
		operationType := record[6]
		amount := record[7]
		foreignAmount := record[8]
		currencyCode := record[9]
		foreignCurrencyCode := record[10]
		description := record[11]
		date := record[12]
		sourceName := record[13]
		destinationName := record[16]
		notes := record[24]

		parsedDate, err := time.Parse("2006-01-02T15:04:05-07:00", date)
		if err != nil {
			return errors.Wrapf(err, "failed to parse date: %s", date)
		}

		amountParsed, err := decimal.NewFromString(amount)
		if err != nil {
			return errors.Wrapf(err, "failed to parse amount: %s", amount)
		}

		targetTx := &transactionsv1.CreateTransactionRequest{
			Notes:           notes,
			Extra:           make(map[string]string), // todo
			LabelIds:        nil,                     // todo
			TransactionDate: timestamppb.New(parsedDate.UTC()),
			Title:           description,
			Transaction:     nil,
		}
		switch operationType {
		case "Withdrawal":
			sourceAccount, ok := req.Accounts[sourceName]
			if !ok {
				return errors.Errorf("source account not found: %s", sourceName)
			}

			withdrawal := &transactionsv1.Withdrawal{
				SourceAccountId: sourceAccount,
				SourceAmount:    amountParsed.Abs().Mul(decimal.NewFromInt(-1)).String(),
				SourceCurrency:  currencyCode,
			}

			if foreignAmount != "" {
				foreignAmountParsed, err := decimal.NewFromString(foreignAmount)
				if err != nil {
					return errors.Wrapf(err, "failed to parse foreign amount: %s", foreignAmount)
				}

				withdrawal.ForeignCurrency = &foreignCurrencyCode
				withdrawal.ForeignAmount = lo.ToPtr(foreignAmountParsed.Abs().Mul(decimal.NewFromInt(-1)).String())
			}

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Withdrawal{
				Withdrawal: withdrawal,
			}
		case "Opening balance", "Deposit":
			destAccount, ok := req.Accounts[destinationName]
			if !ok {
				return errors.Errorf("destination account not found: %s", destinationName)
			}
			// todo validate currency

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Deposit{
				Deposit: &transactionsv1.Deposit{
					DestinationAmount:    amountParsed.Abs().String(),
					DestinationCurrency:  currencyCode,
					DestinationAccountId: destAccount, // mapped account
				},
			}
		case "Transfer":
			sourceAccount, ok := req.Accounts[sourceName]
			if !ok {
				return errors.Errorf("source account not found: %s", destinationName)
			}

			destAccount, ok := req.Accounts[destinationName]
			if !ok {
				return errors.Errorf("destination account not found: %s", destinationName)
			}

			if foreignCurrencyCode == "" {
				foreignCurrencyCode = currencyCode // assume same currency transfer in same currency
			}

			if currencyCode != foreignCurrencyCode && foreignAmount == "" {
				return errors.Errorf("foreign amount is required for currency conversion from %s to %s", currencyCode, foreignCurrencyCode)
			}

			if foreignAmount == "" { // assuming same currency transfer
				foreignAmount = amount
			}

			foreignAmountParsed, err := decimal.NewFromString(foreignAmount)
			if err != nil {
				return errors.Wrapf(err, "failed to parse foreign amount: %s", foreignAmount)
			}

			// todo validate currency

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_TransferBetweenAccounts{
				TransferBetweenAccounts: &transactionsv1.TransferBetweenAccounts{
					SourceAccountId:      sourceAccount,
					DestinationAccountId: destAccount,
					SourceAmount:         amountParsed.Abs().Mul(decimal.NewFromInt(-1)).String(),
					DestinationAmount:    foreignAmountParsed.Abs().String(),
					SourceCurrency:       currencyCode,
					DestinationCurrency:  foreignCurrencyCode,
				},
			}
		default:
			return errors.Errorf("unsupported operation type: %s", operationType)
		}

		fmt.Println(targetTx)
	}

	return nil
}
