package importers

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"bytes"
	"context"
	"encoding/csv"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type FireflyImporter struct {
	transactionService *transactions.Service
}

func (f *FireflyImporter) Import(ctx context.Context, req *importv1.ImportTransactionsRequest) (*importv1.ImportTransactionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FireflyImporter) Type() importv1.ImportSource {
	//TODO implement me
	panic("implement me")
}

func NewFireflyImporter(
	txSvc *transactions.Service,
) *FireflyImporter {
	return &FireflyImporter{
		transactionService: txSvc,
	}
}

type ImportRequest struct {
	Data     []byte
	Accounts []*database.Account
}

func (f *FireflyImporter) Importv2(
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

	var allTransactions []*transactionsv1.CreateTransactionRequest

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
		destinationName := record[16]
		notes := record[24]
		sourceType := record[15]
		journalID := record[2]

		destinationAccountType := record[18]
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
			sourceAccount, ok := accountMap[sourceName]
			if !ok {
				return errors.Errorf("source account not found: %s", sourceName)
			}

			withdrawal := &transactionsv1.Withdrawal{
				SourceAccountId: sourceAccount.ID,
				SourceAmount:    amountParsed.Abs().Mul(decimal.NewFromInt(-1)).String(),
				SourceCurrency:  currencyCode,
			}

			if sourceAccount.Currency != currencyCode {
				return errors.Newf("source account currency %s does not match transaction currency %s for journal %s", sourceAccount.Currency, currencyCode, journalID)
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
			if sourceType == "Debt" {
				sourceAccount, ok := accountMap[sourceName]
				if !ok {
					return errors.Errorf("source account not found: %s", sourceName)
				}

				if sourceAccount.Currency != currencyCode {
					return errors.Errorf("source account currency %s does not match transaction currency %s for journal %s", sourceAccount.Currency, currencyCode, journalID)
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
					return errors.Errorf("destination account not found: %s", destinationName)
				}

				if destAccount.Currency != currencyCode {
					return errors.Errorf("destination account currency %s does not match transaction currency %s for journal %s", destAccount.Currency, currencyCode, journalID)
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
					return errors.Errorf("source account not found: %s", sourceName)
				}

				if destAccount.Currency != currencyCode {
					return errors.Errorf("destination account currency %s does not match transaction currency %s for journal %s", destAccount.Currency, currencyCode, journalID)
				}

				rec.Reconciliation.DestinationAccountId = destAccount.ID
			} else {
				rec.Reconciliation.DestinationAmount = amountParsed.Abs().String()

				destAccount, ok := accountMap[destinationName]
				if !ok {
					return errors.Errorf("destination account not found: %s", destinationName)
				}

				if destAccount.Currency != currencyCode {
					return errors.Errorf("destination account currency %s does not match transaction currency %s for journal %s", destAccount.Currency, currencyCode, journalID)
				}

				rec.Reconciliation.DestinationAccountId = destAccount.ID
			}

			targetTx.Transaction = rec
		case "Transfer":
			sourceAccount, ok := accountMap[sourceName]
			if !ok {
				return errors.Errorf("source account not found: %s", sourceName)
			}

			destAccount, ok := accountMap[destinationName]
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

			if sourceAccount.Currency != currencyCode {
				return errors.Errorf("source account currency %s does not match transaction currency %s for journal %s", sourceAccount.Currency, currencyCode, journalID)
			}

			if destAccount.Currency != foreignCurrencyCode {
				return errors.Errorf("destination account currency %s does not match foreign transaction currency %s for journal %s", destAccount.Currency, foreignCurrencyCode, journalID)
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
			return errors.Errorf("unsupported operation type: %s", operationType)
		}

		allTransactions = append(allTransactions, targetTx)
	}

	if _, err = f.transactionService.CreateBulk(ctx, allTransactions); err != nil {
		return errors.Wrap(err, "failed to create transactions")
	}

	return nil
}
