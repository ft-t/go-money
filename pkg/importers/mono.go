package importers

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/hex"
	"strings"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	return importv1.ImportSource_IMPORT_SOURCE_MONOBANK
}

func (m *Mono) Parse(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	decodedFiles, err := m.DecodeFiles(req.Data)
	if err != nil {
		return nil, err
	}

	records, err := m.splitCsv(ctx, decodedFiles[0]) // single file support only
	if err != nil {
		return nil, err
	}

	parsed, err := m.parseMessages(ctx, records)
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
			return nil, errors.Wrapf(err, "failed to parse csv line: %s", string(rawCsv))
		}

		if len(linesData) == 0 {
			return nil, errors.New("empty csv line")
		}

		tx, _ := m.parseTransaction(linesData[0], raw.Message)

		if tx != nil {
			transactions = append(transactions, tx)
		}
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
