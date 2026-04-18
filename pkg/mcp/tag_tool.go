package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleListTags(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.TagsSvc.ListTags(queryCtx, &tagsv1.ListTagsRequest{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list tags: %v", err)), nil
	}

	if len(resp.Tags) == 0 {
		return mcp.NewToolResultText("No tags found"), nil
	}

	result, err := json.MarshalIndent(resp.Tags, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format tags: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleCreateTag(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	color, _ := args["color"].(string)
	icon, _ := args["icon"].(string)

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.TagsSvc.CreateTag(queryCtx, &tagsv1.CreateTagRequest{
		Name:  name,
		Color: color,
		Icon:  icon,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create tag: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tag created with id %d", resp.Tag.Id)), nil
}

func (s *Server) handleUpdateTag(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	tagID, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("id parameter is required"), nil
	}

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("name parameter is required"), nil
	}

	color, _ := args["color"].(string)
	icon, _ := args["icon"].(string)

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	_, err := s.cfg.TagsSvc.UpdateTag(queryCtx, &tagsv1.UpdateTagRequest{
		Id:    int32(tagID),
		Name:  name,
		Color: color,
		Icon:  icon,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update tag: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tag %d updated", int32(tagID))), nil
}

func (s *Server) handleDeleteTag(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	tagID, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("id parameter is required"), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	err := s.cfg.TagsSvc.DeleteTag(queryCtx, &tagsv1.DeleteTagRequest{
		Id: int32(tagID),
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete tag: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tag %d deleted", int32(tagID))), nil
}

func (s *Server) handleBulkSetTransactionTags(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	assignmentsRaw, ok := args["assignments"].([]any)
	if !ok || len(assignmentsRaw) == 0 {
		return mcp.NewToolResultError("assignments parameter is required and must be a non-empty array"), nil
	}

	assignments := make([]transactions.TagsAssignment, 0, len(assignmentsRaw))
	for i, item := range assignmentsRaw {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("assignment[%d] must be an object", i)), nil
		}

		txID, ok := itemMap["transaction_id"].(float64)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("assignment[%d].transaction_id is required and must be a number", i)), nil
		}

		tagIDsRaw, ok := itemMap["tag_ids"].([]any)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("assignment[%d].tag_ids is required and must be an array", i)), nil
		}

		tagIDs := make([]int32, 0, len(tagIDsRaw))
		for j, raw := range tagIDsRaw {
			id, ok := raw.(float64)
			if !ok {
				return mcp.NewToolResultError(fmt.Sprintf("assignment[%d].tag_ids[%d] must be a number", i, j)), nil
			}
			tagIDs = append(tagIDs, int32(id))
		}

		assignments = append(assignments, transactions.TagsAssignment{
			TransactionID: int64(txID),
			TagIDs:        tagIDs,
		})
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	if err := s.cfg.TransactionSvc.BulkSetTags(queryCtx, assignments); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update transactions: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Updated %d transactions", len(assignments))), nil
}
