package handlers

import (
	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	configurationv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/configuration/v1"
	currencyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/currency/v1"
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	usersv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/users/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package handlers_test -source=interfaces.go

type TransactionsSvc interface {
	Create(
		ctx context.Context,
		req *transactionsv1.CreateTransactionRequest,
	) (*transactionsv1.CreateTransactionResponse, error)

	List(
		ctx context.Context,
		req *transactionsv1.ListTransactionsRequest,
	) (*transactionsv1.ListTransactionsResponse, error)

	Update(
		ctx context.Context,
		msg *transactionsv1.UpdateTransactionRequest,
	) (*transactionsv1.UpdateTransactionResponse, error)
}

type UserSvc interface {
	Login(
		ctx context.Context,
		req *usersv1.LoginRequest,
	) (*usersv1.LoginResponse, error)

	Create(
		ctx context.Context,
		req *usersv1.CreateRequest,
	) (*usersv1.CreateResponse, error)
}

type ImportSvc interface {
	Import(ctx context.Context, req *importv1.ImportTransactionsRequest) (*importv1.ImportTransactionsResponse, error)
}

type AccountSvc interface {
	Create(
		ctx context.Context,
		req *accountsv1.CreateAccountRequest,
	) (*accountsv1.CreateAccountResponse, error)

	Update(
		ctx context.Context,
		req *accountsv1.UpdateAccountRequest,
	) (*accountsv1.UpdateAccountResponse, error)

	List(
		ctx context.Context,
		req *accountsv1.ListAccountsRequest,
	) (*accountsv1.ListAccountsResponse, error)

	Delete(
		ctx context.Context,
		req *accountsv1.DeleteAccountRequest,
	) (*accountsv1.DeleteAccountResponse, error)

	CreateBulk(
		ctx context.Context,
		req *accountsv1.CreateAccountsBulkRequest,
	) (*accountsv1.CreateAccountsBulkResponse, error)
}

type ConfigSvc interface {
	GetConfiguration(
		ctx context.Context,
		_ *configurationv1.GetConfigurationRequest,
	) (*configurationv1.GetConfigurationResponse, error)
}

type CurrencySvc interface {
	GetCurrencies(
		ctx context.Context,
		_ *currencyv1.GetCurrenciesRequest,
	) (*currencyv1.GetCurrenciesResponse, error)

	CreateCurrency(
		ctx context.Context,
		req *currencyv1.CreateCurrencyRequest,
	) (*currencyv1.CreateCurrencyResponse, error)

	UpdateCurrency(
		ctx context.Context,
		req *currencyv1.UpdateCurrencyRequest,
	) (*currencyv1.UpdateCurrencyResponse, error)

	DeleteCurrency(
		ctx context.Context,
		req *currencyv1.DeleteCurrencyRequest,
	) (*currencyv1.DeleteCurrencyResponse, error)
}

type DecimalSvc interface {
	ToString(ctx context.Context, amount decimal.Decimal, currency string) string
}

type ConverterSvc interface {
	Convert(
		ctx context.Context,
		fromCurrency string,
		toCurrency string,
		amount decimal.Decimal,
	) (decimal.Decimal, error)
}

type TagSvc interface {
	GetAllTags(ctx context.Context) ([]*database.Tag, error)
	CreateTag(ctx context.Context, req *tagsv1.CreateTagRequest) (*tagsv1.CreateTagResponse, error)
	DeleteTag(ctx context.Context, req *tagsv1.DeleteTagRequest) error
	UpdateTag(ctx context.Context, req *tagsv1.UpdateTagRequest) (*tagsv1.UpdateTagResponse, error)
	ImportTags(ctx context.Context, req *tagsv1.ImportTagsRequest) (*tagsv1.ImportTagsResponse, error)
	ListTags(ctx context.Context, msg *tagsv1.ListTagsRequest) (*tagsv1.ListTagsResponse, error)
}

type RulesSvc interface {
	DeleteRule(ctx context.Context, req *rulesv1.DeleteRuleRequest) (*rulesv1.DeleteRuleResponse, error)
	CreateRule(ctx context.Context, req *rulesv1.CreateRuleRequest) (*rulesv1.CreateRuleResponse, error)
	ListRules(ctx context.Context, req *rulesv1.ListRulesRequest) (*rulesv1.ListRulesResponse, error)
	UpdateRule(ctx context.Context, req *rulesv1.UpdateRuleRequest) (*rulesv1.UpdateRuleResponse, error)
}

type DryRunSvc interface {
	DryRunRule(ctx context.Context, req *rulesv1.DryRunRuleRequest) (*rulesv1.DryRunRuleResponse, error)
}
