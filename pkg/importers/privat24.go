package importers

import (
	"context"
	"sort"
	"strings"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
)

const (
	privat24MinCols       = 8
	privat24DateLayout    = "02.01.2006 15:04:05"
	privat24HeaderDateCol = "Дата"
)

type Privat24 struct {
	*BaseParser
}

func NewPrivat24(base *BaseParser) *Privat24 {
	return &Privat24{
		BaseParser: base,
	}
}

func (p *Privat24) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_PRIVATE_24
}

func (p *Privat24) Parse(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	decodedFiles, err := p.DecodeFiles(req.Data)
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

	parsed, err := p.ParseMessages(ctx, allRecords)
	if err != nil {
		return nil, err
	}

	accountNumberToAccountMap, err := p.GetAccountMapByNumbers(req.Accounts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account map by numbers")
	}

	createRequests, err := p.ToCreateRequests(
		ctx,
		parsed,
		req.SkipRules,
		accountNumberToAccountMap,
		p.Type(),
	)
	if err != nil {
		return nil, err
	}

	return &ParseResponse{
		CreateRequests: createRequests,
	}, nil
}

func (p *Privat24) ParseMessages(
	_ context.Context,
	rawArr []*Record,
) ([]*Transaction, error) {
	var transactions []*Transaction

	for _, raw := range rawArr {
		fileData, err := xlsx.OpenBinary(raw.Data)
		if err != nil {
			transactions = append(transactions, &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    errors.Wrap(err, "failed to open excel"),
			})
			continue
		}

		if len(fileData.Sheets) == 0 {
			transactions = append(transactions, &Transaction{
				ID:              uuid.NewString(),
				OriginalMessage: raw.Message,
				ParsingError:    errors.New("no sheets found"),
			})
			continue
		}

		sheet := fileData.Sheets[0]

		headerIdx := findPrivat24HeaderRow(sheet)
		if headerIdx < 0 {
			transactions = append(transactions, &Transaction{
				ID:              uuid.NewString(),
				OriginalMessage: raw.Message,
				ParsingError:    errors.New("header row not found"),
			})
			continue
		}

		for i := headerIdx + 1; i < len(sheet.Rows); i++ {
			row := sheet.Rows[i]

			if len(row.Cells) < privat24MinCols {
				continue
			}

			if strings.TrimSpace(row.Cells[0].String()) == "" {
				continue
			}

			tx := p.parseRow(row, raw.Message)
			transactions = append(transactions, tx)
		}
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	return transactions, nil
}

func (p *Privat24) parseRow(
	row *xlsx.Row,
	message *Message,
) *Transaction {
	tx := &Transaction{
		ID:              uuid.NewString(),
		OriginalMessage: message,
	}

	tx.Raw = privat24RowRaw(row.Cells)

	dateStr := strings.TrimSpace(row.Cells[0].String())
	date, err := time.Parse(privat24DateLayout, dateStr)
	if err != nil {
		tx.ParsingError = errors.Wrapf(err, "failed to parse date: %s", dateStr)
		return tx
	}

	category := strings.TrimSpace(row.Cells[1].String())
	card := normalizePrivat24Card(row.Cells[2].String())
	description := strings.TrimSpace(row.Cells[3].String())
	cardAmountStr := strings.TrimSpace(row.Cells[4].String())
	cardCurrency := strings.TrimSpace(row.Cells[5].String())
	txAmountStr := strings.TrimSpace(row.Cells[6].String())
	txCurrency := strings.TrimSpace(row.Cells[7].String())

	cardAmount, err := decimal.NewFromString(cardAmountStr)
	if err != nil {
		tx.ParsingError = errors.Wrapf(err, "failed to parse card amount: %s", cardAmountStr)
		return tx
	}

	txAmount, err := decimal.NewFromString(txAmountStr)
	if err != nil {
		tx.ParsingError = errors.Wrapf(err, "failed to parse tx amount: %s", txAmountStr)
		return tx
	}

	tx.Date = date
	tx.DateFromMessage = date.Format("15:04")
	tx.Description = description
	tx.OriginalTxType = category

	if description == "" {
		tx.Description = category
	}

	cardAmountAbs := cardAmount.Abs()
	txAmountAbs := txAmount.Abs()

	if cardAmount.IsNegative() {
		tx.Type = TransactionTypeExpense
		tx.SourceAccount = card
		tx.SourceAmount = cardAmountAbs
		tx.SourceCurrency = cardCurrency
		tx.DestinationAmount = txAmountAbs
		tx.DestinationCurrency = txCurrency
	} else {
		tx.Type = TransactionTypeIncome
		tx.DestinationAccount = card
		tx.DestinationAmount = cardAmountAbs
		tx.DestinationCurrency = cardCurrency
		tx.SourceAmount = txAmountAbs
		tx.SourceCurrency = txCurrency
	}

	tx.DeduplicationKeys = []string{
		strings.Join([]string{
			tx.Date.Format(time.RFC3339),
			card,
			cardAmountStr,
			cardCurrency,
			txAmountStr,
			txCurrency,
			description,
			category,
		}, "$$"),
	}

	return tx
}

func findPrivat24HeaderRow(sheet *xlsx.Sheet) int {
	for i, row := range sheet.Rows {
		if len(row.Cells) == 0 {
			continue
		}
		if strings.TrimSpace(row.Cells[0].String()) == privat24HeaderDateCol {
			return i
		}
	}
	return -1
}

func normalizePrivat24Card(raw string) string {
	return strings.ReplaceAll(strings.TrimSpace(raw), " ", "")
}

func privat24RowRaw(cells []*xlsx.Cell) string {
	parts := make([]string, 0, len(cells))
	for i, c := range cells {
		parts = append(parts, c.String())
		if i > 20 {
			break
		}
	}
	return strings.Join(parts, "-")
}
