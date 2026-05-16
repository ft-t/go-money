package importers

import (
	"bytes"
	"context"
	"encoding/csv"
	"regexp"
	"sort"
	"strings"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	mbankMinCols    = 5
	mbankDateLayout = "2006-01-02"
	mbankAnchorCol  = "Data operacji"
)

var (
	mbankBOM   = []byte{0xEF, 0xBB, 0xBF}
	mbankNRBRe = regexp.MustCompile(`\d{20,}`)
)

type Mbank struct {
	*BaseParser
}

func NewMbank(base *BaseParser) *Mbank {
	return &Mbank{
		BaseParser: base,
	}
}

func (m *Mbank) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_MBANK
}

func (m *Mbank) Parse(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	decodedFiles, err := m.DecodeFiles(req.Data)
	if err != nil {
		return nil, err
	}

	var allRecords []*Record

	for _, fileData := range decodedFiles {
		allRecords = append(allRecords, &Record{
			Data:    fileData,
			Message: &Message{},
		})
	}

	parsed, err := m.ParseMessages(ctx, allRecords)
	if err != nil {
		return nil, err
	}

	accountNumberToAccountMap, err := m.GetAccountMapByNumbers(req.Accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account map by numbers")
	}

	createRequests, err := m.ToCreateRequests(
		ctx,
		parsed,
		req.SkipRules,
		accountNumberToAccountMap,
		m.Type(),
	)
	if err != nil {
		return nil, err
	}

	return &ParseResponse{
		CreateRequests: createRequests,
	}, nil
}

func (m *Mbank) ParseMessages(
	_ context.Context,
	rawArr []*Record,
) ([]*Transaction, error) {
	var transactions []*Transaction

	for _, raw := range rawArr {
		reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(raw.Data, mbankBOM)))
		reader.Comma = ';'
		reader.FieldsPerRecord = -1
		reader.LazyQuotes = true

		rows, err := reader.ReadAll()
		if err != nil {
			transactions = append(transactions, &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    errors.Wrap(err, "failed to read csv"),
			})
			continue
		}

		anchorIdx := findMbankAnchorRow(rows)
		if anchorIdx < 0 {
			transactions = append(transactions, &Transaction{
				ID:              uuid.NewString(),
				OriginalMessage: raw.Message,
				ParsingError:    errors.New("header row not found"),
			})
			continue
		}

		accountNumber := findMbankAccountNumber(rows[:anchorIdx])

		for i := anchorIdx + 1; i < len(rows); i++ {
			row := rows[i]

			if len(row) < mbankMinCols {
				break
			}

			if strings.TrimSpace(row[0]) == "" {
				break
			}

			tx := m.parseRow(row, accountNumber, raw.Message)
			transactions = append(transactions, tx)
		}
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	return transactions, nil
}

func (m *Mbank) parseRow(
	row []string,
	accountNumber string,
	message *Message,
) *Transaction {
	tx := &Transaction{
		ID:              uuid.NewString(),
		OriginalMessage: message,
		Raw:             strings.Join(row, ";"),
	}

	dateStr := strings.TrimSpace(row[0])
	date, err := time.Parse(mbankDateLayout, dateStr)
	if err != nil {
		tx.ParsingError = errors.Wrapf(err, "failed to parse date: %s", dateStr)
		return tx
	}

	description := mbankCollapseSpaces(row[1])
	category := strings.TrimSpace(row[3])
	amountStr := strings.TrimSpace(row[4])

	amount, currency, err := parseMbankAmount(amountStr)
	if err != nil {
		tx.ParsingError = errors.Wrapf(err, "failed to parse amount: %s", amountStr)
		return tx
	}

	tx.Date = date
	tx.Description = description
	tx.OriginalTxType = category

	if tx.Description == "" {
		tx.Description = category
	}

	amountAbs := amount.Abs()

	if amount.IsNegative() {
		tx.Type = TransactionTypeExpense
		tx.SourceAccount = accountNumber
		tx.SourceAmount = amountAbs
		tx.SourceCurrency = currency
		tx.DestinationAmount = amountAbs
		tx.DestinationCurrency = currency
	} else {
		tx.Type = TransactionTypeIncome
		tx.DestinationAccount = accountNumber
		tx.DestinationAmount = amountAbs
		tx.DestinationCurrency = currency
		tx.SourceAmount = amountAbs
		tx.SourceCurrency = currency
	}

	tx.DeduplicationKeys = []string{
		strings.Join([]string{
			date.Format(time.RFC3339),
			accountNumber,
			amountStr,
			currency,
			description,
			category,
		}, "$$"),
	}

	return tx
}

func findMbankAnchorRow(rows [][]string) int {
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}

		if normalizeMbankHeaderCell(row[0]) == mbankAnchorCol {
			return i
		}
	}

	return -1
}

func normalizeMbankHeaderCell(s string) string {
	s = strings.TrimPrefix(s, string(mbankBOM))
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")

	return strings.TrimSpace(s)
}

func findMbankAccountNumber(headerRows [][]string) string {
	for _, row := range headerRows {
		for _, cell := range row {
			if match := mbankNRBRe.FindString(cell); match != "" {
				return match
			}
		}
	}

	return ""
}

func mbankCollapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func parseMbankAmount(raw string) (decimal.Decimal, string, error) {
	fields := strings.Fields(raw)
	if len(fields) < 2 {
		return decimal.Zero, "", errors.Newf("unexpected amount format: %q", raw)
	}

	currency := fields[len(fields)-1]
	numberPart := strings.ReplaceAll(strings.Join(fields[:len(fields)-1], ""), ",", ".")

	value, err := decimal.NewFromString(numberPart)
	if err != nil {
		return decimal.Zero, "", errors.Wrapf(err, "failed to parse number %q", numberPart)
	}

	return value, currency, nil
}
