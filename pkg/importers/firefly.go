package importers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	chunkSize = 5000
)

type FireflyImporter struct {
	transactionService TransactionSvc
	currencyConverter  CurrencyConverterSvc
}

func NewFireflyImporter(
	txSvc TransactionSvc,
	converter CurrencyConverterSvc,
) *FireflyImporter {
	return &FireflyImporter{
		transactionService: txSvc,
		currencyConverter:  converter,
	}
}

type ImportRequest struct {
	Data            []byte
	Accounts        []*database.Account
	Tags            map[string]*database.Tag
	Categories      map[string]*database.Category
	SkipRules       bool
	TreatDatesAsUtc bool
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

func (f *FireflyImporter) Import(
	ctx context.Context,
	req *ImportRequest,
) (*importv1.ImportTransactionsResponse, error) {
	reader := csv.NewReader(bytes.NewBuffer(req.Data))
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
		accountMap[acc.Name] = acc
	}

	for _, record := range records {
		operationType := record[headerMap["type"]]
		amount := record[headerMap["amount"]]
		foreignAmount := record[headerMap["foreign_amount"]]
		currencyCode := record[headerMap["currency_code"]]
		foreignCurrencyCode := record[headerMap["foreign_currency_code"]]
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
			Notes:                   notes,
			Extra:                   make(map[string]string),
			TagIds:                  nil, // mapped
			TransactionDate:         timestamppb.New(parsedDate.UTC()),
			Title:                   description,
			Transaction:             nil,
			InternalReferenceNumber: &key,
			SkipRules:               req.SkipRules,
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
				SourceAmount:    amountParsed.Neg().String(),
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

				expense.FxSourceCurrency = &foreignCurrencyCode
				expense.FxSourceAmount = lo.ToPtr(foreignAmountParsed.Abs().Mul(decimal.NewFromInt(-1)).String())
			}

			secondAccResp, secAccErr := f.getSecondAccount(ctx, &GetSecondaryAccountRequest{
				InitialAmount:     amountParsed,
				InitialCurrency:   currencyCode,
				Accounts:          accountMap,
				TargetAccountName: destinationName,
				TransactionType:   gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			})
			if secAccErr != nil {
				return nil, errors.Wrapf(secAccErr, "failed to get secondary account for transaction: %s", key)
			}

			expense.DestinationAccountId = secondAccResp.SecondaryAccount.ID
			expense.DestinationAmount = secondAccResp.SecondaryAmount.Abs().String()
			expense.DestinationCurrency = secondAccResp.SecondaryAccount.Currency

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Expense{
				Expense: expense,
			}
		case "Opening balance", "Deposit":
			if sourceType == "Debt" {
				sourceAccount, ok := accountMap[sourceName]
				if !ok {
					return nil, errors.Errorf("source account not found: %s", sourceName)
				}

				if sourceAccount.Currency != currencyCode {
					return nil, errors.Errorf(
						"source account currency %s does not match transaction currency %s for journal %s",
						sourceAccount.Currency,
						currencyCode,
						journalID,
					)
				}

				expense := &transactionsv1.Expense{
					SourceAmount:    amountParsed.Neg().String(),
					SourceCurrency:  currencyCode,
					SourceAccountId: sourceAccount.ID,
				}

				secondAccResp, secAccErr := f.getSecondAccount(ctx, &GetSecondaryAccountRequest{
					InitialAmount:     amountParsed,
					InitialCurrency:   currencyCode,
					Accounts:          accountMap,
					TargetAccountName: destinationName, // most likely will be always initial balance account
					TransactionType:   gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
				})
				if secAccErr != nil {
					return nil, errors.Wrapf(secAccErr, "failed to get secondary account for transaction: %s", key)
				}

				expense.DestinationAccountId = secondAccResp.SecondaryAccount.ID
				expense.DestinationAmount = secondAccResp.SecondaryAmount.Abs().String()
				expense.DestinationCurrency = secondAccResp.SecondaryAccount.Currency

				targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Expense{
					Expense: expense,
				}
			} else {
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

				secondAccResp, secAccErr := f.getSecondAccount(ctx, &GetSecondaryAccountRequest{
					InitialAmount:     amountParsed,
					InitialCurrency:   currencyCode,
					Accounts:          accountMap,
					TargetAccountName: sourceName,
					TransactionType:   gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
				})
				if secAccErr != nil {
					return nil, errors.Wrapf(secAccErr, "failed to get secondary account for transaction: %s", key)
				}

				targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Income{
					Income: &transactionsv1.Income{
						SourceAccountId:      secondAccResp.SecondaryAccount.ID,
						DestinationAccountId: destAccount.ID,
						SourceAmount:         secondAccResp.SecondaryAmount.Neg().String(),
						DestinationAmount:    amountParsed.Abs().String(),
						SourceCurrency:       secondAccResp.SecondaryAccount.Currency,
						DestinationCurrency:  currencyCode,
					},
				}
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
				rec.Adjustment.DestinationAmount = amountParsed.Neg().String()

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

			if foreignCurrencyCode == "" {
				foreignCurrencyCode = currencyCode // assume same currency transfer in same currency
			}

			if currencyCode != foreignCurrencyCode && foreignAmount == "" {
				return nil, errors.Errorf(
					"foreign amount is required for currency conversion from %s to %s",
					currencyCode,
					foreignCurrencyCode,
				)
			}

			if foreignAmount == "" { // assuming same currency transfer
				foreignAmount = amount
			}

			foreignAmountParsed, err := decimal.NewFromString(foreignAmount)
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

			if destAccount.Currency != foreignCurrencyCode {
				return nil, errors.Errorf(
					"destination account currency %s does not match foreign transaction currency %s for journal %s",
					destAccount.Currency,
					foreignCurrencyCode,
					journalID,
				)
			}

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_TransferBetweenAccounts{
				TransferBetweenAccounts: &transactionsv1.TransferBetweenAccounts{
					SourceAccountId:      sourceAccount.ID,
					DestinationAccountId: destAccount.ID,
					SourceAmount:         amountParsed.Abs().Mul(decimal.NewFromInt(-1)).String(),
					DestinationAmount:    foreignAmountParsed.Abs().String(),
					SourceCurrency:       currencyCode,
					DestinationCurrency:  foreignCurrencyCode,
				},
			}
		default:
			return nil, errors.Errorf("unsupported operation type: %s", operationType)
		}

		newTxs[key] = targetTx
	}

	journalIDs := lo.Keys(newTxs)
	duplicateCount := 0

	for _, chunk := range lo.Chunk(journalIDs, chunkSize) {
		var existingRecords []string

		if err = database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).
			Model(&database.ImportDeduplication{}).
			Where("import_source = ?", f.Type().Number()).
			Where("key in ?", chunk).
			Pluck("key", &existingRecords).Error; err != nil {
			return nil, errors.Wrap(err, "failed to check existing transactions")
		}

		for _, record := range existingRecords {
			delete(newTxs, record)

			duplicateCount += 1
		}
	}

	if len(newTxs) == 0 {
		return &importv1.ImportTransactionsResponse{
			ImportedCount:  0,
			DuplicateCount: int32(duplicateCount),
		}, nil
	}

	var allTransactions []*transactions.BulkRequest
	for _, tx := range newTxs {
		allTransactions = append(allTransactions, &transactions.BulkRequest{
			Req: tx,
		})
	}

	sort.Slice(allTransactions, func(i, j int) bool {
		return allTransactions[i].Req.TransactionDate.AsTime().Before(allTransactions[j].Req.TransactionDate.AsTime())
	})

	tx := database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).Begin()
	defer tx.Rollback()
	ctx = database.WithContext(ctx, tx)

	transactionResp, transactionErr := f.transactionService.CreateBulkInternal(ctx, allTransactions, tx)
	if transactionErr != nil {
		return nil, errors.Wrap(transactionErr, "failed to create transactions")
	}

	var deduplicationRecords []*database.ImportDeduplication
	for _, record := range transactionResp {
		deduplicationRecords = append(deduplicationRecords, &database.ImportDeduplication{
			ImportSource:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			Key:           *record.Transaction.InternalReferenceNumber,
			CreatedAt:     time.Now(),
			TransactionID: record.Transaction.Id,
		})
	}

	for _, chunk := range lo.Chunk(deduplicationRecords, chunkSize) {
		if err = tx.Create(&chunk).Error; err != nil {
			return nil, errors.Wrap(err, "failed to create deduplication records")
		}
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return &importv1.ImportTransactionsResponse{
		ImportedCount:  int32(len(allTransactions)),
		DuplicateCount: int32(duplicateCount),
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

func (f *FireflyImporter) getSecondAccount(
	ctx context.Context,
	req *GetSecondaryAccountRequest,
) (*GetSecondaryAccountResponse, error) {
	secondaryAccount, ok := req.Accounts[req.TargetAccountName]
	if !ok {
		dest, err := f.getDefaultAccountForTransactionType(
			gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
			req.Accounts,
		)

		if err != nil {
			return nil, errors.Wrapf(err, "failed to get default account for transaction type: %s",
				gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE)
		}

		secondaryAccount = dest
	}

	finalAmount := req.InitialAmount.Abs()

	if secondaryAccount.Currency != req.InitialCurrency {
		converted, convertErr := f.currencyConverter.Convert(
			ctx,
			req.InitialCurrency,
			secondaryAccount.Currency,
			finalAmount,
		)
		if convertErr != nil {
			return nil, errors.Wrapf(convertErr,
				"failed to convert amount from %s to %s",
				req.InitialCurrency,
				secondaryAccount.Currency,
			)
		}

		finalAmount = converted
	}

	return &GetSecondaryAccountResponse{
		SecondaryAccount: secondaryAccount,
		SecondaryAmount:  finalAmount,
	}, nil
}

func (f *FireflyImporter) getDefaultAccountForTransactionType(
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
