package mcp

import (
	"context"

	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/shopspring/decimal"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package mcp_test -source=interfaces.go

type CategoryService interface {
	CreateCategory(ctx context.Context, req *categoriesv1.CreateCategoryRequest) (*categoriesv1.CreateCategoryResponse, error)
	UpdateCategory(ctx context.Context, req *categoriesv1.UpdateCategoryRequest) (*categoriesv1.UpdateCategoryResponse, error)
}

type RulesService interface {
	ListRules(ctx context.Context, req *rulesv1.ListRulesRequest) (*rulesv1.ListRulesResponse, error)
	CreateRule(ctx context.Context, req *rulesv1.CreateRuleRequest) (*rulesv1.CreateRuleResponse, error)
	UpdateRule(ctx context.Context, req *rulesv1.UpdateRuleRequest) (*rulesv1.UpdateRuleResponse, error)
}

type DryRunService interface {
	DryRunRule(ctx context.Context, req *rulesv1.DryRunRuleRequest) (*rulesv1.DryRunRuleResponse, error)
}

type TagsService interface {
	ListTags(ctx context.Context, req *tagsv1.ListTagsRequest) (*tagsv1.ListTagsResponse, error)
	CreateTag(ctx context.Context, req *tagsv1.CreateTagRequest) (*tagsv1.CreateTagResponse, error)
	UpdateTag(ctx context.Context, req *tagsv1.UpdateTagRequest) (*tagsv1.UpdateTagResponse, error)
	DeleteTag(ctx context.Context, req *tagsv1.DeleteTagRequest) error
}

type TransactionService interface {
	Create(ctx context.Context, req *transactionsv1.CreateTransactionRequest) (*transactionsv1.CreateTransactionResponse, error)
	Update(ctx context.Context, req *transactionsv1.UpdateTransactionRequest) (*transactionsv1.UpdateTransactionResponse, error)
	BulkSetCategory(ctx context.Context, assignments []transactions.CategoryAssignment) error
	BulkSetTags(ctx context.Context, assignments []transactions.TagsAssignment) error
}

type CurrencyConverterService interface {
	Quote(ctx context.Context, from, to string, amount decimal.Decimal) (*currency.Quote, error)
}
