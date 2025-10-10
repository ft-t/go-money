package importers

import (
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo/mutable"
	"github.com/shopspring/decimal"
	"github.com/tealeg/xlsx"
)

type ParibasDataExtractor interface {
	Extract(ctx context.Context, cells []*xlsx.Cell) (ParibasData, error)
}

type ParibasDataExtractorV1 struct{}

type ParibasData struct {
	Date                    time.Time
	DateFromMessage         string
	Currency                string
	TransactionCurrency     string
	Amount                  decimal.Decimal
	AmountString            string
	TransactionAmount       decimal.Decimal
	TransactionAmountString string
	Sender                  string
	Receiver                string
	Description             string
	Account                 string
	DestinationAccount      string
	TransactionType         string
	Raw                     string
	ExecutedAt              string
}

func (d ParibasDataExtractorV1) Extract(_ context.Context, cells []*xlsx.Cell) (ParibasData, error) {
	date, cellErr := cells[0].GetTime(false)
	if cellErr != nil {
		return ParibasData{}, errors.Join(cellErr, errors.Newf("can not parse date: %s", cells[0].String()))
	}

	amount := cells[3].String()
	amountParsed, amountErr := decimal.NewFromString(amount)
	if amountErr != nil {
		return ParibasData{}, errors.Join(amountErr, errors.Newf("can not parse amount: %s", amount))
	}

	currency := cells[4].String()
	senderOrReceiver := cells[5].String()
	description := cells[6].String()

	rawAccount := cells[7].String()

	accountArr := toLines(strings.ToLower(rawAccount))
	
	mutable.Reverse(accountArr)
	account := accountArr[0]

	transactionType := cells[8].String()

	kwotaStr := cells[9].String()
	kwotaParsed, kwotaErr := decimal.NewFromString(kwotaStr)
	if kwotaErr != nil {
		return ParibasData{}, errors.Join(kwotaErr, errors.Newf("can not parse kwota: %s", kwotaStr))
	}

	transactionCurrency := cells[10].String()

	if description == "" {
		description = transactionType
	}

	account = stripAccountPrefix(account)
	destinationAccount := stripAccountPrefix(toLines(senderOrReceiver)[0])

	executedAt := cells[1].String()
	return ParibasData{
		Date:                    date,
		DateFromMessage:         date.Format("15:04"),
		TransactionType:         transactionType,
		Raw:                     strings.Join([]string{description, senderOrReceiver, rawAccount, transactionType}, "\n"),
		Description:             description,
		TransactionCurrency:     transactionCurrency,
		Currency:                currency,
		Amount:                  amountParsed,
		AmountString:            amount,
		TransactionAmount:       kwotaParsed,
		TransactionAmountString: kwotaStr,
		ExecutedAt:              executedAt,
		Account:                 account,
		DestinationAccount:      destinationAccount,
	}, nil
}

type ParibasDataExtractorV2 struct{}

func (d ParibasDataExtractorV2) Extract(_ context.Context, cells []*xlsx.Cell) (ParibasData, error) {
	date, cellErr := cells[0].GetTime(false)
	if cellErr != nil {
		return ParibasData{}, errors.Join(cellErr, errors.Newf("can not parse date: %s", cells[0].String()))
	}

	amount := cells[3].String()
	amountParsed, amountErr := decimal.NewFromString(amount)
	if amountErr != nil {
		return ParibasData{}, errors.Join(amountErr, errors.Newf("can not parse amount: %s", amount))
	}

	currency := cells[4].String()
	sender := cells[5].String()
	receiver := cells[6].String()
	description := cells[7].String()

	rawAccount := cells[8].String()
	accountArr := toLines(strings.ToLower(rawAccount))

	mutable.Reverse(accountArr)
	account := accountArr[0]

	transactionType := cells[9].String()

	kwotaStr := cells[10].String()
	kwotaParsed, kwotaErr := decimal.NewFromString(kwotaStr)
	if kwotaErr != nil {
		return ParibasData{}, errors.Join(kwotaErr, errors.Newf("can not parse kwota: %s", kwotaStr))
	}

	transactionCurrency := cells[11].String()

	if description == "" {
		description = transactionType
	}

	account = stripAccountPrefix(account)

	var destinationAccountRaw string
	if strings.Contains(sender, account) {
		destinationAccountRaw = receiver
	} else {
		destinationAccountRaw = sender
	}
	destinationAccount := stripAccountPrefix(toLines(destinationAccountRaw)[0])

	executedAt := cells[1].String()
	return ParibasData{
		Date:                    date,
		DateFromMessage:         date.Format("15:04"),
		TransactionType:         transactionType,
		Raw:                     strings.Join([]string{description, destinationAccountRaw, rawAccount, transactionType}, "\n"),
		Description:             description,
		TransactionCurrency:     transactionCurrency,
		Currency:                currency,
		Amount:                  amountParsed,
		AmountString:            amount,
		TransactionAmount:       kwotaParsed,
		TransactionAmountString: kwotaStr,
		ExecutedAt:              executedAt,
		Account:                 account,
		DestinationAccount:      destinationAccount,
	}, nil
}
