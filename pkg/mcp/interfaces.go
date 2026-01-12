package mcp

import (
	"context"

	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
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
