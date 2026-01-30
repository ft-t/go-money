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
	DB          *gorm.DB
	Docs        string
	CategorySvc CategoryService
	RulesSvc    RulesService
	DryRunSvc   DryRunService
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
