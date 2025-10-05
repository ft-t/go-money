package importers

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"github.com/cockroachdb/errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

const (
	simpleExpenseLinesCount  = 3
	remoteTransferLinesCount = 3
	incomeTransferLinesCount = 3
	partialRefundLinesCount  = 2
	creditPaymentLinesCount  = 2
)

const (
	unk                = "UNK"
	incomeTransferText = "зарахування переказу з картки через приват24"
)

var (
	simpleExpenseRegex        = regexp.MustCompile(`(\d+.?\d+)([A-Z]{3}) (.*)$`)
	balanceRegex              = regexp.MustCompile(`Бал\. .*(\w{3})`)
	remoteTransferRegex       = simpleExpenseRegex
	incomeTransferRegex       = simpleExpenseRegex
	internalTransferToRegex   = regexp.MustCompile(`(\d+.?\d+)([A-Z]{3}) (Переказ на свою карт[^ ]+ (?:(\d+\*\*\d+) )?(.*))$`)
	internalTransferFromRegex = regexp.MustCompile(`(\d+.?\d+)([A-Z]{3}) (Переказ зі своєї карт[^ ]+ (\*?\d+\*?\*?\d+) ?(.*)?)$`)
)

type Privat24 struct {
	*BaseParser
}

func (p *Privat24) Type() importv1.ImportSource {
	return importv1.ImportSource_IMPORT_SOURCE_PRIVATE_24
}

func NewPrivat24(
	base *BaseParser,
) *Privat24 {
	return &Privat24{
		BaseParser: base,
	}
}

func (p *Privat24) ExtractMessages(
	rawInput string,
) []string {
	rawInput = strings.ReplaceAll(rawInput, "\r\n", "\n")

	var builder strings.Builder

	var messages []string

	for _, r := range strings.Split(rawInput, "\n") {
		line := strings.TrimSpace(r)

		if line == "\n" || line == "" {
			messages = append(messages, builder.String())
			builder.Reset()
			continue // end of message
		}

		builder.WriteString(line)
		builder.WriteString("\n")
	}

	if builder.Len() != 0 {
		messages = append(messages, builder.String())
	}

	return messages
}

func (p *Privat24) Parse(
	ctx context.Context,
	req *ParseRequest,
) (*ParseResponse, error) {
	messages := p.ExtractMessages(req.Data[0])

	var records []*Record

	for _, message := range messages {
		lines := toLines(message)

		header := lines[0] // is header in format PrivatBank, [10/1/2025 9:50 AM]

		createdAt, err := p.ParseHeaderDate(header)
		if err != nil {
			zerolog.Ctx(ctx).Err(err).Str("input", header).Msg("failed to parse header date")

			return nil, errors.Wrap(err, "failed to parse header date")
		}

		records = append(records, &Record{
			Data: []byte(strings.Join(lines[1:], "\n")),
			Message: &Message{
				CreatedAt: createdAt,
			},
		})
	}

	parsed, err := p.ParseMessages(ctx, records)
	if err != nil {
		return nil, errors.WithStack(err)
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
	ctx context.Context,
	rawArr []*Record,
) ([]*Transaction, error) {
	var finalTx []*Transaction

	for _, rawItem := range rawArr {
		raw := string(rawItem.Data)
		lower := strings.ToLower(raw)
		lines := toLines(lower)

		if len(lines) == 0 {
			finalTx = append(finalTx, &Transaction{
				Raw:             raw,
				OriginalMessage: rawItem.Message,
				ParsingError:    errors.New("empty input"),
			})

			continue
		}

		if strings.HasSuffix(lines[0], "переказ зі своєї карти") { // external transfer to another bank
			remote, err := p.ParseRemoteTransfer(ctx, raw, rawItem.Message.CreatedAt)
			finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
			continue
		}

		if strings.Contains(lower, "переказ на свою карт") ||
			strings.Contains(lower, "переказ зі своєї карт") { // internal transfer
			remote, err := p.ParseInternalTransfer(ctx, raw, rawItem.Message.CreatedAt)

			finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
			continue
		}

		if strings.Contains(lower, "переказ через ") || strings.HasSuffix(lines[0], incomeTransferText) { // remote transfer
			if strings.Contains(lower, "відправник:") || strings.HasSuffix(lines[0], incomeTransferText) { // income
				remote, err := p.ParseIncomeTransfer(ctx, raw, rawItem.Message.CreatedAt)

				finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
				continue
			}

			remote, err := p.ParseRemoteTransfer(ctx, raw, rawItem.Message.CreatedAt)

			finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
			continue
		}

		if strings.HasSuffix(lines[0], "зарахування переказу на картку") || strings.Contains(lower, "повернення.") ||
			strings.HasSuffix(lines[0], "зарахування переказу через приват24 зі своєї картки") ||
			strings.Contains(lines[0], "зарахування переказу.") {
			remote, err := p.ParseIncomingCardTransfer(ctx, raw, rawItem.Message.CreatedAt)

			finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
			continue
		}

		if strings.HasSuffix(lines[0], "зарахування") {
			remote, err := p.ParsePartialRefund(ctx, raw, rawItem.Message.CreatedAt)

			finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
			continue
		}

		if len(lines) == 2 && strings.HasSuffix(lines[0], " списання") {
			remote, err := p.ParseCreditPayment(ctx, raw, rawItem.Message.CreatedAt)

			finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
			continue
		}

		remote, err := p.ParseSimpleExpense(ctx, raw, rawItem.Message.CreatedAt)

		finalTx = p.appendTxOrError(finalTx, remote, err, raw, rawItem)
		continue
	}

	merged, err := p.Merge(ctx, finalTx)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

func (p *Privat24) Merge(
	_ context.Context,
	messages []*Transaction,
) ([]*Transaction, error) {
	var finalTransactions []*Transaction

	for _, tx := range messages {
		// currently we have a transfer transaction, lets ensure that we dont have duplicates
		isDuplicate := false
		for _, f := range finalTransactions {
			if f.DateFromMessage == "" || tx.DateFromMessage == "" {
				continue // missing date, can not merge
			}

			fDate, err := time.Parse("15:04", f.DateFromMessage)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			txDate, err := time.Parse("15:04", tx.DateFromMessage)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			minDiff := math.Abs(fDate.Sub(txDate).Minutes())
			if minDiff > 5 { // diff > 5m
				continue // not our tx
			}

			if tx.InternalTransferDirectionTo && f.InternalTransferDirectionTo {
				continue // not our tx
			}

			if tx.SourceAccount == f.SourceAccount {
				if tx.DestinationAccount == unk {
					tx.DestinationAccount = f.DestinationAccount
				}
				if f.DestinationAccount == unk {
					f.DestinationAccount = tx.DestinationAccount
				}
			}

			if tx.SourceAccount == "" &&
				tx.Description == "Зарахування переказу через Приват24 зі своєї картки" &&
				f.Description == "Переказ на свою карту через Приват24" {

				tx.SourceAccount = f.SourceAccount
				tx.SourceAmount = f.SourceAmount
				tx.SourceCurrency = f.SourceCurrency

				f.DestinationAccount = tx.DestinationAccount
				f.DestinationAmount = tx.DestinationAmount
				f.DestinationCurrency = tx.DestinationCurrency

				f.Type = TransactionTypeInternalTransfer
				tx.Type = TransactionTypeInternalTransfer
			}

			if f.SourceAccount == "" && // revers of the above
				f.Description == "Зарахування переказу через Приват24 зі своєї картки" &&
				tx.Description == "Переказ на свою карту через Приват24" {
				f.SourceAccount = tx.SourceAccount
				f.SourceAmount = tx.SourceAmount
				f.SourceCurrency = tx.SourceCurrency

				tx.DestinationAccount = f.DestinationAccount
				tx.DestinationAmount = f.DestinationAmount
				tx.DestinationCurrency = f.DestinationCurrency

				f.Type = TransactionTypeInternalTransfer
				tx.Type = TransactionTypeInternalTransfer
			}

			if tx.DestinationAccount != "" &&
				tx.Description == "Зарахування переказу на картку" &&
				f.DestinationAccount == "" && f.Description == "Переказ зі своєї карти" &&
				f.SourceAccount != "" {
				f.DestinationAccount = tx.DestinationAccount
				tx.SourceAccount = f.SourceAccount

				f.Type = TransactionTypeInternalTransfer
				tx.Type = TransactionTypeInternalTransfer
			}

			if tx.DestinationAccount == "" &&
				tx.Description == "Переказ зі своєї карти" &&
				f.DestinationAccount != "" &&
				f.Description == "Зарахування переказу на картку" {
				tx.DestinationAccount = f.DestinationAccount
				f.SourceAccount = tx.SourceAccount

				f.Type = TransactionTypeInternalTransfer
				tx.Type = TransactionTypeInternalTransfer
			}

			if f.Type != TransactionTypeInternalTransfer && tx.Type != TransactionTypeInternalTransfer {
				continue
			}

			// privat fuck yourself. sometime card does not have first digit
			if strings.HasPrefix(tx.Description, "Переказ зі своєї картки") &&
				strings.HasPrefix(tx.SourceAccount, "*") &&
				strings.HasPrefix(f.Description, "Переказ на свою картку") &&
				strings.HasPrefix(f.DestinationAccount, "*") {
				tx.SourceAccount = f.SourceAccount
				f.DestinationAccount = tx.DestinationAccount
			}

			// reverse previous
			if strings.HasPrefix(f.Description, "Переказ зі своєї картки") &&
				strings.HasPrefix(f.SourceAccount, "*") &&
				strings.HasPrefix(tx.Description, "Переказ на свою картку") &&
				strings.HasPrefix(tx.DestinationAccount, "*") {
				f.SourceAccount = tx.SourceAccount
				tx.DestinationAccount = f.DestinationAccount
			}

			// privat really fuck you.
			if strings.HasPrefix(tx.Description, "Переказ на свою картку") &&
				strings.HasPrefix(tx.DestinationAccount, "*") &&
				strings.HasPrefix(f.Description, "Зарахування переказу") &&
				f.SourceAccount == "" {
				tx.DestinationAccount = f.DestinationAccount
				f.SourceAccount = tx.SourceAccount
			}

			// reverse
			if strings.HasPrefix(f.Description, "Переказ на свою картку") &&
				strings.HasPrefix(f.DestinationAccount, "*") &&
				strings.HasPrefix(tx.Description, "Зарахування переказу") &&
				tx.SourceAccount == "" {
				f.DestinationAccount = tx.DestinationAccount
				tx.SourceAccount = f.SourceAccount
			}

			if tx.DestinationAccount != f.DestinationAccount ||
				tx.SourceAccount != f.SourceAccount {
				continue
			}

			if f.DestinationCurrency == "" && tx.DestinationCurrency != "" {
				f.DestinationCurrency = tx.DestinationCurrency
			}
			if f.SourceCurrency == "" && tx.SourceCurrency != "" {
				f.SourceCurrency = tx.SourceCurrency
			}

			if f.DestinationAmount.Equal(decimal.Zero) && tx.DestinationAmount.GreaterThan(decimal.Zero) {
				f.DestinationAmount = tx.DestinationAmount
			}
			if f.SourceAmount.Equal(decimal.Zero) && tx.SourceAmount.GreaterThan(decimal.Zero) {
				f.SourceAmount = tx.SourceAmount
			}

			f.Type = TransactionTypeInternalTransfer
			tx.Type = TransactionTypeInternalTransfer

			// otherwise we have a duplicate
			f.DuplicateTransactions = append(f.DuplicateTransactions, tx)
			isDuplicate = true
		}

		if isDuplicate {
			continue
		}

		finalTransactions = append(finalTransactions, tx)
	}

	return finalTransactions, nil
}

func (p *Privat24) appendTxOrError(finalTx []*Transaction, tx *Transaction, err error, raw string, item *Record) []*Transaction {
	return appendTxOrError(finalTx, tx, err, raw, item)
}

func appendTxOrError(finalTx []*Transaction, tx *Transaction, err error, raw string, item *Record) []*Transaction {
	if !lo.IsNil(tx) {
		tx.OriginalMessage = item.Message
		finalTx = append(finalTx, tx)
	}

	if !lo.IsNil(err) {
		finalTx = append(finalTx, &Transaction{
			Raw:             raw,
			ParsingError:    err,
			OriginalMessage: item.Message,
		})
	}

	return finalTx
}

func (p *Privat24) ParseIncomingCardTransfer(
	_ context.Context,
	raw string,
	date time.Time,
) (*Transaction, error) {
	lines := toLines(raw)

	if len(lines) < partialRefundLinesCount {
		return nil, errors.Newf("expected %d lines, got %d", partialRefundLinesCount, len(lines))
	}

	matches := incomeTransferRegex.FindStringSubmatch(lines[0])
	if len(matches) != 4 {
		return nil, errors.Newf("expected 4 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) != 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	finalTx := &Transaction{
		ID:                  uuid.NewString(),
		Date:                date,
		DestinationCurrency: matches[2],
		Description:         matches[3],
		DestinationAmount:   amount,
		Type:                TransactionTypeIncome,
		DestinationAccount:  source[0],
		Raw:                 raw,
		DateFromMessage:     source[1],
	}

	return finalTx, nil
}

func (p *Privat24) ParsePartialRefund(
	_ context.Context,
	raw string,
	date time.Time,
) (*Transaction, error) {
	lines := toLines(raw)

	if len(lines) < partialRefundLinesCount {
		return nil, errors.Newf("expected %d lines, got %d", incomeTransferLinesCount, len(lines))
	}

	matches := incomeTransferRegex.FindStringSubmatch(lines[0])
	if len(matches) != 4 {
		return nil, errors.Newf("expected 4 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) != 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	finalTx := &Transaction{
		ID:                  uuid.NewString(),
		Date:                date,
		DestinationCurrency: matches[2],
		Description:         matches[3],
		DestinationAmount:   amount,
		Type:                TransactionTypeIncome,
		DestinationAccount:  source[0],
		Raw:                 raw,
		DateFromMessage:     source[1],
	}

	return finalTx, nil
}

func (p *Privat24) ParseIncomeTransfer(
	_ context.Context,
	raw string,
	date time.Time,
) (*Transaction, error) {
	lines := toLines(raw)

	if len(lines) < incomeTransferLinesCount {
		return nil, errors.Newf("expected %d lines, got %d", incomeTransferLinesCount, len(lines))
	}

	matches := incomeTransferRegex.FindStringSubmatch(lines[0])
	if len(matches) != 4 {
		return nil, errors.Newf("expected 4 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) != 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	finalTx := &Transaction{
		ID:                  uuid.NewString(),
		Date:                date,
		DestinationCurrency: matches[2],
		Description:         matches[3],
		DestinationAmount:   amount,
		Type:                TransactionTypeIncome,
		DestinationAccount:  source[0],
		Raw:                 raw,
		DateFromMessage:     source[1],
		SourceCurrency:      matches[2],
		SourceAmount:        amount.Abs(),
	}

	return finalTx, nil
}

func (p *Privat24) ParseInternalTransfer(
	ctx context.Context,
	raw string,
	date time.Time,
) (*Transaction, error) {
	lines := toLines(raw)

	isTo := strings.Contains(strings.ToLower(lines[0]), "переказ на свою карт")

	if isTo {
		return p.parseInternalTransferTo(ctx, raw, lines, date)
	}

	return p.parseInternalTransferFrom(ctx, raw, lines, date)
}

func (p *Privat24) parseInternalTransferFrom(
	_ context.Context,
	raw string,
	lines []string,
	date time.Time,
) (*Transaction, error) {
	matches := internalTransferFromRegex.FindStringSubmatch(lines[0])

	if len(matches) < 5 {
		return nil, errors.Newf("expected 6 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) != 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	destinationAccount := p.formatDestinationAccount(matches[4])

	finalTx := &Transaction{
		ID:                          uuid.NewString(),
		Date:                        date,
		DestinationCurrency:         matches[2],
		Description:                 matches[3],
		DestinationAmount:           amount,
		Type:                        TransactionTypeInternalTransfer,
		SourceAccount:               destinationAccount,
		DestinationAccount:          source[0],
		InternalTransferDirectionTo: false,
		DateFromMessage:             source[1],
		Raw:                         raw,
	}

	return finalTx, nil
}

func (p *Privat24) parseInternalTransferTo(
	_ context.Context,
	raw string,
	lines []string,
	date time.Time,
) (*Transaction, error) {
	matches := internalTransferToRegex.FindStringSubmatch(lines[0])

	if len(matches) != 6 && len(matches) != 5 {
		return nil, errors.Newf("expected 5-6 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) != 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	destRaw := matches[4]
	if destRaw == "" && len(matches) == 6 {
		destRaw = matches[5]
	}

	destinationAccount := p.formatDestinationAccount(destRaw)

	if !strings.Contains(destinationAccount, "*") {
		destinationAccount = unk
	}

	finalTx := &Transaction{
		ID:                          uuid.NewString(),
		Date:                        date,
		SourceCurrency:              matches[2],
		Description:                 matches[3],
		SourceAmount:                amount,
		Type:                        TransactionTypeInternalTransfer,
		SourceAccount:               source[0],
		DestinationAccount:          destinationAccount,
		InternalTransferDirectionTo: true,
		DateFromMessage:             source[1],
		Raw:                         raw,
	}

	return finalTx, nil
}

func (p *Privat24) formatDestinationAccount(destinationAccount string) string {
	if len(destinationAccount) != 6 {
		return destinationAccount
	}

	return fmt.Sprintf("%s*%s", string(destinationAccount[0]), destinationAccount[4:])
}

func (p *Privat24) ParseRemoteTransfer(
	_ context.Context,
	raw string,
	date time.Time,
) (*Transaction, error) {
	lines := toLines(raw)

	if len(lines) < remoteTransferLinesCount {
		return nil, errors.Newf("expected %d lines, got %d", remoteTransferLinesCount, len(lines))
	}

	matches := remoteTransferRegex.FindStringSubmatch(lines[0])
	if len(matches) != 4 {
		return nil, errors.Newf("expected 4 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) != 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	finalTx := &Transaction{
		ID:              uuid.NewString(),
		Date:            date,
		SourceCurrency:  matches[2],
		Description:     matches[3],
		SourceAmount:    amount,
		Type:            TransactionTypeRemoteTransfer,
		SourceAccount:   source[0],
		Raw:             raw,
		DateFromMessage: source[1],
	}

	return finalTx, nil
}

func (p *Privat24) ParseCreditPayment(
	_ context.Context,
	raw string,
	date time.Time,
) (*Transaction, error) {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")

	lines := strings.Split(raw, "\n")
	if len(lines) < creditPaymentLinesCount {
		return nil, errors.Newf("expected %d lines, got %d", creditPaymentLinesCount, len(lines))
	}

	matches := simpleExpenseRegex.FindStringSubmatch(lines[0])
	if len(matches) != 4 {
		return nil, errors.Newf("expected 4 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) < 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	finalTx := &Transaction{
		ID:              uuid.NewString(),
		Date:            date,
		SourceCurrency:  matches[2],
		Description:     matches[3],
		SourceAmount:    amount,
		Type:            TransactionTypeExpense,
		SourceAccount:   source[0],
		Raw:             raw,
		DateFromMessage: "", // for some reason here its not time.. wtf
	}

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "курс ") { // apply exchange rate logic
			sp := strings.Split(line, " ")

			if len(sp) != 3 {
				return nil, errors.Newf("expected 3 parts for курс, got %v", spew.Sdump(sp))
			}

			currencies := strings.Split(sp[2], "/")
			if len(currencies) != 2 {
				return nil, errors.Newf("expected 2 currencies, got %v", spew.Sdump(currencies))
			}

			rate, rateErr := decimal.NewFromString(sp[1])
			if rateErr != nil {
				return nil, errors.Join(rateErr, errors.Newf("failed to parse rate %s", sp[1]))
			}

			if currencies[1] == finalTx.SourceCurrency {
				finalTx.DestinationCurrency = finalTx.SourceCurrency
				finalTx.DestinationAmount = finalTx.SourceAmount

				finalTx.SourceCurrency = currencies[0]
				finalTx.SourceAmount = amount.Mul(rate)
			} else if currencies[0] == finalTx.SourceCurrency {
				finalTx.DestinationCurrency = finalTx.SourceCurrency
				finalTx.DestinationAmount = finalTx.SourceAmount

				finalTx.SourceCurrency = currencies[1]
				finalTx.SourceAmount = amount.Div(rate)
			} else {
				return nil, errors.Newf("currency mismatch: %s %s", currencies[0], currencies[1])
			}
		}
	}

	if finalTx.DestinationCurrency == "" && finalTx.DestinationAmount.IsZero() {
		finalTx.DestinationCurrency = finalTx.SourceCurrency
		finalTx.DestinationAmount = finalTx.SourceAmount.Abs()
	}

	for _, line := range lines {
		balMatch := balanceRegex.FindStringSubmatch(line)
		if len(balMatch) != 2 {
			continue
		}

		if balMatch[1] != finalTx.SourceCurrency {
			return nil, errors.Newf("currency mismatch: %s != %s", balMatch[1], finalTx.SourceCurrency)
		}
	}

	return finalTx, nil
}

func (p *Privat24) ParseSimpleExpense(
	_ context.Context,
	raw string,
	date time.Time,
) (*Transaction, error) {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")

	lines := strings.Split(raw, "\n")
	if len(lines) < simpleExpenseLinesCount {
		return nil, errors.Newf("expected %d lines, got %d", simpleExpenseLinesCount, len(lines))
	}

	matches := simpleExpenseRegex.FindStringSubmatch(lines[0])
	if len(matches) != 4 {
		return nil, errors.Newf("expected 4 matches, got %v", spew.Sdump(matches))
	}

	amount, err := decimal.NewFromString(matches[1])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := strings.Split(lines[1], " ")
	if len(source) < 2 {
		return nil, errors.Newf("expected 2 source parts, got %v", spew.Sdump(source))
	}

	finalTx := &Transaction{
		ID:              uuid.NewString(),
		Date:            date,
		SourceCurrency:  matches[2],
		Description:     matches[3],
		SourceAmount:    amount,
		Type:            TransactionTypeExpense,
		SourceAccount:   source[0],
		Raw:             raw,
		DateFromMessage: source[1],
	}

	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "курс ") { // apply exchange rate logic
			sp := strings.Split(line, " ")

			if len(sp) != 3 {
				return nil, errors.Newf("expected 3 parts for курс, got %v", spew.Sdump(sp))
			}

			currencies := strings.Split(sp[2], "/")
			if len(currencies) != 2 {
				return nil, errors.Newf("expected 2 currencies, got %v", spew.Sdump(currencies))
			}

			rate, rateErr := decimal.NewFromString(sp[1])
			if rateErr != nil {
				return nil, errors.Join(rateErr, errors.Newf("failed to parse rate %s", sp[1]))
			}

			if currencies[1] == finalTx.SourceCurrency {
				finalTx.DestinationCurrency = finalTx.SourceCurrency
				finalTx.DestinationAmount = finalTx.SourceAmount

				finalTx.SourceCurrency = currencies[0]
				finalTx.SourceAmount = amount.Mul(rate)
			} else if currencies[0] == finalTx.SourceCurrency {
				finalTx.DestinationCurrency = finalTx.SourceCurrency
				finalTx.DestinationAmount = finalTx.SourceAmount

				finalTx.SourceCurrency = currencies[1]
				finalTx.SourceAmount = amount.Div(rate)
			} else {
				return nil, errors.Newf("currency mismatch: %s %s", currencies[0], currencies[1])
			}
		}
	}

	if finalTx.DestinationCurrency == "" && finalTx.DestinationAmount.IsZero() {
		finalTx.DestinationCurrency = finalTx.SourceCurrency
		finalTx.DestinationAmount = finalTx.SourceAmount.Abs()
	}

	for _, line := range lines {
		balMatch := balanceRegex.FindStringSubmatch(line)
		if len(balMatch) != 2 {
			continue
		}

		if balMatch[1] != finalTx.SourceCurrency {
			return nil, errors.Newf("currency mismatch: %s != %s", balMatch[1], finalTx.SourceCurrency)
		}
	}

	return finalTx, nil
}

func (p *Privat24) ParseHeaderDate(header string) (time.Time, error) {
	startIdx := strings.Index(header, "[")
	endIdx := strings.Index(header, "]")

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return time.Time{}, errors.Newf("invalid header format: %s", header)
	}

	dateStr := strings.TrimSpace(header[startIdx+1 : endIdx])

	parsedTime, err := time.Parse("1/2/2006 3:04 PM", dateStr)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "failed to parse date: %s", dateStr)
	}

	return parsedTime, nil
}
