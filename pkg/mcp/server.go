package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gorm.io/gorm"
)

type Server struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	db         *gorm.DB
	cfg        *ServerConfig
}

type ServerConfig struct {
	DB             *gorm.DB
	Docs           string
	CategorySvc    CategoryService
	RulesSvc       RulesService
	DryRunSvc      DryRunService
	TagsSvc        TagsService
	TransactionSvc TransactionService
	CurrencySvc    CurrencyConverterService
}

func NewServer(cfg *ServerConfig) *Server {
	mcpServer := server.NewMCPServer(
		"go-money",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
		server.WithRecovery(),
	)

	s := &Server{
		mcpServer: mcpServer,
		db:        cfg.DB,
		cfg:       cfg,
	}

	s.registerTools()
	s.registerResources()

	s.httpServer = server.NewStreamableHTTPServer(mcpServer)

	return s
}

func (s *Server) registerTools() {
	queryTool := mcp.NewTool(
		"query",
		mcp.WithDescription(fmt.Sprintf("Run a read-only SQL query against the Go Money database. Schema: %v", s.cfg.Docs)),
		mcp.WithString(
			"sql",
			mcp.Description("The SQL SELECT query to execute"),
			mcp.Required(),
		),
	)
	s.mcpServer.AddTool(queryTool, s.handleQuery)

	bulkSetTransactionCategoryTool := mcp.NewTool(
		"bulk_set_transaction_category",
		mcp.WithDescription("Set or clear categories for multiple transactions in a single call"),
		mcp.WithArray(
			"assignments",
			mcp.Description("Array of objects with transaction_id (required) and category_id (optional, null to clear)"),
			mcp.Required(),
		),
	)
	s.mcpServer.AddTool(bulkSetTransactionCategoryTool, s.handleBulkSetTransactionCategory)

	createCategoryTool := mcp.NewTool(
		"create_category",
		mcp.WithDescription("Create a new category"),
		mcp.WithString(
			"name",
			mcp.Description("The name of the category"),
			mcp.Required(),
		),
	)
	s.mcpServer.AddTool(createCategoryTool, s.handleCreateCategory)

	updateCategoryTool := mcp.NewTool(
		"update_category",
		mcp.WithDescription("Update an existing category"),
		mcp.WithNumber(
			"id",
			mcp.Description("The ID of the category to update"),
			mcp.Required(),
		),
		mcp.WithString(
			"name",
			mcp.Description("The new name for the category"),
			mcp.Required(),
		),
	)
	s.mcpServer.AddTool(updateCategoryTool, s.handleUpdateCategory)

	listRulesTool := mcp.NewTool(
		"list_rules",
		mcp.WithDescription("List all transaction rules. Rules are Lua scripts that automatically modify transactions based on conditions."),
	)
	s.mcpServer.AddTool(listRulesTool, s.handleListRules)

	dryRunRuleTool := mcp.NewTool(
		"dry_run_rule",
		mcp.WithDescription("Test a rule against a transaction without persisting changes. Returns the transaction state before and after rule execution."),
		mcp.WithNumber(
			"transaction_id",
			mcp.Description("The ID of the transaction to test the rule against (use 0 for scheduled rules that create transactions)"),
			mcp.Required(),
		),
		mcp.WithString(
			"script",
			mcp.Description("The Lua script to test"),
			mcp.Required(),
		),
		mcp.WithString(
			"title",
			mcp.Description("Title/name for the rule being tested"),
		),
	)
	s.mcpServer.AddTool(dryRunRuleTool, s.handleDryRunRule)

	createRuleTool := mcp.NewTool(
		"create_rule",
		mcp.WithDescription("Create a new transaction rule. Rules are Lua scripts that automatically modify transactions based on conditions."),
		mcp.WithString(
			"title",
			mcp.Description("The title/name of the rule"),
			mcp.Required(),
		),
		mcp.WithString(
			"script",
			mcp.Description("The Lua script for the rule"),
			mcp.Required(),
		),
		mcp.WithNumber(
			"sort_order",
			mcp.Description("Order in which the rule is executed (lower numbers run first, default 0)"),
		),
		mcp.WithBoolean(
			"enabled",
			mcp.Description("Whether the rule is enabled (default true)"),
		),
		mcp.WithBoolean(
			"is_final_rule",
			mcp.Description("If true, stops rule execution in group when this rule matches (default false)"),
		),
		mcp.WithString(
			"group_name",
			mcp.Description("Group name for organizing rules (rules in same group are processed together)"),
		),
	)
	s.mcpServer.AddTool(createRuleTool, s.handleCreateRule)

	updateRuleTool := mcp.NewTool(
		"update_rule",
		mcp.WithDescription("Update an existing transaction rule"),
		mcp.WithNumber(
			"id",
			mcp.Description("The ID of the rule to update"),
			mcp.Required(),
		),
		mcp.WithString(
			"title",
			mcp.Description("The title/name of the rule"),
			mcp.Required(),
		),
		mcp.WithString(
			"script",
			mcp.Description("The Lua script for the rule"),
			mcp.Required(),
		),
		mcp.WithNumber(
			"sort_order",
			mcp.Description("Order in which the rule is executed (lower numbers run first)"),
		),
		mcp.WithBoolean(
			"enabled",
			mcp.Description("Whether the rule is enabled"),
		),
		mcp.WithBoolean(
			"is_final_rule",
			mcp.Description("If true, stops rule execution in group when this rule matches"),
		),
		mcp.WithString(
			"group_name",
			mcp.Description("Group name for organizing rules"),
		),
	)
	s.mcpServer.AddTool(updateRuleTool, s.handleUpdateRule)

	listTagsTool := mcp.NewTool(
		"list_tags",
		mcp.WithDescription("List all tags"),
	)
	s.mcpServer.AddTool(listTagsTool, s.handleListTags)

	createTagTool := mcp.NewTool(
		"create_tag",
		mcp.WithDescription("Create a new tag"),
		mcp.WithString(
			"name",
			mcp.Description("The name of the tag"),
			mcp.Required(),
		),
		mcp.WithString(
			"color",
			mcp.Description("The color of the tag"),
		),
		mcp.WithString(
			"icon",
			mcp.Description("The icon of the tag"),
		),
	)
	s.mcpServer.AddTool(createTagTool, s.handleCreateTag)

	updateTagTool := mcp.NewTool(
		"update_tag",
		mcp.WithDescription("Update an existing tag"),
		mcp.WithNumber(
			"id",
			mcp.Description("The ID of the tag to update"),
			mcp.Required(),
		),
		mcp.WithString(
			"name",
			mcp.Description("The new name for the tag"),
			mcp.Required(),
		),
		mcp.WithString(
			"color",
			mcp.Description("The color of the tag"),
		),
		mcp.WithString(
			"icon",
			mcp.Description("The icon of the tag"),
		),
	)
	s.mcpServer.AddTool(updateTagTool, s.handleUpdateTag)

	deleteTagTool := mcp.NewTool(
		"delete_tag",
		mcp.WithDescription("Delete a tag"),
		mcp.WithNumber(
			"id",
			mcp.Description("The ID of the tag to delete"),
			mcp.Required(),
		),
	)
	s.mcpServer.AddTool(deleteTagTool, s.handleDeleteTag)

	bulkSetTransactionTagsTool := mcp.NewTool(
		"bulk_set_transaction_tags",
		mcp.WithDescription("Set or replace tags for multiple transactions in a single call"),
		mcp.WithArray(
			"assignments",
			mcp.Description("Array of objects with transaction_id (required, number) and tag_ids (required, array of numbers — replaces all tags; empty array clears)"),
			mcp.Required(),
		),
	)
	s.mcpServer.AddTool(bulkSetTransactionTagsTool, s.handleBulkSetTransactionTags)

	createExpenseTool := mcp.NewTool(
		"create_expense",
		mcp.WithDescription("Create an expense transaction. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: fx_source_amount (negative), fx_source_currency, notes, extra (map), tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id."),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("source_account_id", mcp.Description("Source account id"), mcp.Required()),
		mcp.WithString("source_amount", mcp.Description("Decimal string; must be negative"), mcp.Required()),
		mcp.WithString("source_currency", mcp.Description("ISO-4217 source currency"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string; must be positive"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("fx_source_amount", mcp.Description("Optional foreign source amount, negative decimal string")),
		mcp.WithString("fx_source_currency", mcp.Description("Optional foreign source currency")),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(createExpenseTool, s.handleCreateExpense)

	createIncomeTool := mcp.NewTool(
		"create_income",
		mcp.WithDescription("Create an income transaction. Required: transaction_date (RFC3339), title, source_account_id, source_amount (decimal string), source_currency, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id."),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("source_account_id", mcp.Description("Source account id"), mcp.Required()),
		mcp.WithString("source_amount", mcp.Description("Decimal string"), mcp.Required()),
		mcp.WithString("source_currency", mcp.Description("ISO-4217 source currency"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(createIncomeTool, s.handleCreateIncome)

	createTransferTool := mcp.NewTool(
		"create_transfer",
		mcp.WithDescription("Create a transfer between accounts. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id."),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("source_account_id", mcp.Description("Source account id"), mcp.Required()),
		mcp.WithString("source_amount", mcp.Description("Decimal string; must be negative"), mcp.Required()),
		mcp.WithString("source_currency", mcp.Description("ISO-4217 source currency"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string; must be positive"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(createTransferTool, s.handleCreateTransfer)

	createAdjustmentTool := mcp.NewTool(
		"create_adjustment",
		mcp.WithDescription("Create a balance adjustment. The source account is resolved automatically (default adjustment account) and the source amount is derived via the currency converter. Required: transaction_date (RFC3339), title, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id."),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(createAdjustmentTool, s.handleCreateAdjustment)

	updateExpenseTool := mcp.NewTool(
		"update_expense",
		mcp.WithDescription("Create an expense transaction. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: fx_source_amount (negative), fx_source_currency, notes, extra (map), tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace)."),
		mcp.WithNumber("id", mcp.Description("Transaction id to replace"), mcp.Required()),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("source_account_id", mcp.Description("Source account id"), mcp.Required()),
		mcp.WithString("source_amount", mcp.Description("Decimal string; must be negative"), mcp.Required()),
		mcp.WithString("source_currency", mcp.Description("ISO-4217 source currency"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string; must be positive"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("fx_source_amount", mcp.Description("Optional foreign source amount, negative decimal string")),
		mcp.WithString("fx_source_currency", mcp.Description("Optional foreign source currency")),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(updateExpenseTool, s.handleUpdateExpense)

	updateIncomeTool := mcp.NewTool(
		"update_income",
		mcp.WithDescription("Create an income transaction. Required: transaction_date (RFC3339), title, source_account_id, source_amount (decimal string), source_currency, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace)."),
		mcp.WithNumber("id", mcp.Description("Transaction id to replace"), mcp.Required()),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("source_account_id", mcp.Description("Source account id"), mcp.Required()),
		mcp.WithString("source_amount", mcp.Description("Decimal string"), mcp.Required()),
		mcp.WithString("source_currency", mcp.Description("ISO-4217 source currency"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(updateIncomeTool, s.handleUpdateIncome)

	updateTransferTool := mcp.NewTool(
		"update_transfer",
		mcp.WithDescription("Create a transfer between accounts. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace)."),
		mcp.WithNumber("id", mcp.Description("Transaction id to replace"), mcp.Required()),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("source_account_id", mcp.Description("Source account id"), mcp.Required()),
		mcp.WithString("source_amount", mcp.Description("Decimal string; must be negative"), mcp.Required()),
		mcp.WithString("source_currency", mcp.Description("ISO-4217 source currency"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string; must be positive"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(updateTransferTool, s.handleUpdateTransfer)

	updateAdjustmentTool := mcp.NewTool(
		"update_adjustment",
		mcp.WithDescription("Create a balance adjustment. The source account is resolved automatically (default adjustment account) and the source amount is derived via the currency converter. Required: transaction_date (RFC3339), title, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace)."),
		mcp.WithNumber("id", mcp.Description("Transaction id to replace"), mcp.Required()),
		mcp.WithString("transaction_date", mcp.Description("RFC3339 date-time"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Short description"), mcp.Required()),
		mcp.WithNumber("destination_account_id", mcp.Description("Destination account id"), mcp.Required()),
		mcp.WithString("destination_amount", mcp.Description("Decimal string"), mcp.Required()),
		mcp.WithString("destination_currency", mcp.Description("ISO-4217 destination currency"), mcp.Required()),
		mcp.WithString("notes", mcp.Description("Free-text notes")),
		mcp.WithObject("extra", mcp.Description("String-to-string metadata map")),
		mcp.WithArray("tag_ids", mcp.Description("Tag ids to attach")),
		mcp.WithString("reference_number", mcp.Description("External reference number")),
		mcp.WithArray("internal_reference_numbers", mcp.Description("Internal reference numbers")),
		mcp.WithString("group_key", mcp.Description("Group key for linked transactions")),
		mcp.WithBoolean("skip_rules", mcp.Description("Skip rule engine if true")),
		mcp.WithNumber("category_id", mcp.Description("Category id")),
	)
	s.mcpServer.AddTool(updateAdjustmentTool, s.handleUpdateAdjustment)

	convertCurrencyTool := mcp.NewTool(
		"convert_currency",
		mcp.WithDescription("Convert an amount between two currencies using stored exchange rates. Rates are denominated vs base currency: amount / from_rate → base → × to_rate. Same-currency calls pass through (both rates = 1). Returns converted amount plus from_rate, to_rate, and base_currency so the caller can verify the math."),
		mcp.WithString("from", mcp.Description("ISO-4217 source currency code"), mcp.Required()),
		mcp.WithString("to", mcp.Description("ISO-4217 target currency code"), mcp.Required()),
		mcp.WithString("amount", mcp.Description("Decimal amount as string"), mcp.Required()),
	)
	s.mcpServer.AddTool(convertCurrencyTool, s.handleConvertCurrency)
}

func (s *Server) registerResources() {
	schemaResource := mcp.NewResource(
		"context://schema",
		"Database Schema",
		mcp.WithResourceDescription("Go Money database schema documentation"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.mcpServer.AddResource(schemaResource, s.handleSchemaResource)
}

func (s *Server) handleSchemaResource(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     s.cfg.Docs,
		},
	}, nil
}

func (s *Server) Handler() http.Handler {
	return s.httpServer
}

func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}
