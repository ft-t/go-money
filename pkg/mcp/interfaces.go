package mcp

import (
	"context"

	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package mcp_test -source=interfaces.go

type CategoryService interface {
	CreateCategory(ctx context.Context, req *categoriesv1.CreateCategoryRequest) (*categoriesv1.CreateCategoryResponse, error)
	UpdateCategory(ctx context.Context, req *categoriesv1.UpdateCategoryRequest) (*categoriesv1.UpdateCategoryResponse, error)
}
