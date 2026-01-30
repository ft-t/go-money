package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleBulkSetTransactionCategory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	assignmentsRaw, ok := args["assignments"].([]any)
	if !ok || len(assignmentsRaw) == 0 {
		return mcp.NewToolResultError("assignments parameter is required and must be a non-empty array"), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	for i, item := range assignmentsRaw {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("assignment[%d] must be an object", i)), nil
		}

		txID, ok := itemMap["transaction_id"].(float64)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("assignment[%d].transaction_id is required and must be a number", i)), nil
		}

		var categoryID *int32
		if catID, exists := itemMap["category_id"]; exists && catID != nil {
			catIDFloat, ok := catID.(float64)
			if !ok {
				return mcp.NewToolResultError(fmt.Sprintf("assignment[%d].category_id must be a number or null", i)), nil
			}
			catIDInt := int32(catIDFloat)
			categoryID = &catIDInt
		}

		if err := s.db.WithContext(queryCtx).
			Table("transactions").
			Where("id = ?", int64(txID)).
			Update("category_id", categoryID).Error; err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to update transaction %d: %v", int64(txID), err)), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("Updated %d transactions", len(assignmentsRaw))), nil
}
