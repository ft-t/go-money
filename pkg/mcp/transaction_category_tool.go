package mcp

import (
	"context"
	"fmt"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleSetTransactionCategory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	transactionID, ok := args["transaction_id"].(float64)
	if !ok {
		return mcp.NewToolResultError("transaction_id parameter is required"), nil
	}

	var categoryID *int32
	if catID, exists := args["category_id"]; exists && catID != nil {
		catIDFloat, ok := catID.(float64)
		if !ok {
			return mcp.NewToolResultError("category_id must be a number or null"), nil
		}
		catIDInt := int32(catIDFloat)
		categoryID = &catIDInt
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var tx database.Transaction
	if err := s.db.WithContext(queryCtx).First(&tx, int64(transactionID)).Error; err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("transaction not found: %v", err)), nil
	}

	tx.CategoryID = categoryID

	if err := s.db.WithContext(queryCtx).Save(&tx).Error; err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update transaction: %v", err)), nil
	}

	if categoryID == nil {
		return mcp.NewToolResultText(fmt.Sprintf("Transaction %d category cleared", int64(transactionID))), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Transaction %d category set to %d", int64(transactionID), *categoryID)), nil
}
