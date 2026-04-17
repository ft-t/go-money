package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/shopspring/decimal"
)

func (s *Server) handleConvertCurrency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	from, _ := args["from"].(string)
	to, _ := args["to"].(string)
	amountStr, _ := args["amount"].(string)
	if from == "" || to == "" || amountStr == "" {
		return mcp.NewToolResultError("from, to, and amount are required"), nil
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid amount: %v", err)), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	quote, err := s.cfg.CurrencySvc.Quote(queryCtx, from, to, amount)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to convert: %v", err)), nil
	}

	payload := map[string]string{
		"from":          quote.From,
		"to":            quote.To,
		"amount":        quote.Amount.String(),
		"converted":     quote.Converted.String(),
		"from_rate":     quote.FromRate.String(),
		"to_rate":       quote.ToRate.String(),
		"base_currency": quote.BaseCurrency,
	}

	result, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(result)), nil
}
