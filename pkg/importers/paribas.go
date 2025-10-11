package importers

import (
	"context"
	"sort"
	"strings"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
)

type Paribas struct {
	*BaseParser
	dataExtractors map[string]ParibasDataExtractor
}

func NewParibas(base *BaseParser) *Paribas {
	return &Paribas{
		BaseParser: base,
		dataExtractors: map[string]ParibasDataExtractor{
			"v1": ParibasDataExtractorV1{},
			"v2": ParibasDataExtractorV2{},
		},
	}
}

func (p *Paribas) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_BNP_PARIBAS_POLSKA
}

func (p *Paribas) Parse(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
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

func (p *Paribas) extractFromCellV1(cells []*xlsx.Cell) string {
	var values []string

	for i, c := range cells {
		values = append(values, c.String())

		if i > 20 {
			break
		}
	}

	return strings.Join(values, "-")
}

func (p *Paribas) getExtractor(cells []*xlsx.Cell) (ParibasDataExtractor, error) {
	if len(cells) < 6 {
		return nil, errors.New("row count is to short to determine the extractor type")
	}
	if cells[5].String() == "Nadawca" {
		return p.dataExtractors["v2"], nil
	}
	return p.dataExtractors["v1"], nil
}

func (p *Paribas) ParseMessages(
	ctx context.Context,
	rawArr []*Record,
) ([]*Transaction, error) {
	var transactions []*Transaction

	for _, raw := range rawArr {
		fileData, err := xlsx.OpenBinary(raw.Data)
		if err != nil {
			tx := &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    errors.Wrap(err, "failed to open excel"),
			}
			transactions = append(transactions, tx)
			continue
		}

		if len(fileData.Sheets) == 0 {
			tx := &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    errors.New("no sheets found"),
			}
			transactions = append(transactions, tx)
			continue
		}

		sheet := fileData.Sheets[0]

		if len(sheet.Rows) < 2 {
			tx := &Transaction{
				ID:              uuid.NewString(),
				Raw:             string(raw.Data),
				OriginalMessage: raw.Message,
				ParsingError:    errors.New("no rows found"),
			}
			transactions = append(transactions, tx)
			continue
		}

		extractor, err := p.getExtractor(sheet.Rows[0].Cells)
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

		for i := 1; i < len(sheet.Rows); i++ {
			row := sheet.Rows[i]

			if len(row.Cells) < 6 {
				continue
			}

			if zeroVal := row.Cells[0].String(); strings.TrimSpace(zeroVal) == "" {
				continue
			}

			tx := p.parseRow(ctx, row, extractor, raw.Message)
			transactions = append(transactions, tx)
		}
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	merged, err := p.merge(ctx, transactions)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

func (p *Paribas) parseRow(
	ctx context.Context,
	row *xlsx.Row,
	extractor ParibasDataExtractor,
	message *Message,
) *Transaction {
	tx := &Transaction{
		ID:                          uuid.NewString(),
		Type:                        0,
		SourceAmount:                decimal.Decimal{},
		SourceCurrency:              "",
		DestinationAmount:           decimal.Decimal{},
		DestinationCurrency:         "",
		Description:                 "",
		SourceAccount:               "",
		DestinationAccount:          "",
		DateFromMessage:             "",
		Raw:                         "",
		InternalTransferDirectionTo: false,
		DuplicateTransactions:       nil,
		OriginalMessage:             message,
	}

	tx.DeduplicationKeys = append(tx.DeduplicationKeys, p.extractFromCellV1(row.Cells))

	data, parseErr := extractor.Extract(ctx, row.Cells)
	if parseErr != nil {
		tx.ParsingError = parseErr
		tx.Raw = p.extractFromCellV1(row.Cells)
		return tx
	}

	tx.Date = data.Date
	tx.DateFromMessage = data.DateFromMessage
	tx.OriginalTxType = data.TransactionType
	tx.Raw = p.extractFromCellV1(row.Cells)
	tx.Description = data.Description

	var transactionType = data.TransactionType
	var currency = data.Currency
	var transactionCurrency = data.TransactionCurrency
	var amountParsed = data.Amount
	var kwotaParsed = data.TransactionAmount
	var account = data.Account
	var destinationAccount = data.DestinationAccount
	var executedAt = data.ExecutedAt
	var amount = data.AmountString
	var kwotaStr = data.TransactionAmountString

	skipExtraChecks := false

	switch transactionType {
	case "Transakcja kartą", "Transakcja BLIK", "Prowizje i opłaty",
		"Blokada środków", "Operacja gotówkowa", "Inne operacje", "Przelew podatkowy":
		if amountParsed.GreaterThan(decimal.Zero) {
			tx.Type = TransactionTypeIncome
			tx.SourceAmount = amountParsed.Abs()
			tx.SourceCurrency = currency
			tx.DestinationCurrency = transactionCurrency
			tx.DestinationAmount = kwotaParsed.Abs()
			tx.DestinationAccount = account
		} else {
			tx.Type = TransactionTypeExpense
			tx.SourceAccount = account
			tx.SourceAmount = amountParsed.Abs()
			tx.SourceCurrency = currency
			tx.DestinationCurrency = transactionCurrency
			tx.DestinationAmount = kwotaParsed.Abs()
			skipExtraChecks = true
		}
	case "Przelew zagraniczny":
		if kwotaParsed.IsPositive() {
			tx.Type = TransactionTypeIncome
			tx.DestinationAccount = account
			tx.DestinationAmount = amountParsed.Abs()
			tx.DestinationCurrency = currency

			tx.SourceCurrency = transactionCurrency
			tx.SourceAmount = kwotaParsed.Abs()
			tx.SourceAccount = destinationAccount
		} else {
			tx.Type = TransactionTypeExpense
			tx.SourceAccount = account
			tx.SourceAmount = amountParsed.Abs()
			tx.SourceCurrency = currency
			tx.DestinationCurrency = transactionCurrency
			tx.DestinationAmount = kwotaParsed.Abs()
			tx.DestinationAccount = destinationAccount
		}
	case "Przelew przychodzący":
		tx.Type = TransactionTypeIncome
		tx.DestinationAccount = account
		tx.DestinationAmount = amountParsed.Abs()
		tx.DestinationCurrency = currency

		tx.SourceCurrency = transactionCurrency
		tx.SourceAmount = kwotaParsed.Abs()
		tx.SourceAccount = destinationAccount

		skipExtraChecks = true
	case "Przelew wychodzący", "Przelew na telefon", "Spłata karty":
		tx.Type = TransactionTypeRemoteTransfer
		tx.DestinationAccount = destinationAccount
		tx.DestinationAmount = amountParsed.Abs()
		tx.DestinationCurrency = currency
		tx.SourceCurrency = currency
		tx.SourceAmount = amountParsed.Abs()
		tx.SourceAccount = account
	default:
		tx.ParsingError = errors.Newf("unknown transaction type: %s", transactionType)
		return tx
	}

	tx.DeduplicationKeys = append(tx.DeduplicationKeys,
		strings.Join([]string{
			tx.SourceCurrency,
			tx.DestinationCurrency,
			tx.SourceAccount,
			tx.DestinationAccount,
			tx.Date.Format("2006-01-02"),
			tx.SourceAmount.String(),
			tx.DestinationAmount.String(),
			tx.Description,
			tx.OriginalNadawcaName,
			tx.OriginalTxType,
			transactionType,
		}, "$$"),
	)

	if transactionType == "Blokada środków" && executedAt == "" {
		tx.ParsingError = errors.New("transaction is still pending. will skip from firefly for now")
		return tx
	}

	if !skipExtraChecks {
		if transactionCurrency != currency {
			tx.ParsingError = errors.Newf("currency mismatch: %s != %s", transactionCurrency, currency)
			return tx
		}

		if amount != kwotaStr {
			tx.ParsingError = errors.Newf("amount mismatch: %s != %s", amount, kwotaStr)
			return tx
		}
	}

	return tx
}

func (p *Paribas) merge(
	_ context.Context,
	transactions []*Transaction,
) ([]*Transaction, error) {
	var final []*Transaction

	var isPaymentByCard = func(tx *Transaction) bool {
		return tx.OriginalTxType == "Spłata karty"
	}

	var filteredTransactions []*Transaction

	for _, tx := range transactions {
		if tx.ParsingError != nil {
			final = append(final, tx)
			continue
		}

		if tx.Description == "Spłata karty" || tx.Description == "Card repayment" {
			final = append(final, tx)
		} else {
			filteredTransactions = append(filteredTransactions, tx)
		}
	}

	for _, tx := range filteredTransactions {

		isDuplicate := false

		for _, f := range final {
			if tx.OriginalTxType == "Prowizje i opłaty" {
				continue
			}

			if tx.Type == TransactionTypeExpense {
				continue
			}

			isCreditPaymentTx := isPaymentByCard(tx)

			if !isCreditPaymentTx && f.Description != tx.Description {
				continue
			}

			if isCreditPaymentTx && (f.Description != "Spłata karty" && f.Description != "Card repayment") {
				continue
			}

			if !f.Date.Equal(tx.Date) {
				continue
			}

			if len(f.DuplicateTransactions) > 0 {
				continue
			}

			if f.SourceCurrency != "" && tx.SourceCurrency != "" && f.DestinationCurrency != "" && tx.DestinationCurrency != "" {
				if !f.SourceAmount.Equal(tx.SourceAmount) &&
					tx.DestinationCurrency == f.DestinationCurrency && tx.SourceCurrency == f.SourceCurrency {
					continue
				}
			}

			if f.SourceAccount != "" && tx.SourceAccount != "" && f.DestinationAccount != "" && tx.DestinationAccount != "" {
				if f.SourceAccount == tx.SourceAccount && f.DestinationAccount == tx.DestinationAccount && tx.OriginalTxType == f.OriginalTxType {
					continue
				}
			}

			if f.SourceCurrency == "" {
				f.SourceCurrency = tx.SourceCurrency
			}
			if f.DestinationCurrency == "" {
				f.DestinationCurrency = tx.DestinationCurrency
			}
			if f.SourceAmount.IsZero() {
				f.SourceAmount = tx.SourceAmount
			}
			if f.DestinationAmount.IsZero() {
				f.DestinationAmount = tx.DestinationAmount
			}
			if f.SourceAccount == "" {
				f.SourceAccount = tx.SourceAccount
			}
			if f.DestinationAccount == "" {
				f.DestinationAccount = tx.DestinationAccount
			}

			if tx.OriginalTxType == "Przelew przychodzący" {
				f.DestinationAmount = tx.DestinationAmount
				f.DestinationCurrency = tx.DestinationCurrency
			}

			if tx.OriginalTxType == "Przelew wychodzący" {
				f.SourceAmount = tx.SourceAmount
				f.SourceCurrency = tx.SourceCurrency
			}

			f.Type = TransactionTypeInternalTransfer
			tx.Type = TransactionTypeInternalTransfer

			isDuplicate = true

			f.DuplicateTransactions = append(f.DuplicateTransactions, tx)
			break
		}

		if isDuplicate {
			continue
		}

		final = append(final, tx)

	}

	return final, nil
}
