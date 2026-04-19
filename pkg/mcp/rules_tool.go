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

// RulesLuaAPIDoc describes the Lua runtime surface available to transaction
// rule scripts. Injected into create_rule, update_rule, and dry_run_rule tool
// descriptions so agents have enough context to author working scripts without
// reading external docs.
const RulesLuaAPIDoc = `
Lua runtime: gopher-lua with gopher-lua-libs preloaded (string, table, math, etc.).
Each rule runs against a cloned transaction; a rule is considered "applied" when
it mutates any tx field. Scripts must not raise errors — use early ` + "`return`" + ` to exit.
Rule groups run in sort_order; within a group ` + "`is_final_rule=true`" + ` stops the group
when the rule mutates the transaction.

Globals:
  tx       — current transaction (methods below)
  helpers  — utility namespace

Transaction API (` + "`tx:field()`" + ` = get, ` + "`tx:field(value)`" + ` = set):

  String fields:
    tx:title()               / tx:title("value")
    tx:notes()               / tx:notes("value")
    tx:sourceCurrency()      / tx:sourceCurrency("USD")
    tx:destinationCurrency() / tx:destinationCurrency("USD")
    tx:referenceNumber()     / tx:referenceNumber("REF123")

  Int fields:
    tx:sourceAccountID()      / tx:sourceAccountID(42)
    tx:destinationAccountID() / tx:destinationAccountID(42)

  Nullable int fields (pass nil to clear / reset):
    tx:categoryID()      / tx:categoryID(10) / tx:categoryID(nil)
    tx:transactionType() / tx:transactionType(3)
      0=UNSPECIFIED 1=TRANSFER_BETWEEN_ACCOUNTS 2=INCOME
      3=EXPENSE     4=REVERSAL                  5=ADJUSTMENT

  Decimal amounts (number or nil):
    tx:sourceAmount()      / tx:sourceAmount(12.34)      / tx:sourceAmount(nil)
    tx:destinationAmount() / tx:destinationAmount(12.34) / tx:destinationAmount(nil)
    tx:getSourceAmountWithDecimalPlaces(2)       -- rounded to N decimals
    tx:getDestinationAmountWithDecimalPlaces(2)

  Tags (tag IDs are ints):
    tx:addTag(tagID)
    tx:removeTag(tagID)
    tx:getTags()        -- Lua array of tag IDs
    tx:removeAllTags()

  Internal reference numbers (string array):
    tx:getInternalReferenceNumbers()
    tx:addInternalReferenceNumber("value")
    tx:setInternalReferenceNumbers({"a","b"})
    tx:removeInternalReferenceNumber("value")

  Date/time:
    tx:transactionDateTimeSetTime(hour, minute)
    tx:transactionDateTimeAddDate(years, months, days)

Helpers API:
  helpers:getAccountByID(id)
    Returns account object with fields: ID, Name, Currency, CurrentBalance,
    Type, AccountNumber, Iban (and a few more).
  helpers:convertCurrency("FROM", "TO", amount)
    Returns converted number, rounded to target currency decimals.

Nil-safety: ` + "`tx:title()`" + ` / ` + "`tx:notes()`" + ` can be nil on sparse imports — use
` + "`tx:title() or \"\"`" + ` before string.find.

Literal substring match (disable Lua patterns with 4th arg ` + "`true`" + `):
  if string.find(tx:title(), "Trading 212", 1, true) then ... end

Example — categorize by title keyword list:
  local keywords = { "GOOGLE -ADS", "MERCHANT X" }
  for _, keyword in ipairs(keywords) do
      if string.find(tx:title(), keyword, 1, true) then
          tx:categoryID(10)
          break
      end
  end

Example — match title OR notes, two sources:
  local title_keywords = { "Alice S", "Alice" }
  for _, keyword in ipairs(title_keywords) do
      if string.find(tx:title(), keyword, 1, true) then
          tx:categoryID(36)
          return
      end
  end
  local notes = tx:notes() or ""
  if string.find(notes, "ALICE SURNAME", 1, true) or
     string.find(notes, "021600146217XXXXXXXXXXXXX", 1, true) then
      tx:categoryID(36)
  end

Example — reclassify as transfer to specific account:
  local title = tx:title() or ""
  local notes = tx:notes() or ""
  if string.find(title, "Trading 212", 1, true) or
     string.find(notes, "Trading 212", 1, true) then
      tx:transactionType(1)            -- TRANSFER_BETWEEN_ACCOUNTS
      tx:destinationAccountID(47)      -- target account id
  end
`

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
