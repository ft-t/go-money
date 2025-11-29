package importers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FireflyImporter struct {
	transactionService TransactionSvc
	currencyConverter  CurrencyConverterSvc
	*BaseParser
}

func NewFireflyImporter(
	txSvc TransactionSvc,
	converter CurrencyConverterSvc,
	parser *BaseParser,
) *FireflyImporter {
	return &FireflyImporter{
		transactionService: txSvc,
		currencyConverter:  converter,
		BaseParser:         parser,
	}
}

func (f *FireflyImporter) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_FIREFLY
}

func (f *FireflyImporter) ParseDate(
	date string,
	treatAsUtc bool,
) (time.Time, error) {
	parsedDate, err := time.Parse("2006-01-02T15:04:05-07:00", date)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "failed to parse date: %s", date)
	}

	if treatAsUtc {
		parsedDate = time.Date(
			parsedDate.Year(),
			parsedDate.Month(),
			parsedDate.Day(),
			parsedDate.Hour(),
			parsedDate.Minute(),
			parsedDate.Second(),
			parsedDate.Nanosecond(),
			time.UTC,
		)
	}

	return parsedDate, nil
}

func (f *FireflyImporter) Parse(
	ctx context.Context,
	req *ParseRequest,
) (*ParseResponse, error) {
	decodedFiles, err := f.DecodeFiles(req.Data)
	if err != nil {
		return nil, err
	}

	finalResp := &ParseResponse{}

	for index, file := range decodedFiles {
		resp, fileErr := f.ParseSingleFile(ctx, req, file)

		if fileErr != nil {
			return nil, errors.Wrapf(fileErr, "failed to parse file at index %d", index)
		}

		finalResp.CreateRequests = append(finalResp.CreateRequests, resp.CreateRequests...)
	}

	return finalResp, nil
}

func (f *FireflyImporter) ParseSingleFile(
	ctx context.Context,
	req *ParseRequest,
	content []byte,
) (*ParseResponse, error) {
	reader := csv.NewReader(bytes.NewReader(content))
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read CSV data")
	}

	if len(records) == 0 {
		return nil, errors.New("no records found in CSV data")
	}

	headerMap := map[string]int{}
	for i, header := range records[0] {
		headerMap[header] = i
	}

	records = records[1:] // Skip header row
	mutable.Reverse(records)

	newTxs := map[string]*transactionsv1.CreateTransactionRequest{} // key is originalId
	accountMap := map[string]*database.Account{}
	for _, acc := range req.Accounts {
		if _, ok := accountMap[acc.Name]; ok {
			return nil, errors.Errorf("duplicate account found: %s. Please rename account, and reimport both account and transactions", acc.Name)
		}
		accountMap[acc.Name] = acc
	}

	for _, record := range records {
		operationType := record[headerMap["type"]]
		amount := record[headerMap["amount"]]
		foreignAmount := record[headerMap["foreign_amount"]]
		currencyCode := record[headerMap["currency_code"]]
		destinationCurrencyCode := record[headerMap["foreign_currency_code"]]
		description := record[headerMap["description"]]
		date := record[headerMap["date"]]
		sourceName := record[headerMap["source_name"]]
		sourceType := record[headerMap["source_type"]]
		notes := record[headerMap["notes"]]
		journalID := record[headerMap["journal_id"]]

		destinationName := record[headerMap["destination_name"]]
		destinationAccountType := record[headerMap["destination_type"]]

		categoryName := record[headerMap["category"]]

		possibleTags := f.toTag(categoryName, "category:")
		possibleTags = append(possibleTags, f.toTag(record[headerMap["budget"]], "budget:")...)
		possibleTags = append(possibleTags, f.toTag(record[headerMap["bill"]], "bill:")...)

		for _, remoteTag := range strings.Split(record[headerMap["tags"]], ",") {
			possibleTags = append(possibleTags, f.toTag(remoteTag, "tag:")...)
		}

		parsedDate, err := f.ParseDate(date, req.TreatDatesAsUtc)
		if err != nil {
			return nil, err
		}

		amountParsed, err := decimal.NewFromString(amount)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse amount: %s", amount)
		}

		key := fmt.Sprintf("firefly_%v", journalID)

		targetTx := &transactionsv1.CreateTransactionRequest{
			Notes:                    notes,
			Extra:                    make(map[string]string),
			TagIds:                   nil, // mapped
			TransactionDate:          timestamppb.New(parsedDate.UTC()),
			Title:                    description,
			Transaction:              nil,
			InternalReferenceNumbers: []string{key},
			SkipRules:                req.SkipRules,
		}

		if v, ok := req.Categories[categoryName]; ok {
			targetTx.CategoryId = &v.ID
		}

		if len(req.Tags) > 0 {
			for _, tag := range lo.Uniq(possibleTags) {
				if _, ok := req.Tags[tag]; ok {
					targetTx.TagIds = append(targetTx.TagIds, req.Tags[tag].ID)
				}
			}
		}

		if operationType == "Withdrawal" && destinationAccountType == "Debt" { // debt repayment, so basically a transfer
			operationType = "Transfer"
		}

		switch operationType {
		case "Withdrawal":
			sourceAccount, ok := accountMap[sourceName]
			if !ok {
				return nil, errors.Errorf("source account not found: %s", sourceName)
			}

			expense := &transactionsv1.Expense{
				SourceAccountId: sourceAccount.ID,
				SourceAmount:    amountParsed.Abs().Neg().String(),
				SourceCurrency:  currencyCode,
			}

			if sourceAccount.Currency != currencyCode {
				return nil, errors.Newf("source account currency %s does not match transaction currency %s for journal %s", sourceAccount.Currency, currencyCode, journalID)
			}

			if foreignAmount != "" {
				foreignAmountParsed, err := decimal.NewFromString(foreignAmount)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse foreign amount: %s", foreignAmount)
				}

				expense.FxSourceCurrency = &destinationCurrencyCode
				expense.FxSourceAmount = lo.ToPtr(foreignAmountParsed.Abs().Mul(decimal.NewFromInt(-1)).String())
			}

			secondAccResp, secAccErr := f.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   amountParsed,
				InitialCurrency: currencyCode,
				Accounts:        accountMap,
				AccountName:     destinationName,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			})
			if secAccErr != nil {
				return nil, errors.Wrapf(secAccErr, "failed to get secondary account for transaction: %s", key)
			}

			expense.DestinationAccountId = secondAccResp.Account.ID
			expense.DestinationAmount = secondAccResp.AmountInAccountCurrency.Abs().String()
			expense.DestinationCurrency = secondAccResp.Account.Currency

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Expense{
				Expense: expense,
			}
		case "Opening balance":
			switch sourceType {
			case "Initial balance account", "":
				targetAccount, ok := accountMap[destinationName]
				if !ok {
					return nil, errors.Errorf("destination account not found: %s", destinationName)
				}
				targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Adjustment{
					Adjustment: &transactionsv1.Adjustment{
						DestinationAmount:    amountParsed.Abs().String(),
						DestinationCurrency:  currencyCode,
						DestinationAccountId: targetAccount.ID,
					},
				}
			case "Debt":
				targetAccount, ok := accountMap[sourceName]
				if !ok {
					return nil, errors.Errorf("source account not found: %s", sourceName)
				}
				targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Adjustment{
					Adjustment: &transactionsv1.Adjustment{
						DestinationAmount:    amountParsed.String(),
						DestinationCurrency:  currencyCode,
						DestinationAccountId: targetAccount.ID,
					},
				}
			default:
				return nil, errors.Errorf("unsupported source type for opening balance: %s", sourceType)
			}
		case "Deposit":

			destAccount, ok := accountMap[destinationName]
			if !ok {
				return nil, errors.Errorf("destination account not found: %s", destinationName)
			}

			if destAccount.Currency != currencyCode {
				return nil, errors.Errorf(
					"destination account currency %s does not match transaction currency %s for journal %s",
					destAccount.Currency,
					currencyCode,
					journalID,
				)
			}

			secondAccResp, secAccErr := f.GetAccountAndAmount(ctx, &GetAccountRequest{
				InitialAmount:   amountParsed,
				InitialCurrency: currencyCode,
				Accounts:        accountMap,
				AccountName:     sourceName,
				TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			})
			if secAccErr != nil {
				return nil, errors.Wrapf(secAccErr, "failed to get secondary account for transaction: %s", key)
			}

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Income{
				Income: &transactionsv1.Income{
					SourceAccountId:      secondAccResp.Account.ID,
					DestinationAccountId: destAccount.ID,
					SourceAmount:         secondAccResp.AmountInAccountCurrency.Abs().Neg().String(),
					DestinationAmount:    amountParsed.Abs().String(),
					SourceCurrency:       secondAccResp.Account.Currency,
					DestinationCurrency:  currencyCode,
				},
			}
		case "Reconciliation":
			rec := &transactionsv1.CreateTransactionRequest_Adjustment{
				Adjustment: &transactionsv1.Adjustment{
					DestinationAmount:    "",
					DestinationCurrency:  currencyCode,
					DestinationAccountId: 0,
				},
			}

			if destinationAccountType == "Reconciliation account" {
				rec.Adjustment.DestinationAmount = amountParsed.Abs().Neg().String()

				destAccount, ok := accountMap[sourceName] // yes, sourceName is used here as destination account
				if !ok {
					return nil, errors.Errorf("source account not found: %s", sourceName)
				}

				if destAccount.Currency != currencyCode {
					return nil, errors.Errorf(
						"destination account currency %s does not match transaction currency %s for journal %s",
						destAccount.Currency,
						currencyCode,
						journalID,
					)
				}

				rec.Adjustment.DestinationAccountId = destAccount.ID
			} else {
				rec.Adjustment.DestinationAmount = amountParsed.Abs().String()

				destAccount, ok := accountMap[destinationName]
				if !ok {
					return nil, errors.Errorf("destination account not found: %s", destinationName)
				}

				if destAccount.Currency != currencyCode {
					return nil, errors.Errorf(
						"destination account currency %s does not match transaction currency %s for journal %s",
						destAccount.Currency,
						currencyCode,
						journalID,
					)
				}

				rec.Adjustment.DestinationAccountId = destAccount.ID
			}

			targetTx.Transaction = rec // source account is handled in transaction service
		case "Transfer":
			sourceAccount, ok := accountMap[sourceName]
			if !ok {
				return nil, errors.Errorf("source account not found: %s", sourceName)
			}

			destAccount, ok := accountMap[destinationName]
			if !ok {
				return nil, errors.Errorf("destination account not found: %s", destinationName)
			}

			if destinationCurrencyCode == "" {
				destinationCurrencyCode = currencyCode // assume same currency transfer in same currency
			}

			if currencyCode != destinationCurrencyCode && foreignAmount == "" {
				return nil, errors.Errorf(
					"foreign amount is required for currency conversion from %s to %s",
					currencyCode,
					destinationCurrencyCode,
				)
			}

			if foreignAmount == "" { // assuming same currency transfer
				foreignAmount = amount
			}

			destinationAmountParsed, err := decimal.NewFromString(foreignAmount)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse foreign amount: %s", foreignAmount)
			}

			if sourceAccount.Currency != currencyCode {
				return nil, errors.Errorf(
					"source account currency %s does not match transaction currency %s for journal %s",
					sourceAccount.Currency,
					currencyCode,
					journalID,
				)
			}

			if destAccount.Currency != destinationCurrencyCode {
				destConverted, err := f.currencyConverter.Convert(ctx, currencyCode, destinationCurrencyCode, destinationAmountParsed)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to convert amount from %s to %s", currencyCode, destinationCurrencyCode)
				}

				destinationAmountParsed = destConverted
				//return nil, errors.Errorf(
				//	"destination account currency %s does not match foreign transaction currency %s for journal %s",
				//	destAccount.Currency,
				//	destinationCurrencyCode,
				//	journalID,
				//)
			}

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_TransferBetweenAccounts{
				TransferBetweenAccounts: &transactionsv1.TransferBetweenAccounts{
					SourceAccountId:      sourceAccount.ID,
					DestinationAccountId: destAccount.ID,
					SourceAmount:         amountParsed.Abs().Neg().String(),
					DestinationAmount:    destinationAmountParsed.Abs().String(),
					SourceCurrency:       currencyCode,
					DestinationCurrency:  destAccount.Currency,
				},
			}
		default:
			return nil, errors.Errorf("unsupported operation type: %s", operationType)
		}

		newTxs[key] = targetTx
	}

	return &ParseResponse{
		CreateRequests: lo.Values(newTxs),
	}, nil
}

func (f *FireflyImporter) toTag(
	input string,
	prefix string,
) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	return []string{
		input,
		prefix + input,
	}
}
