package importers

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"strings"
	"time"
)

type FireflyImporter struct {
	transactionService TransactionSvc
}

func NewFireflyImporter(
	txSvc TransactionSvc,
) *FireflyImporter {
	return &FireflyImporter{
		transactionService: txSvc,
	}
}

type ImportRequest struct {
	Data      []byte
	Accounts  []*database.Account
	Tags      map[string]*database.Tag
	SkipRules bool
}

func (f *FireflyImporter) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_FIREFLY
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

	records = records[1:] // Skip header row
	mutable.Reverse(records)

	newTxs := map[string]*transactionsv1.CreateTransactionRequest{} // key is originalId
	accountMap := map[string]*database.Account{}
	for _, acc := range req.Accounts {
		accountMap[acc.Name] = acc
	}

	for _, record := range records {
		operationType := record[6]
		amount := record[7]
		foreignAmount := record[8]
		currencyCode := record[9]
		foreignCurrencyCode := record[10]
		description := record[11]
		date := record[12]
		sourceName := record[13]
		notes := record[24]
		sourceType := record[15]
		journalID := record[2]

		destinationName := record[16]
		destinationAccountType := record[18]

		possibleTags := f.toTag(record[20], "category:")
		possibleTags = append(possibleTags, f.toTag(record[21], "budget:")...)
		possibleTags = append(possibleTags, f.toTag(record[22], "bill:")...)

		for _, remoteTag := range strings.Split(record[23], ",") {
			possibleTags = append(possibleTags, f.toTag(remoteTag, "tag:")...)
		}

		parsedDate, err := time.Parse("2006-01-02T15:04:05-07:00", date)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse date: %s", date)
		}

		amountParsed, err := decimal.NewFromString(amount)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse amount: %s", amount)
		}

		targetTx := &transactionsv1.CreateTransactionRequest{
			Notes:                   notes,
			Extra:                   make(map[string]string),
			TagIds:                  nil, // mapped
			TransactionDate:         timestamppb.New(parsedDate.UTC()),
			Title:                   description,
			Transaction:             nil,
			InternalReferenceNumber: lo.ToPtr(fmt.Sprintf("firefly_%v", journalID)),
			SkipRules:               req.SkipRules,
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

			withdrawal := &transactionsv1.Withdrawal{
				SourceAccountId: sourceAccount.ID,
				SourceAmount:    amountParsed.Abs().Mul(decimal.NewFromInt(-1)).String(),
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

				withdrawal.ForeignCurrency = &foreignCurrencyCode
				withdrawal.ForeignAmount = lo.ToPtr(foreignAmountParsed.Abs().Mul(decimal.NewFromInt(-1)).String())
			}

			targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Withdrawal{
				Withdrawal: withdrawal,
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

				targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Withdrawal{
					Withdrawal: &transactionsv1.Withdrawal{
						SourceAmount:    amountParsed.Abs().Mul(decimal.NewFromInt(-1)).String(),
						SourceCurrency:  currencyCode,
						SourceAccountId: sourceAccount.ID,
					},
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

				targetTx.Transaction = &transactionsv1.CreateTransactionRequest_Deposit{
					Deposit: &transactionsv1.Deposit{
						DestinationCurrency:  currencyCode, // todo validate currency
						DestinationAccountId: destAccount.ID,
						DestinationAmount:    amountParsed.Abs().String(),
					},
				}
			}
		case "Reconciliation":
			rec := &transactionsv1.CreateTransactionRequest_Reconciliation{
				Reconciliation: &transactionsv1.Reconciliation{
					DestinationAmount:    "",
					DestinationCurrency:  currencyCode,
					DestinationAccountId: 0,
				},
			}
			if destinationAccountType == "Reconciliation account" {
				rec.Reconciliation.DestinationAmount = amountParsed.Abs().Mul(decimal.NewFromInt(-1)).String()

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

				rec.Reconciliation.DestinationAccountId = destAccount.ID
			} else {
				rec.Reconciliation.DestinationAmount = amountParsed.Abs().String()

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

				rec.Reconciliation.DestinationAccountId = destAccount.ID
			}

			targetTx.Transaction = rec
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

		newTxs[journalID] = targetTx
	}

	journalIDs := lo.Keys(newTxs)

	var existingRecords []string

	if err = database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).
		Model(&database.ImportDeduplication{}).
		Where("import_source = ?", f.Type().Number()).
		Where("key in ?", journalIDs).
		Pluck("key", &existingRecords).Error; err != nil {
		return nil, errors.Wrap(err, "failed to check existing transactions")
	}

	for _, record := range existingRecords {
		delete(newTxs, record)
	}

	if len(newTxs) == 0 {
		return &importv1.ImportTransactionsResponse{
			ImportedCount:  0,
			DuplicateCount: int32(len(existingRecords)),
		}, nil
	}

	var allTransactions []*transactionsv1.CreateTransactionRequest
	for _, tx := range newTxs {
		allTransactions = append(allTransactions, tx)
	}

	sort.Slice(allTransactions, func(i, j int) bool {
		return allTransactions[i].TransactionDate.AsTime().Before(allTransactions[j].TransactionDate.AsTime())
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

	if err = tx.Create(&deduplicationRecords).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create deduplication records")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return &importv1.ImportTransactionsResponse{
		ImportedCount:  int32(len(allTransactions)),
		DuplicateCount: int32(len(existingRecords)),
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
