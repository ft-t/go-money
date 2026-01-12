package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleListRules(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.RulesSvc.ListRules(queryCtx, &rulesv1.ListRulesRequest{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list rules: %v", err)), nil
	}

	if len(resp.Rules) == 0 {
		return mcp.NewToolResultText("No rules found"), nil
	}

	result, err := json.MarshalIndent(resp.Rules, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format rules: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleDryRunRule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	transactionID, ok := args["transaction_id"].(float64)
	if !ok {
		return mcp.NewToolResultError("transaction_id parameter is required"), nil
	}

	script, ok := args["script"].(string)
	if !ok || script == "" {
		return mcp.NewToolResultError("script parameter is required"), nil
	}

	title, _ := args["title"].(string)
	if title == "" {
		title = "Test Rule"
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.DryRunSvc.DryRunRule(queryCtx, &rulesv1.DryRunRuleRequest{
		TransactionId: int64(transactionID),
		Rule: &gomoneypbv1.Rule{
			Title:  title,
			Script: script,
		},
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("dry run failed: %v", err)), nil
	}

	output := struct {
		RuleApplied bool                     `json:"rule_applied"`
		Before      *gomoneypbv1.Transaction `json:"before"`
		After       *gomoneypbv1.Transaction `json:"after"`
	}{
		RuleApplied: resp.RuleApplied,
		Before:      resp.Before,
		After:       resp.After,
	}

	result, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleCreateRule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	script, ok := args["script"].(string)
	if !ok || script == "" {
		return mcp.NewToolResultError("script parameter is required"), nil
	}

	rule := &gomoneypbv1.Rule{
		Title:   title,
		Script:  script,
		Enabled: true,
	}

	if sortOrder, ok := args["sort_order"].(float64); ok {
		rule.SortOrder = int32(sortOrder)
	}

	if enabled, ok := args["enabled"].(bool); ok {
		rule.Enabled = enabled
	}

	if isFinalRule, ok := args["is_final_rule"].(bool); ok {
		rule.IsFinalRule = isFinalRule
	}

	if groupName, ok := args["group_name"].(string); ok {
		rule.GroupName = groupName
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.RulesSvc.CreateRule(queryCtx, &rulesv1.CreateRuleRequest{
		Rule: rule,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create rule: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Rule created with id %d", resp.Rule.Id)), nil
}

func (s *Server) handleUpdateRule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	ruleID, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("id parameter is required"), nil
	}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	script, ok := args["script"].(string)
	if !ok || script == "" {
		return mcp.NewToolResultError("script parameter is required"), nil
	}

	rule := &gomoneypbv1.Rule{
		Id:     int32(ruleID),
		Title:  title,
		Script: script,
	}

	if sortOrder, ok := args["sort_order"].(float64); ok {
		rule.SortOrder = int32(sortOrder)
	}

	if enabled, ok := args["enabled"].(bool); ok {
		rule.Enabled = enabled
	}

	if isFinalRule, ok := args["is_final_rule"].(bool); ok {
		rule.IsFinalRule = isFinalRule
	}

	if groupName, ok := args["group_name"].(string); ok {
		rule.GroupName = groupName
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.RulesSvc.UpdateRule(queryCtx, &rulesv1.UpdateRuleRequest{
		Rule: rule,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update rule: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Rule %d updated", resp.Rule.Id)), nil
}
