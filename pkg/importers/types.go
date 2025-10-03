package importers

import (
	"time"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
)

type GetAccountRequest struct {
	InitialAmount   decimal.Decimal
	InitialCurrency string

	Accounts    map[string]*database.Account
	AccountName string

	TransactionType gomoneypbv1.TransactionType
}

type GetSecondaryAccountResponse struct {
	Account                 *database.Account
	AmountInAccountCurrency decimal.Decimal // always positive
}

type ImportRequest struct {
	Data            []string
	Accounts        []*database.Account
	Tags            map[string]*database.Tag
	Categories      map[string]*database.Category
	SkipRules       bool
	TreatDatesAsUtc bool
}

type ParseRequest struct {
	ImportRequest
}

type ParseResponse struct {
	CreateRequests []*transactionsv1.CreateTransactionRequest
}

type DeduplicationItem struct {
	CreateRequest            *transactionsv1.CreateTransactionRequest
	DuplicationTransactionID *int64
}

type Transaction struct {
	ID   string
	Type TransactionType

	SourceAmount        decimal.Decimal
	SourceCurrency      string
	DestinationAmount   decimal.Decimal
	DestinationCurrency string

	Date               time.Time
	Description        string
	SourceAccount      string
	DestinationAccount string
	DateFromMessage    string
	Raw                string

	InternalTransferDirectionTo bool
	DuplicateTransactions       []*Transaction

	OriginalMessage   *Message `json:"-"`
	DeduplicationKeys []string

	OriginalTxType      string
	OriginalNadawcaName string
	ParsingError        error `json:"-"`
}

type Record struct {
	Message *Message
	Data    []byte
}

type Message struct {
	CreatedAt time.Time // todo
}

type TransactionType int32

const (
	TransactionTypeUnknown          = TransactionType(0)
	TransactionTypeIncome           = TransactionType(1)
	TransactionTypeExpense          = TransactionType(2)
	TransactionTypeInternalTransfer = TransactionType(3)
	TransactionTypeRemoteTransfer   = TransactionType(4)
)
