package importers

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
)

type Revolut struct {
	*BaseParser
}

func NewRevolut(base *BaseParser) *Revolut {
	return &Revolut{
		BaseParser: base,
	}
}

func (r *Revolut) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_REVOLUT
}

func (r *Revolut) AccountName(input string) string {
	return fmt.Sprintf("revolut_%s", input)
}

func (r *Revolut) Parse(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	decodedFiles, err := r.DecodeFiles(req.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode files")
	}

	var allRecords []*Record

	for _, fileData := range decodedFiles {
		records, err := r.splitCsv(ctx, fileData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to split csv")
		}
		allRecords = append(allRecords, records...)
	}

	parsed, err := r.ParseMessages(ctx, allRecords)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse messages")
	}

	accountNumberToAccountMap, err := r.GetAccountMapByNumbers(req.Accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account map by numbers")
	}

	createRequests, err := r.ToCreateRequests(
		ctx,
		parsed,
		req.SkipRules,
		accountNumberToAccountMap,
		r.Type(),
	)
	if err != nil {
		return nil, err
	}

	return &ParseResponse{
		CreateRequests: createRequests,
	}, nil
}

func (r *Revolut) splitCsv(
	_ context.Context,
	data []byte,
) ([]*Record, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1

	linesData, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(linesData) == 0 || len(linesData) == 1 {
		return nil, errors.New("empty file")
	}

	headerIndex := 1

	var records []*Record
	for i := headerIndex; i < len(linesData); i++ {
		targetLines := linesData[i : i+1]
		if len(targetLines) == 0 {
			break
		}

		if len(targetLines[0]) == 0 || targetLines[0][0] == "" {
			break
		}

		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		if err = writer.WriteAll(targetLines); err != nil {
			return nil, err
		}

		writer.Flush()

		records = append(records, &Record{
			Data: buf.Bytes(),
		})
	}

	return records, nil
}

func (r *Revolut) ParseMessages(
	ctx context.Context,
	rawArr []*Record,
) ([]*Transaction, error) {
	var transactions []*Transaction

	for _, raw := range rawArr {
		reader := csv.NewReader(bytes.NewReader(raw.Data))
		reader.FieldsPerRecord = -1

		linesData, err := reader.ReadAll()
		if err != nil {
			tx := &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    err,
			}
			transactions = append(transactions, tx)
			continue
		}

		if len(linesData) == 0 {
			tx := &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    errors.New("empty CSV data"),
			}
			transactions = append(transactions, tx)
			continue
		}

		startIndex := 0
		if len(linesData) > 1 && (linesData[0][0] == "Type" || linesData[0][0] == "\ufeffType" || strings.HasPrefix(linesData[0][0], "Type")) {
			startIndex = 1
		}

		for i := startIndex; i < len(linesData); i++ {
			if len(linesData[i]) == 0 || linesData[i][0] == "" {
				continue
			}

			tx := &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
			}
			transactions = append(transactions, tx)

			parsingErr := r.parseTransaction(tx, linesData[i])
			if parsingErr != nil {
				tx.ParsingError = parsingErr
				continue
			}
		}
	}

	return r.adjustAmount(r.merge(ctx, transactions)), nil
}

func (r *Revolut) adjustAmount(
	transactions []*Transaction,
) []*Transaction {
	for _, tx := range transactions {
		tx.SourceAmount = tx.SourceAmount.Abs()
		tx.DestinationAmount = tx.DestinationAmount.Abs()
	}

	return transactions
}

func (r *Revolut) merge(
	_ context.Context,
	transactions []*Transaction,
) []*Transaction {
	var finalTransactions []*Transaction
	var duplicates []*Transaction

	for _, tx := range transactions {
		if tx.Type != TransactionTypeInternalTransfer {
			finalTransactions = append(finalTransactions, tx)
			continue
		}

		if lo.Contains(duplicates, tx) {
			continue
		}

		merged := false
		for _, t := range transactions {
			if t == tx || t.Type != TransactionTypeInternalTransfer {
				continue
			}

			if lo.Contains(duplicates, t) {
				continue
			}

			if lo.Contains(finalTransactions, t) {
				continue
			}

			if t.Description == tx.Description && t.Date.Equal(tx.Date) {
				tx.DuplicateTransactions = append(tx.DuplicateTransactions, t)

				if tx.SourceAmount.LessThan(decimal.Zero) {
					tx.DestinationAmount = t.DestinationAmount
					tx.DestinationCurrency = t.DestinationCurrency
					tx.DestinationAccount = t.DestinationAccount
				} else {
					tx.SourceAmount = t.SourceAmount
					tx.SourceCurrency = t.SourceCurrency
					tx.SourceAccount = t.SourceAccount
				}

				duplicates = append(duplicates, t)
				finalTransactions = append(finalTransactions, tx)
				merged = true
				break
			}
		}

		if !merged {
			finalTransactions = append(finalTransactions, tx)
		}
	}

	return finalTransactions
}

func (r *Revolut) parseTransaction(
	tx *Transaction,
	data []string,
) error {
	if len(data) < 8 {
		return errors.Newf("expected len > 8, got %d", len(data))
	}

	invisibleChars := strings.TrimFunc(data[2], func(r rune) bool {
		return !unicode.IsGraphic(r)
	})

	operationType := data[0]

	operationTime, timeErr := time.Parse("2006-01-02 15:04:05", invisibleChars)
	if timeErr != nil {
		return errors.Wrapf(timeErr, "failed to parse operation time %s", data[2])
	}

	sourceAmount, err := decimal.NewFromString(data[5])
	if err != nil {
		return errors.Wrapf(err, "failed to parse source amount %s", data[5])
	}

	supportedStates := []string{
		"COMPLETED",
		"PENDING",
	}

	state := data[8]

	if !lo.Contains(supportedStates, state) {
		return errors.Newf("unsupported state %s", state)
	}

	tx.Type = TransactionTypeExpense
	tx.Date = operationTime

	tx.SourceAmount = sourceAmount
	tx.SourceCurrency = data[7]
	tx.SourceAccount = r.AccountName(tx.SourceCurrency)

	tx.Description = fmt.Sprintf("%s.%s", operationType, data[4])

	tx.DeduplicationKeys = []string{
		strings.Join([]string{
			operationType,
			data[2],
			data[4],
			data[5],
			data[7],
		}, "_"),
	}

	if operationType == "EXCHANGE" {
		tx.Type = TransactionTypeInternalTransfer

		if sourceAmount.GreaterThan(decimal.Zero) {
			tx.DestinationCurrency = tx.SourceCurrency
			tx.DestinationAmount = sourceAmount.Abs()
			tx.DestinationAccount = r.AccountName(tx.DestinationCurrency)

			tx.SourceCurrency = ""
			tx.SourceAmount = decimal.Zero
			tx.SourceAccount = ""
		}

		return nil
	}

	tx.SourceAmount = sourceAmount.Abs()

	if sourceAmount.GreaterThan(decimal.Zero) {
		return errors.New("income transactions not supported")
	}

	return nil
}
