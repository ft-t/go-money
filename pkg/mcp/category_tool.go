package mcp

import (
	"context"
	"fmt"

	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleCreateCategory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.CategorySvc.CreateCategory(queryCtx, &categoriesv1.CreateCategoryRequest{
		Category: &gomoneypbv1.Category{
			Name: name,
		},
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create category: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Category created with id %d", resp.Category.Id)), nil
}

func (s *Server) handleUpdateCategory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	categoryID, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("id parameter is required"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.CategorySvc.UpdateCategory(queryCtx, &categoriesv1.UpdateCategoryRequest{
		Category: &gomoneypbv1.Category{
			Id:   int32(categoryID),
			Name: name,
		},
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update category: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Category %d updated to '%s'", resp.Category.Id, resp.Category.Name)), nil
}
